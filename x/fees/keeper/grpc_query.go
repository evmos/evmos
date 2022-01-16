package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"

	"github.com/tharsis/evmos/x/fees/types"
)

var _ types.QueryServer = Keeper{}

// WithdrawAddresses return registered pairs
func (k Keeper) WithdrawAddresses(c context.Context, req *types.QueryWithdrawAddressesRequest) (*types.QueryWithdrawAddressesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var withdrawAddresses []types.ContractWithdrawAddress
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContractOwner)

	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		contractWithdrawAddress := types.ContractWithdrawAddress{
			ContractAddress: common.BytesToAddress(key).Hex(),
			WithdrawAddress: common.BytesToAddress(value).Hex(),
		}
		withdrawAddresses = append(withdrawAddresses, contractWithdrawAddress)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryWithdrawAddressesResponse{
		WithdrawAddresses: withdrawAddresses,
		Pagination:        pageRes,
	}, nil
}

// WithdrawAddress returns a given registered token pair
func (k Keeper) WithdrawAddress(c context.Context, req *types.QueryWithdrawAddressRequest) (*types.QueryWithdrawAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if err := ethermint.ValidateAddress(req.ContractAddress); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			err.Error(),
		)
	}

	contract := common.HexToAddress(req.ContractAddress)

	withdrawAddr, found := k.GetContractWithdrawAddress(ctx, contract)
	if !found {
		return nil, status.Errorf(codes.NotFound, "withdraw address for contract '%s'", req.ContractAddress)
	}

	return &types.QueryWithdrawAddressResponse{
		WithdrawAddress: withdrawAddr.Hex(),
	}, nil
}

// Params return distribution module params
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
