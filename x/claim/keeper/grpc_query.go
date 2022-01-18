package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tharsis/evmos/x/claim/types"
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

// ClaimRecords returns all the the claimable records
func (k Keeper) ClaimRecords(
	goCtx context.Context,
	req *types.QueryClaimRecordsRequest,
) (*types.QueryClaimRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimRecords)

	claimRecords := []types.ClaimRecordAddress{}

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(key, value []byte) error {
			var cr types.ClaimRecord
			if err := k.cdc.Unmarshal(value, &cr); err != nil {
				return err
			}

			cra := types.ClaimRecordAddress{
				Address:                sdk.AccAddress(key[1:]).String(),
				InitialClaimableAmount: cr.InitialClaimableAmount,
				ActionsCompleted:       cr.ActionsCompleted,
			}

			claimRecords = append(claimRecords, cra)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryClaimRecordsResponse{
		Claims:     claimRecords,
		Pagination: pageRes,
	}, nil
}

// ClaimRecord returns initial claimable amount per user and the claims per action
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

	return &types.QueryClaimRecordResponse{
		InitialClaimableAmount: claimRecord.InitialClaimableAmount,
		Claims:                 claims,
	}, nil
}
