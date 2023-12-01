package cosmos_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/utils"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v15/x/feemarket/types"
)

type AnteTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory

	enableFeemarket bool
	enableLondonHF  bool
	evmParamsOption func(*evmtypes.Params)
}

const TestGasLimit uint64 = 100000

var chainID = utils.TestnetChainID + "-1"

func (suite *AnteTestSuite) StateDB() *statedb.StateDB {
	ctx := suite.network.GetContext()
	return statedb.New(ctx, suite.network.App.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash())))
}

func (suite *AnteTestSuite) SetupTest() {
	// Custom genesis for tests
	genesis := make(map[string]interface{})

	if suite.enableFeemarket {
		// setup feemarketGenesis params
		feemarketGenesis := feemarkettypes.DefaultGenesisState()
		feemarketGenesis.Params.EnableHeight = 1
		feemarketGenesis.Params.NoBaseFee = false
		genesis[feemarkettypes.ModuleName] = feemarketGenesis
	}

	evmGenesis := evmtypes.DefaultGenesisState()
	if !suite.enableLondonHF {
		maxInt := sdkmath.NewInt(math.MaxInt64)
		evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
		evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
		evmGenesis.Params.ChainConfig.GrayGlacierBlock = &maxInt
		evmGenesis.Params.ChainConfig.MergeNetsplitBlock = &maxInt
		evmGenesis.Params.ChainConfig.ShanghaiBlock = &maxInt
		evmGenesis.Params.ChainConfig.CancunBlock = &maxInt
		genesis[evmtypes.ModuleName] = evmGenesis
	}

	if suite.evmParamsOption != nil {
		suite.evmParamsOption(&evmGenesis.Params)
	}

	// setup corresponding delegation/s and rewards
	keys := keyring.New(2)

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
		network.WithCustomGenesis(genesis),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	// We're using TestMsg amino encoding in some tests, so register it here.
	tf.WithCustomInterfaces([]factory.CustomInterface{
		{
			Intfce: &testdata.TestMsg{},
			Name:   "testdata.TestMsg",
			Copts:  nil,
		},
	})

	suite.network = nw
	suite.factory = tf
	suite.handler = gh
	suite.keyring = keys
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{
		enableLondonHF: true,
	})
}
