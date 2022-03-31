package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterFeeContract registers a contract to receive transaction fees
func (k Keeper) RegisterFeeContract(
	goCtx context.Context,
	msg *types.MsgRegisterFeeContract,
) (*types.MsgRegisterFeeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	contract := common.HexToAddress(msg.Contract)

	if k.IsFeeRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract %s is already registered", contract)
	}

	deployer, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	derivedContractAddr := common.BytesToAddress(deployer)

	for _, nonce := range msg.Nonces {
		derivedContractAddr = crypto.CreateAddress(derivedContractAddr, nonce)
	}

	if contract != derivedContractAddr {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrorInvalidSigner,
			"%s not contract deployer or wrong nonce", msg.FromAddress,
		)
	}

	// check that the contract is deployed, to avoid spam registrations
	// TODO

	k.SetFee(ctx, types.FeeContract{
		Contract:        msg.Contract,
		Owner:           msg.FromAddress,
		WithdrawAddress: msg.WithdrawAddress,
	})

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeRegisterFeeContract,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.FromAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
				sdk.NewAttribute(types.AttributeKeyWithdrawAddress, msg.WithdrawAddress),
			),
		},
	)

	return &types.MsgRegisterFeeContractResponse{}, nil
}

// CancelFeeContract deletes the fee for a contract
func (k Keeper) CancelFeeContract(
	goCtx context.Context,
	msg *types.MsgCancelFeeContract,
) (*types.MsgCancelFeeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	feeContract, ok := k.GetFee(ctx, common.HexToAddress(msg.Contract))
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract %s is not registered", msg.Contract)
	}

	if msg.FromAddress != feeContract.Owner {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the contract deployer", msg.FromAddress)
	}

	k.DeleteFee(ctx, feeContract)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeCancelFeeContract,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.FromAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			),
		},
	)

	return &types.MsgCancelFeeContractResponse{}, nil
}

// UpdateFeeContract updates the withdraw address for a contract
func (k Keeper) UpdateFeeContract(
	goCtx context.Context,
	msg *types.MsgUpdateFeeContract,
) (*types.MsgUpdateFeeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	feeContract, ok := k.GetFee(ctx, common.HexToAddress(msg.Contract))
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract %s is not registered", msg.Contract)
	}

	if msg.FromAddress != feeContract.Owner {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the contract deployer", msg.FromAddress)
	}

	feeContract.WithdrawAddress = msg.WithdrawAddress
	k.SetFee(ctx, feeContract)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateFeeContract,
				sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
				sdk.NewAttribute(sdk.AttributeKeySender, msg.FromAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawAddress, msg.WithdrawAddress),
			),
		},
	)

	return &types.MsgUpdateFeeContractResponse{}, nil
}
