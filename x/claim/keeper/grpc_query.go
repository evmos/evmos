package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/claim/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// TotalUnclaimed returns the total amount unclaimed from the airdrop.
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

// ClaimRecords returns initial claimable amount per user and the claims per action
func (k Keeper) ClaimRecords(
	goCtx context.Context,
	req *types.QueryClaimRecordsRequest,
) (*types.QueryClaimRecordsResponse, error) {
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

	params := k.GetParams(ctx)
	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

	claims := make([]types.Claim, len(actions))
	for i, action := range actions {
		claims[i] = types.Claim{
			Action:          action,
			Completed:       claimRecord.HasClaimedAction(action),
			ClaimableAmount: k.GetClaimableAmountForAction(ctx, addr, claimRecord, action, params),
		}
	}

	return &types.QueryClaimRecordsResponse{
		InitialClaimableAmount: claimRecord.InitialClaimableAmount,
		Claims:                 claims,
	}, nil
}
