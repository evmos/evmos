package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/claim/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// TotalUnclaimed.
func (k Keeper) TotalUnclaimed(c context.Context, _ *types.QueryTotalUnclaimedRequest) (*types.QueryTotalUnclaimedResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	moduleAccBal := k.GetModuleAccountBalances(ctx)

	return &types.QueryTotalUnclaimedResponse{
		Coins: moduleAccBal,
	}, nil
}

// Params returns params of the claim module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// ClaimRecord returns claimable amount per user
func (k Keeper) ClaimRecord(
	goCtx context.Context,
	req *types.QueryClaimRecordRequest,
) (*types.QueryClaimRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "claim record for address '%s'", req.Address)
	}

	return &types.QueryClaimRecordResponse{
		ClaimRecord: claimRecord,
	}, nil
}

// ClaimableForAction return the total amount claimable by an address for
// an action.
func (k Keeper) ClaimableForAction(
	goCtx context.Context,
	req *types.QueryClaimableForActionRequest,
) (*types.QueryClaimableForActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Action == types.ActionUnspecified || req.Action > types.ActionIBCTransfer {
		return nil, status.Error(codes.InvalidArgument, types.ErrInvalidAction.Error())
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)

	amt := sdk.ZeroInt()

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if found {
		amt = k.GetClaimableAmountForAction(ctx, addr, claimRecord, req.Action, params)
	}

	return &types.QueryClaimableForActionResponse{
		Coins: sdk.Coins{{Denom: params.ClaimDenom, Amount: amt}},
	}, nil
}

// TotalClaimable returns the total amount claimable by a user.
func (k Keeper) TotalClaimable(
	goCtx context.Context,
	req *types.QueryTotalClaimableRequest,
) (*types.QueryTotalClaimableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)

	amt, err := k.GetUserTotalClaimable(ctx, addr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalClaimableResponse{
		Coins: sdk.Coins{{Denom: params.ClaimDenom, Amount: amt}},
	}, nil
}
