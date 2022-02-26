package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tharsis/evmos/x/vesting/types"
)

var _ types.QueryServer = Keeper{}

// Unvested returns the unvested amount of tokens for a vesting account
func (k Keeper) Unvested(
	goCtx context.Context,
	req *types.QueryUnvestedRequest,
) (*types.QueryUnvestedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	clawbackAccount, err := k.validateQuery(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	unvested := clawbackAccount.GetVestingCoins(ctx.BlockTime())
	return &types.QueryUnvestedResponse{
		Unvested: unvested,
	}, nil
}

// Vested returns the unvested amount of tokens for a vesting account
func (k Keeper) Vested(
	goCtx context.Context,
	req *types.QueryVestedRequest,
) (*types.QueryVestedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	clawbackAccount, err := k.validateQuery(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	vested := clawbackAccount.GetVestedOnly(ctx.BlockTime())
	return &types.QueryVestedResponse{
		Vested: vested,
	}, nil
}

// Locked returns the unvested amount of tokens for a vesting account
func (k Keeper) Locked(
	goCtx context.Context,
	req *types.QueryLockedRequest,
) (*types.QueryLockedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	clawbackAccount, err := k.validateQuery(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	locked := clawbackAccount.GetLockedOnly(ctx.BlockTime())
	return &types.QueryLockedResponse{
		Locked: locked,
	}, nil
}

func (k Keeper) validateQuery(
	ctx sdk.Context,
	address string,
) (*types.ClawbackVestingAccount, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get vesting account
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, status.Errorf(codes.NotFound,
			"account for address '%s'", address,
		)
	}

	// Check if clawback vesting account
	clawbackAccount, isClawback := acc.(*types.ClawbackVestingAccount)
	if !isClawback {
		return nil, status.Errorf(codes.FailedPrecondition,
			"account for address '%s'", address,
		)
	}

	return clawbackAccount, nil
}
