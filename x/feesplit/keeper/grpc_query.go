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
	ethermint "github.com/evmos/ethermint/types"

	"github.com/evmos/evmos/v6/x/feesplit/types"
)

var _ types.QueryServer = Keeper{}

// FeeSplits returns all FeeSplits that have been registered for fee distribution
func (k Keeper) FeeSplits(
	c context.Context,
	req *types.QueryFeeSplitsRequest,
) (*types.QueryFeeSplitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var feeSplits []types.FeeSplit
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeSplit)

	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var fee types.FeeSplit
		if err := k.cdc.Unmarshal(value, &fee); err != nil {
			return err
		}
		feeSplits = append(feeSplits, fee)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeeSplitsResponse{
		FeeSplits:  feeSplits,
		Pagination: pageRes,
	}, nil
}

// FeeSplit returns the FeeSplit that has been registered for fee distribution for a given
// contract
func (k Keeper) FeeSplit(
	c context.Context,
	req *types.QueryFeeSplitRequest,
) (*types.QueryFeeSplitResponse, error) {
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

	// check if the contract is a non-zero hex address
	if err := ethermint.ValidateNonZeroAddress(req.ContractAddress); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for contract %s, should be non-zero hex ('0x...')", req.ContractAddress,
		)
	}

	feeSplit, found := k.GetFeeSplit(ctx, common.HexToAddress(req.ContractAddress))
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"fees registered contract '%s'",
			req.ContractAddress,
		)
	}

	return &types.QueryFeeSplitResponse{FeeSplit: feeSplit}, nil
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

// DeployerFeeSplits returns all contracts that have been registered for fee
// distribution by a given deployer
func (k Keeper) DeployerFeeSplits( // nolint: dupl
	c context.Context,
	req *types.QueryDeployerFeeSplitsRequest,
) (*types.QueryDeployerFeeSplitsResponse, error) {
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

	var contracts []string
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixDeployer(deployer),
	)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, _ []byte) error {
		contracts = append(contracts, common.BytesToAddress(key).Hex())
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDeployerFeeSplitsResponse{
		ContractAddresses: contracts,
		Pagination:        pageRes,
	}, nil
}

// WithdrawerFeeSplits returns all fees for a given withdraw address
func (k Keeper) WithdrawerFeeSplits( // nolint: dupl
	c context.Context,
	req *types.QueryWithdrawerFeeSplitsRequest,
) (*types.QueryWithdrawerFeeSplitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if strings.TrimSpace(req.WithdrawerAddress) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"withdraw address is empty",
		)
	}

	deployer, err := sdk.AccAddressFromBech32(req.WithdrawerAddress)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for withdraw addr %s, should be bech32 ('evmos...')", req.WithdrawerAddress,
		)
	}

	var contracts []string
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixWithdrawer(deployer),
	)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, _ []byte) error {
		contracts = append(contracts, common.BytesToAddress(key).Hex())

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryWithdrawerFeeSplitsResponse{
		ContractAddresses: contracts,
		Pagination:        pageRes,
	}, nil
}
