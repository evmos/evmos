package grpc

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v14/testutil/integration/network"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
	revtypes "github.com/evmos/evmos/v14/x/revenue/v1/types"
)

// GrpcHandler is an interface that defines the methods that are used to query
// the network's modules via gRPC.
type GrpcHandler interface {
	// EVM methods
	GetEvmAccount(address common.Address) (*evmtypes.QueryAccountResponse, error)
	EstimateGas(args []byte, GasCap uint64) (*evmtypes.EstimateGasResponse, error)
	GetEvmParams() (*evmtypes.QueryParamsResponse, error)

	// Bank methods
	GetBalance(address sdktypes.AccAddress, denom string) (*banktypes.QueryBalanceResponse, error)

	// Account methods
	GetAccount(address string) (authtypes.AccountI, error)

	// FeeMarket methods
	GetBaseFee() (*feemarkettypes.QueryBaseFeeResponse, error)

	// Revenue methods
	GetRevenue(address common.Address) (*revtypes.QueryRevenueResponse, error)
	GetRevenueParams() (*revtypes.QueryParamsResponse, error)
}

var _ GrpcHandler = (*IntegrationGrpcHandler)(nil)

// IntegrationGrpcHandler is a helper struct to query the network's modules
// via gRPC. This is to simulate the behavior of a real user and avoid querying
// the modules directly.
type IntegrationGrpcHandler struct {
	network network.Network
}

// NewGrpcHandler creates a new IntegrationGrpcHandler instance.
func NewGrpcHandler(network network.Network) GrpcHandler {
	return &IntegrationGrpcHandler{
		network: network,
	}
}
