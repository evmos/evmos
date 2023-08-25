package grpc

import (
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/testutil/integration/network"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v14/x/inflation/types"
	revtypes "github.com/evmos/evmos/v14/x/revenue/v1/types"
)

// GrpcQueryHelper is a helper struct to query the network's modules
// via gRPC. This is to simulate the behavior of a real user and avoid querying
// the modules directly.
type GrpcQueryHelper struct {
	network network.NetworkManager
}

// NewGrpcQueryHelper creates a new GrpcQueryHelper instance.
func NewGrpcQueryHelper(network network.NetworkManager) *GrpcQueryHelper {
	return &GrpcQueryHelper{
		network: network,
	}
}

// -------------- Config --------------

func (gqh *GrpcQueryHelper) GetChainID() string {
	return gqh.network.GetChainID()
}

// TODO - this can be changed by a query param to the EVM module
func (gqh *GrpcQueryHelper) GetDenom() string {
	return gqh.network.GetDenom()
}

// -------------- Clients --------------
// Clietns need to be generated on every request to get the latest state
// from the network's context.

func (gqh *GrpcQueryHelper) getQueryHelper() *baseapp.QueryServiceTestHelper {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	interfaceRegistry := encCfg.InterfaceRegistry
	return baseapp.NewQueryServerTestHelper(gqh.network.GetContext(), interfaceRegistry)
}

func (gqh *GrpcQueryHelper) getEvmClient() evmtypes.QueryClient {
	queryHelper := gqh.getQueryHelper()
	evmtypes.RegisterQueryServer(queryHelper, gqh.network.App.EvmKeeper)
	return evmtypes.NewQueryClient(queryHelper)
}

func (gqh *GrpcQueryHelper) getRevenueClient() revtypes.QueryClient {
	queryHelper := gqh.getQueryHelper()
	revtypes.RegisterQueryServer(queryHelper, gqh.network.App.RevenueKeeper)
	return revtypes.NewQueryClient(queryHelper)
}

func (gqh *GrpcQueryHelper) getBankClient() banktypes.QueryClient {
	queryHelper := gqh.getQueryHelper()
	banktypes.RegisterQueryServer(queryHelper, gqh.network.App.BankKeeper)
	return banktypes.NewQueryClient(queryHelper)
}

func (gqh *GrpcQueryHelper) getTxClient() txtypes.ServiceClient {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	txtypes.RegisterInterfaces(interfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(gqh.network.GetContext(), interfaceRegistry)
	return txtypes.NewServiceClient(queryHelper)
}

func (gqh *GrpcQueryHelper) getFeeMarketClient() feemarkettypes.QueryClient {
	queryHelper := gqh.getQueryHelper()
	feemarkettypes.RegisterQueryServer(queryHelper, gqh.network.App.FeeMarketKeeper)
	return feemarkettypes.NewQueryClient(queryHelper)
}

func (gqh *GrpcQueryHelper) GetInflationClient() infltypes.QueryClient {
	queryHelper := gqh.getQueryHelper()
	infltypes.RegisterQueryServer(queryHelper, gqh.network.App.InflationKeeper)
	return infltypes.NewQueryClient(queryHelper)
}

func (gqh *GrpcQueryHelper) GetAuthClient() authtypes.QueryClient {
	queryHelper := gqh.getQueryHelper()
	authtypes.RegisterQueryServer(queryHelper, gqh.network.App.AccountKeeper)
	return authtypes.NewQueryClient(queryHelper)
}

// -------------- EVM --------------

// -------------- FeeMarket --------------

// -------------- Revenue --------------

// -------------- Bank --------------

// -------------- Tx --------------

// BroadcastTxSync broadcasts the given txBytes to the network and returns the response.
func (gqh *GrpcQueryHelper) BroadcastTxSync(txBytes []byte) (abcitypes.ResponseDeliverTx, error) {
	//..clientCtx := client.Context{}.WithTxConfig(encodingConfig.TxConfig).WithCodec(encodingConfig.Codec)
	req := abcitypes.RequestDeliverTx{Tx: txBytes}
	//..New
	// TODO - this should be change to gRPC
	// client ctx must be configured first
	return gqh.network.App.BaseApp.DeliverTx(req), nil
}

// BroadcastTxSync broadcasts the given txBytes to the network and returns the response.
func (gqh *GrpcQueryHelper) Simulate(txBytes []byte) (*txtypes.SimulateResponse, error) {
	// TODO - this should be change to gRPC
	// client ctx must be configured first
	gas, result, err := gqh.network.App.BaseApp.Simulate(txBytes)
	if err != nil {
		return nil, err
	}

	return &txtypes.SimulateResponse{
		GasInfo: &gas,
		Result:  result,
	}, nil
}

// -------------- Account --------------
