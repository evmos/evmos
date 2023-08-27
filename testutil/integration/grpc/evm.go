package grpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
)

// GetEvmAccount returns the EVM account for the given address.
func (gqh *IntegrationGrpcHandler) GetEvmAccount(address common.Address) (*evmtypes.QueryAccountResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Account(context.Background(), &evmtypes.QueryAccountRequest{
		Address: address.String(),
	})
}

// EstimateGas returns the estimated gas for the given call args.
func (gqh *IntegrationGrpcHandler) EstimateGas(args []byte, GasCap uint64) (*evmtypes.EstimateGasResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.EstimateGas(context.Background(), &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: GasCap,
	})
}

// GetEvmBalance returns the EVM balance for the given address.
func (gqh *IntegrationGrpcHandler) GetEvmParams() (*evmtypes.QueryParamsResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Params(context.Background(), &evmtypes.QueryParamsRequest{})
}
