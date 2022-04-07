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

// DevFeeInfos returns all registered contracts for fee distribution
func (k Keeper) DevFeeInfos(
	c context.Context,
	req *types.QueryDevFeeInfosRequest,
) (*types.QueryDevFeeInfosResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var feeInfos []types.DevFeeInfo
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(key, value []byte) error {
			feeInfo := k.BuildFeeInfo(
				ctx,
				common.BytesToAddress(key),
				sdk.AccAddress(value),
			)
			feeInfos = append(feeInfos, feeInfo)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryDevFeeInfosResponse{
		Fees:       feeInfos,
		Pagination: pageRes,
	}, nil
}

// DevFeeInfo returns a given registered contract
func (k Keeper) DevFeeInfo(
	c context.Context,
	req *types.QueryDevFeeInfoRequest,
) (*types.QueryDevFeeInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if strings.TrimSpace(req.ContractAddress) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"contract address is empty",
		)
	}

	// check if the contract is a hex address
	if err := ethermint.ValidateAddress(req.ContractAddress); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for contract %s, should be hex ('0x...')", req.ContractAddress,
		)
	}

	feeInfo, found := k.GetFeeInfo(ctx, common.HexToAddress(req.ContractAddress))
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"fees registered contract '%s'",
			req.ContractAddress,
		)
	}

	return &types.QueryDevFeeInfoResponse{Fee: feeInfo}, nil
}

// Params returns the fees module params
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// DevFeeInfosPerDeployer returns the fee information for all contracts that a
// deployer has registered
func (k Keeper) DevFeeInfosPerDeployer(
	c context.Context,
	req *types.QueryDevFeeInfosPerDeployerRequest,
) (*types.QueryDevFeeInfosPerDeployerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if strings.TrimSpace(req.DeployerAddress) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"deployer address is empty",
		)
	}

	deployer, err := sdk.AccAddressFromBech32(req.DeployerAddress)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for deployer %s, should be bech32 ('evmos...')", req.DeployerAddress,
		)
	}

	contractAddresses := k.GetFeesInverse(ctx, deployer)
	var feeInfos []types.DevFeeInfo

	for _, contractAddress := range contractAddresses {
		feeInfo, found := k.GetFeeInfo(ctx, contractAddress)
		if found {
			feeInfos = append(feeInfos, feeInfo)
		}
	}

	return &types.QueryDevFeeInfosPerDeployerResponse{Fees: feeInfos}, nil
}
