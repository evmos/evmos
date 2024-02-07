package evm_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/types"
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
	ethSigner                types.Signer
	enableFeemarket          bool
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
	suite.ethSigner = types.LatestSignerForChainID(suite.network.App.EvmKeeper.ChainID())
}

// func (suite *AnteTestSuite) SetupTest2() {
// 	checkTx := false

// 	suite.app = app.EthSetup(checkTx, func(app *app.Evmos, genesis evmostypes.GenesisState) evmostypes.GenesisState {
// 		if suite.enableFeemarket {
// 			// setup feemarketGenesis params
// 			feemarketGenesis := feemarkettypes.DefaultGenesisState()
// 			feemarketGenesis.Params.EnableHeight = 1
// 			feemarketGenesis.Params.NoBaseFee = false
// 			// Verify feeMarket genesis
// 			err := feemarketGenesis.Validate()
// 			suite.Require().NoError(err)
// 			genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
// 		}
// 		evmGenesis := evmtypes.DefaultGenesisState()
// 		evmGenesis.Params.AllowUnprotectedTxs = false
// 		if !suite.enableLondonHF {
// 			maxInt := sdkmath.NewInt(math.MaxInt64)
// 			evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
// 			evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
// 			evmGenesis.Params.ChainConfig.GrayGlacierBlock = &maxInt
// 			evmGenesis.Params.ChainConfig.MergeNetsplitBlock = &maxInt
// 			evmGenesis.Params.ChainConfig.ShanghaiBlock = &maxInt
// 			evmGenesis.Params.ChainConfig.CancunBlock = &maxInt
// 		}
// 		if suite.evmParamsOption != nil {
// 			suite.evmParamsOption(&evmGenesis.Params)
// 		}
// 		genesis[evmtypes.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)
// 		return genesis
// 	})

// 	suite.ctx = suite.app.BaseApp.NewContextLegacy(checkTx, tmproto.Header{Height: 2, ChainID: utils.TestnetChainID + "-1", Time: time.Now().UTC()})
// 	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(evmtypes.DefaultEVMDenom, sdkmath.OneInt())))
// 	suite.ctx = suite.ctx.WithBlockGasMeter(storetypes.NewGasMeter(1000000000000000000))

// 	// set staking denomination to Evmos denom
// 	params, err := suite.app.StakingKeeper.GetParams(suite.ctx)
// 	suite.Require().NoError(err)
// 	params.BondDenom = utils.BaseDenom
// 	err = suite.app.StakingKeeper.SetParams(suite.ctx, params)
// 	suite.Require().NoError(err)

// 	infCtx := suite.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
// 	err = suite.app.AccountKeeper.Params.Set(infCtx, authtypes.DefaultParams())
// 	suite.Require().NoError(err)

// 	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
// 	// We're using TestMsg amino encoding in some tests, so register it here.
// 	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
// 	eip712.SetEncodingConfig(encodingConfig)

// 	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

// 	suite.Require().NotNil(suite.app.AppCodec())

// 	anteHandler := ante.NewAnteHandler(ante.HandlerOptions{
// 		Cdc:                suite.app.AppCodec(),
// 		AccountKeeper:      suite.app.AccountKeeper,
// 		BankKeeper:         suite.app.BankKeeper,
// 		DistributionKeeper: suite.app.DistrKeeper,
// 		EvmKeeper:          suite.app.EvmKeeper,
// 		FeegrantKeeper:     suite.app.FeeGrantKeeper,
// 		IBCKeeper:          suite.app.IBCKeeper,
// 		StakingKeeper:      suite.app.StakingKeeper,
// 		FeeMarketKeeper:    suite.app.FeeMarketKeeper,
// 		SignModeHandler:    encodingConfig.TxConfig.SignModeHandler(),
// 		SigGasConsumer:     ante.SigVerificationGasConsumer,
// 	})

// 	suite.anteHandler = anteHandler
// 	suite.ethSigner = types.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
// }

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

	suite.Run(t, &AnteTestSuite{
		enableLondonHF:           true,
		useLegacyEIP712TypedData: true,
	})
}
