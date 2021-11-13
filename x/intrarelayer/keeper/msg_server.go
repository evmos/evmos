package keeper

import (
	"context"
	"encoding/json"
	"fmt"

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
	var res *evmtypes.MsgEthereumTxResponse

	// Check ownership
	if pair.IsNativeCoin() {
		// Only mint if the module is the owner of the deployed contract
		res, err = k.CallEVM(ctx, erc20, contract, "mint", receiver, msg.Coin.Amount.BigInt())
		if err != nil {
			return nil, err
		}

	} else if pair.IsNativeERC20() {
		// Unescrow tokens from module account if the user is the owner of the erc20 contract
		res, err = k.CallEVM(ctx, erc20, contract, "transfer", receiver, msg.Coin.Amount.BigInt())
		if err != nil {
			return nil, err
		}

		//Only burn cosmos coins if the user is the owner of the erc20 contract
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
			return nil, err
		}
	} else {
		return nil, types.ErrUndefinedOwner
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

	// NOTE: coin fields already validated

	// check ownership
	if pair.IsNativeCoin() {
		return k.convertERC20NativeCoin(ctx, pair, msg)
	} else if pair.IsNativeERC20() {
		return k.convertERC20NativeToken(ctx, pair, msg)
	} else {
		return nil, types.ErrUndefinedOwner
	}
}

func (k Keeper) convertERC20NativeCoin(ctx sdk.Context, pair types.TokenPair, msg *types.MsgConvertERC20) (*types.MsgConvertERC20Response, error) {
	// NOTE: error checked during msg validation
	receiver, _ := sdk.AccAddressFromBech32(msg.Receiver)
	sender := common.HexToAddress(msg.Sender)

	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}
	erc20 := contracts.ERC20BurnableAndMintableContract.ABI
	contract := pair.GetERC20Contract()

	// Only burn if the module is the owner of the contract

	transferData, err := erc20.Pack("transfer", types.ModuleAddress, msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}
	// Escrow tokens to module account
	ret, err := k.ExecuteEVM(ctx, contract, sender, transferData)
	if err != nil {
		return nil, err
	}

	unpackedRet, err := erc20.Unpack("transfer", ret)
	if err != nil {
		return nil, err
	}

	// Check unpackedRet execution
	if len(unpackedRet) == 0 {
		return nil, fmt.Errorf("Failed to execute escrow tokens from user")
	}

	if !unpackedRet[0].(bool) {
		return nil, fmt.Errorf("Failed to execute escrow tokens from user")
	}

	// Burn escrowed tokens
	res, err := k.CallEVM(ctx, erc20, contract, "burn", msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// We send previously escrowed coins to the receiver.
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

func (k Keeper) convertERC20NativeToken(ctx sdk.Context, pair types.TokenPair, msg *types.MsgConvertERC20) (*types.MsgConvertERC20Response, error) {
	// NOTE: error checked during msg validation
	receiver, _ := sdk.AccAddressFromBech32(msg.Receiver)
	sender := common.HexToAddress(msg.Sender)

	// NOTE: coin fields already validated
	coins := sdk.Coins{sdk.Coin{Denom: pair.Denom, Amount: msg.Amount}}
	erc20 := contracts.ERC20BurnableAndMintableContract.ABI
	contract := pair.GetERC20Contract()

	transferData, err := erc20.Pack("transfer", types.ModuleAddress, msg.Amount.BigInt())
	if err != nil {
		return nil, err
	}
	// Escrow coins to module account
	ret, err := k.ExecuteEVM(ctx, contract, sender, transferData)
	if err != nil {
		return nil, err
	}

	unpackedRet, err := erc20.Unpack("transfer", ret)
	if err != nil {
		return nil, err
	}

	// Check unpackedRet execution
	if len(unpackedRet) == 0 {
		return nil, fmt.Errorf("Failed to execute transfer")
	}

	if !unpackedRet[0].(bool) {
		return nil, fmt.Errorf("Failed to execute transfer")
	}

	// Only mint if the module generated the cosmos coins
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}

	// We send recently minted coins to the receiver.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, coins); err != nil {
		return nil, err
	}

	// TODO NEW EVENTS

	// txLogAttrs := make([]sdk.Attribute, 0)
	// for _, log := range res.Logs {
	// 	value, err := json.Marshal(log)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	txLogAttrs = append(txLogAttrs, sdk.NewAttribute(evmtypes.AttributeKeyTxLog, string(value)))
	// }

	// ctx.EventManager().EmitEvents(
	// 	sdk.Events{
	// 		sdk.NewEvent(
	// 			types.EventTypeConvertCoin,
	// 			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
	// 			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
	// 			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
	// 			sdk.NewAttribute(types.AttributeKeyCosmosCoin, pair.Denom),
	// 			sdk.NewAttribute(types.AttributeKeyERC20Token, msg.ContractAddress),
	// 			sdk.NewAttribute(evmtypes.AttributeKeyTxHash, res.Hash),
	// 		),
	// 		sdk.NewEvent(
	// 			evmtypes.EventTypeTxLog,
	// 			txLogAttrs...,
	// 		),
	// 	},
	// )

	return &types.MsgConvertERC20Response{}, nil
}
