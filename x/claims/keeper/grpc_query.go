package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

var _ types.QueryServer = Keeper{}

// TotalUnclaimed returns the total amount unclaimed from the airdrop
func (k Keeper) TotalUnclaimed(
	c context.Context,
	_ *types.QueryTotalUnclaimedRequest,
) (*types.QueryTotalUnclaimedResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	moduleAccBal := k.GetModuleAccountBalances(ctx)

	return &types.QueryTotalUnclaimedResponse{
		Coins: moduleAccBal,
	}, nil
}

// Params returns the params of the claim module
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// ClaimsRecords returns all claims records
func (k Keeper) ClaimsRecords(
	c context.Context,
	req *types.QueryClaimsRecordsRequest,
) (*types.QueryClaimsRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)

	claimsRecords := []types.ClaimsRecordAddress{}

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(key, value []byte) error {
			var cr types.ClaimsRecord
			if err := k.cdc.Unmarshal(value, &cr); err != nil {
				return err
			}

			cra := types.ClaimsRecordAddress{
				Address:                sdk.AccAddress(key).String(),
				InitialClaimableAmount: cr.InitialClaimableAmount,
				ActionsCompleted:       cr.ActionsCompleted,
			}

			claimsRecords = append(claimsRecords, cra)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryClaimsRecordsResponse{
		Claims:     claimsRecords,
		Pagination: pageRes,
	}, nil
}

// ClaimsRecord returns the initial claimable amount per user and the claims per
// action. Claimable amount per action will be 0 if claiming is disabled, before
// the start time, or after end time.
func (k Keeper) ClaimsRecord(
	c context.Context,
	req *types.QueryClaimsRecordRequest,
) (*types.QueryClaimsRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	claimsRecord, found := k.GetClaimsRecord(ctx, addr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "claim record for address '%s'", req.Address)
	}

	params := k.GetParams(ctx)
	actions := []types.Action{types.ActionVote, types.ActionDelegate, types.ActionEVM, types.ActionIBCTransfer}

	claims := make([]types.Claim, len(actions))
	for i, action := range actions {
		claims[i] = types.Claim{
			Action:          action,
			Completed:       claimsRecord.HasClaimedAction(action),
			ClaimableAmount: k.GetClaimableAmountForAction(ctx, claimsRecord, action, params),
		}
	}

	return &types.QueryClaimsRecordResponse{
		InitialClaimableAmount: claimsRecord.InitialClaimableAmount,
		Claims:                 claims,
	}, nil
}
