package keeper

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
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

	erc20 := contracts.ERC20BurnableAndMintableContract.ABI
	contract := pair.GetERC20Contract()

	res, err := k.CallEVM(ctx, erc20, contract, "mint", receiver, msg.Coin.Amount.BigInt())
	if err != nil {
		return nil, err
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

	erc20 := contracts.ERC20BurnableContract.ABI
	contract := pair.GetERC20Contract()

	res, err := k.CallEVM(ctx, erc20, contract, "burn", msg.Amount.BigInt())
	if err != nil {
		return nil, err
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
				sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyERC20Token, msg.ContractAddress),
				sdk.NewAttribute(evmtypes.AttributeKeyTxHash, res.Hash),
			),
			sdk.NewEvent(
				evmtypes.EventTypeTxLog,
				txLogAttrs...,
			),
		},
	)

	return &types.MsgConvertERC20Response{}, nil
}
