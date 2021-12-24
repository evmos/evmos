package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/claim/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) ModuleAccountBalance(c context.Context, _ *types.QueryModuleAccountBalanceRequest) (*types.QueryModuleAccountBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	moduleAccBal := sdk.NewCoins(k.GetModuleAccountBalance(ctx))

	return &types.QueryModuleAccountBalanceResponse{ModuleAccountBalance: moduleAccBal}, nil
}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Claimable returns claimable amount per user
func (k Keeper) ClaimRecord(
	goCtx context.Context,
	req *types.QueryClaimRecordRequest,
) (*types.QueryClaimRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	claimRecord, err := k.GetClaimRecord(ctx, addr)
	return &types.QueryClaimRecordResponse{ClaimRecord: claimRecord}, err
}

// Activities returns activities
func (k Keeper) ClaimableForAction(
	goCtx context.Context,
	req *types.QueryClaimableForActionRequest,
) (*types.QueryClaimableForActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	coins, err := k.GetClaimableAmountForAction(ctx, addr, req.Action)

	return &types.QueryClaimableForActionResponse{
		Coins: coins,
	}, err
}

// Activities returns activities
func (k Keeper) TotalClaimable(
	goCtx context.Context,
	req *types.QueryTotalClaimableRequest,
) (*types.QueryTotalClaimableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	coins, err := k.GetUserTotalClaimable(ctx, addr)

	return &types.QueryTotalClaimableResponse{
		Coins: coins,
	}, err
}
