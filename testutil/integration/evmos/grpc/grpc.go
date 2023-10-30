// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"github.com/ethereum/go-ethereum/common"
	commongrpc "github.com/evmos/evmos/v15/testutil/integration/common/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v15/x/feemarket/types"
	revtypes "github.com/evmos/evmos/v15/x/revenue/v1/types"
)

// Handler is an interface that defines the methods that are used to query
// the network's modules via gRPC.
type Handler interface {
	commongrpc.Handler

	// EVM methods
	GetEvmAccount(address common.Address) (*evmtypes.QueryAccountResponse, error)
	EstimateGas(args []byte, GasCap uint64) (*evmtypes.EstimateGasResponse, error)
	GetEvmParams() (*evmtypes.QueryParamsResponse, error)

	// FeeMarket methods
	GetBaseFee() (*feemarkettypes.QueryBaseFeeResponse, error)

	// Revenue methods
	GetRevenue(address common.Address) (*revtypes.QueryRevenueResponse, error)
	GetRevenueParams() (*revtypes.QueryParamsResponse, error)
}

var _ Handler = (*IntegrationHandler)(nil)

// IntegrationHandler is a helper struct to query the network's modules
// via gRPC. This is to simulate the behavior of a real user and avoid querying
// the modules directly.
type IntegrationHandler struct {
	// We take the IntegrationHandler from common/grpc to get the common methods.
	*commongrpc.IntegrationHandler
	network network.Network
}

// NewIntegrationHandler creates a new IntegrationHandler instance.
func NewIntegrationHandler(network network.Network) Handler {
	return &IntegrationHandler{
		// Is there a better way to do this?
		IntegrationHandler: commongrpc.NewIntegrationHandler(network),
		network:            network,
	}
}
