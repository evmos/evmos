package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

var _ types.QueryServer = Keeper{}

// Incentives return registered incentives
func (k Keeper) FeesContracts(
	c context.Context,
	req *types.QueryFeesContractsRequest,
) (*types.QueryFeesContractsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var incentives []types.FeeContract
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(_, value []byte) error {
			var incentive types.FeeContract
			if err := k.cdc.Unmarshal(value, &incentive); err != nil {
				return err
			}
			incentives = append(incentives, incentive)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeesContractsResponse{
		Fees:       incentives,
		Pagination: pageRes,
	}, nil
}

// Incentive returns a given registered incentive
func (k Keeper) FeesContract(
	c context.Context,
	req *types.QueryFeesContractRequest,
) (*types.QueryFeesContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if strings.TrimSpace(req.Contract) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"contract address is empty",
		)
	}

	// check if the contract is a hex address
	if err := ethermint.ValidateAddress(req.Contract); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for contract %s, should be hex ('0x...')", req.Contract,
		)
	}

	incentive, found := k.GetFee(ctx, common.HexToAddress(req.Contract))
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"incentive with contract '%s'",
			req.Contract,
		)
	}

	return &types.QueryFeesContractResponse{Fees: incentive}, nil
}

// Params return hub contract param
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
