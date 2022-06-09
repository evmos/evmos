package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/tharsis/evmos/v5/x/vesting/types"
)

var _ types.QueryServer = Keeper{}

// Balances returns the locked, unvested and vested amount of tokens for a
// clawback vesting account
func (k Keeper) Balances(
	goCtx context.Context,
	req *types.QueryBalancesRequest,
) (*types.QueryBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get vesting account
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, status.Errorf(
			codes.NotFound,
			"account for address '%s'", req.Address,
		)
	}

	var locked, unvested, vested sdk.Coins
	blockTime := ctx.BlockTime()

	switch vestingAcc := acc.(type) {
	case *types.ClawbackVestingAccount:
		locked = vestingAcc.GetLockedOnly(blockTime)
		unvested = vestingAcc.GetUnvestedOnly(blockTime)
		vested = vestingAcc.GetVestedOnly(blockTime)
	case vestexported.VestingAccount:
		locked = vestingAcc.LockedCoins(blockTime)
		vested = vestingAcc.GetVestedCoins(blockTime)
		unvested = vestingAcc.GetVestingCoins(blockTime)
	default:
		return nil, status.Errorf(
			codes.InvalidArgument,
			"account at address '%s' is not a vesting account ", req.Address,
		)
	}

	return &types.QueryBalancesResponse{
		Locked:   locked,
		Unvested: unvested,
		Vested:   vested,
	}, nil
}
