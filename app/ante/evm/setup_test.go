package evm_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v16/app"
	ante "github.com/evmos/evmos/v16/app/ante"
	"github.com/evmos/evmos/v16/encoding"
	"github.com/evmos/evmos/v16/ethereum/eip712"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"
)

type AnteTestSuite struct {
	suite.Suite

	network   *network.UnitTestNetwork
	handler   grpc.Handler
	keyring   keyring.Keyring
	factory   factory.TxFactory
	clientCtx client.Context

	anteHandler              sdk.AnteHandler
	enableFeemarket          bool
	baseFee                  *sdkmath.Int
	enableLondonHF           bool
	evmParamsOption          func(*evmtypes.Params)
	useLegacyEIP712TypedData bool
}

const TestGasLimit uint64 = 100000

func (suite *AnteTestSuite) SetupTest() {
	keys := keyring.New(2)

	customGenesis := network.CustomGenesisState{}
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	if suite.enableFeemarket {
		feemarketGenesis.Params.EnableHeight = 1
		feemarketGenesis.Params.NoBaseFee = false
	} else {
		feemarketGenesis.Params.NoBaseFee = true
	}
	if suite.baseFee != nil {
		feemarketGenesis.Params.BaseFee = *suite.baseFee
	}
	customGenesis[feemarkettypes.ModuleName] = feemarketGenesis

	evmGenesis := evmtypes.DefaultGenesisState()
	if !suite.enableLondonHF {
		maxInt := sdkmath.NewInt(math.MaxInt64)
		evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
		evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
		evmGenesis.Params.ChainConfig.GrayGlacierBlock = &maxInt
		evmGenesis.Params.ChainConfig.MergeNetsplitBlock = &maxInt
		evmGenesis.Params.ChainConfig.ShanghaiBlock = &maxInt
		evmGenesis.Params.ChainConfig.CancunBlock = &maxInt
	}
	if suite.evmParamsOption != nil {
		suite.evmParamsOption(&evmGenesis.Params)
	}
	customGenesis[evmtypes.ModuleName] = evmGenesis

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	suite.network = nw
	suite.factory = tf
	suite.handler = gh
	suite.keyring = keys

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	eip712.SetEncodingConfig(encodingConfig)

	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	suite.Require().NotNil(suite.network.App.AppCodec())

	anteHandler := ante.NewAnteHandler(ante.HandlerOptions{
		Cdc:                suite.network.App.AppCodec(),
		AccountKeeper:      suite.network.App.AccountKeeper,
		BankKeeper:         suite.network.App.BankKeeper,
		DistributionKeeper: suite.network.App.DistrKeeper,
		EvmKeeper:          suite.network.App.EvmKeeper,
		FeegrantKeeper:     suite.network.App.FeeGrantKeeper,
		IBCKeeper:          suite.network.App.IBCKeeper,
		StakingKeeper:      suite.network.App.StakingKeeper,
		FeeMarketKeeper:    suite.network.App.FeeMarketKeeper,
		SignModeHandler:    encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:     ante.SigVerificationGasConsumer,
	})

	suite.anteHandler = anteHandler
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{
		enableLondonHF: true,
	})

	// Re-run the tests with EIP-712 Legacy encodings to ensure backwards compatibility.
	// LegacyEIP712Extension should not be run with current TypedData encodings, since they are not compatible.
	suite.Run(t, &AnteTestSuite{
		enableLondonHF:           true,
		useLegacyEIP712TypedData: true,
	})
}
