package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterIncentive creates an incentive for a contract
func (k Keeper) RegisterFeeContract(
	goCtx context.Context,
	msg *types.MsgRegisterFeeContract,
) (*types.MsgRegisterFeeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	contract := common.HexToAddress(msg.Contract)
	withdrawAddr, _ := sdk.AccAddressFromBech32(msg.WithdrawAddress)

	// TODO check owner is real owner (origin of deployment tx or ERC)
	fmt.Println("--RegisterContract", from, contract, withdrawAddr)

	if k.IsFeeRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract %s is already registered", contract)
	}

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

	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	contract, _ := sdk.AccAddressFromBech32(msg.Contract)

	// TODO check ownership, remove from keyvalue store
	fmt.Println("--CancelFeeContract", from, contract)

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

	from, _ := sdk.AccAddressFromBech32(msg.FromAddress)
	contract, _ := sdk.AccAddressFromBech32(msg.Contract)
	withdrawAddr, _ := sdk.AccAddressFromBech32(msg.WithdrawAddress)

	// TODO check ownership, update keyvalue store
	fmt.Println("--UpdateFeeContract", from, contract, withdrawAddr)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateFeeContract,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.FromAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			),
		},
	)

	return &types.MsgUpdateFeeContractResponse{}, nil
}
