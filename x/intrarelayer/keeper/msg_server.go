package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/server/config"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	// "github.com/tharsis/evmos/solidity/contracts"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) ConvertCoin(goCtx context.Context, msg *types.MsgConvertCoin) (*types.MsgConvertCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.evmKeeper.WithContext(ctx)

	// NOTE: error checked during msg validation
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)
	receiver := common.HexToAddress(msg.Receiver)

	// NOTE: ignore validation from NewCoin constructor
	coins := sdk.Coins{msg.Coin}

	pair, err := k.MintingEnabled(ctx, sender, receiver.Bytes(), msg.Coin.Denom)
	if err != nil {
		return nil, err
	}

	// escrow coins
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to escrow coins")
	}

	// TODO: use init to compile ABI
	// erc20, err := abi.JSON(strings.NewReader(contracts.ContractsABI))
	// if err != nil {
	// 	return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to create ABI for erc20: %s", err.Error())
	// }

	var erc20 abi.ABI

	contract := pair.GetERC20Contract()

	// pack and mint ERC20 token
	payload, err := erc20.Pack(
		"mint",
		contract,
		common.BytesToAddress(sender),
		receiver,
		msg.Coin.Amount.BigInt(),
	)
	if err != nil {
		return nil, sdkerrors.Wrap(
			types.ErrWritingEthTxPayload,
			errors.Wrap(err, "failed to create transaction payload").Error(),
		)
	}

	nonce := k.evmKeeper.GetNonce(types.ModuleAddress)

	tx := ethtypes.NewMessage(
		types.ModuleAddress,
		&contract,
		nonce,
		big.NewInt(0),        // amount
		config.DefaultGasCap, // gasLimit
		big.NewInt(0),        // gasPrice
		payload,
		ethtypes.AccessList{}, // TODO: add AccessList?
		false,                 // checkNonce
	)

	res, err := k.evmKeeper.ApplyNativeMessage(tx)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, fmt.Errorf("contract call failed: method 'mint' %s, %s", pair.Erc20Address, res.VmError)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	txLogAttrs := make([]sdk.Attribute, 0)
	for _, log := range res.Logs {
		value, err := json.Marshal(log)
		if err != nil {
			return nil, err
		}
		txLogAttrs = append(txLogAttrs, sdk.NewAttribute(evmtypes.AttributeKeyTxLog, string(value)))
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeConvertCoin,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
				sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, msg.Coin.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, pair.Erc20Address),
				sdk.NewAttribute(evmtypes.AttributeKeyTxHash, res.Hash),
			),
			sdk.NewEvent(
				evmtypes.EventTypeTxLog,
				txLogAttrs...,
			),
		},
	)

	return &types.MsgConvertCoinResponse{}, nil
}

func (k Keeper) ConvertERC20(goCtx context.Context, msg *types.MsgConvertERC20) (*types.MsgConvertERC20Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.evmKeeper.WithContext(ctx)

	// NOTE: error checked during msg validation
	receiver, _ := sdk.AccAddressFromBech32(msg.Receiver)
	sender := common.HexToAddress(msg.Sender)

	pair, err := k.MintingEnabled(ctx, sender.Bytes(), receiver, msg.ContractAddress)
	if err != nil {
		return nil, err
	}

	contract := pair.GetERC20Contract()

	// TODO: use init to compile ABI
	// erc20, err := abi.JSON(strings.NewReader(contracts.ContractsABI))
	// if err != nil {
	// 	return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to create ABI for erc20: %s", err.Error())
	// }

	var erc20 abi.ABI

	// pack and burn ERC20
	payload, err := erc20.Pack(
		"burn",
		contract,
		sender,
		common.BytesToAddress(receiver),
		msg.Amount.BigInt(),
	)
	if err != nil {
		return nil, sdkerrors.Wrap(
			types.ErrWritingEthTxPayload,
			errors.Wrap(err, "failed to create transaction payload").Error(),
		)
	}

	nonce := k.evmKeeper.GetNonce(types.ModuleAddress)

	tx := ethtypes.NewMessage(
		types.ModuleAddress,
		&contract,
		nonce,
		big.NewInt(0),        // amount
		config.DefaultGasCap, // gasLimit
		big.NewInt(0),        // gasPrice
		payload,
		ethtypes.AccessList{}, // TODO: add AccessList?
		true,                  // checkNonce
	)

	res, err := k.evmKeeper.ApplyNativeMessage(tx)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, fmt.Errorf("contract call failed: method 'burn' %s, %s", pair.Erc20Address, res.VmError)
	}

	// TODO: check mint event on res Logs

	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}

	// mint and send tokens to recipient
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins); err != nil {
		return nil, err
	}

	// TODO: emit events
	// tx hash
	// pair denom
	// pair erc20
	// conversion amount
	// sender address
	// receiver address

	return &types.MsgConvertERC20Response{}, nil
}
