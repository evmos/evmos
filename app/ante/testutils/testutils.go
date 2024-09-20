// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package testutils

import (
	"math"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/evmos/evmos/v20/app"
	ante "github.com/evmos/evmos/v20/app/ante"
	evmante "github.com/evmos/evmos/v20/app/ante/evm"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/config"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

type AnteTestSuite struct {
	suite.Suite

	network   *network.UnitTestNetwork
	handler   grpc.Handler
	keyring   keyring.Keyring
	factory   factory.TxFactory
	clientCtx client.Context

	anteHandler     sdk.AnteHandler
	enableFeemarket bool
	baseFee         *sdkmath.Int
	enableLondonHF  bool
	evmParamsOption func(*evmtypes.Params)
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

	if suite.evmParamsOption != nil {
		suite.evmParamsOption(&evmGenesis.Params)
	}
	customGenesis[evmtypes.ModuleName] = evmGenesis

	// set block max gas to be less than maxUint64
	cp := app.DefaultConsensusParams
	cp.Block.MaxGas = 1000000000000000000
	customGenesis[consensustypes.ModuleName] = cp

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

	encodingConfig := nw.GetEncodingConfig()

	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	suite.Require().NotNil(suite.network.App.AppCodec())

	chainConfig := config.DefaultChainConfig(suite.network.GetChainID())
	if !suite.enableLondonHF {
		maxInt := sdkmath.NewInt(math.MaxInt64)
		chainConfig.LondonBlock = maxInt.BigInt()
		chainConfig.ArrowGlacierBlock = maxInt.BigInt()
		chainConfig.GrayGlacierBlock = maxInt.BigInt()
		chainConfig.MergeNetsplitBlock = maxInt.BigInt()
		chainConfig.ShanghaiBlock = maxInt.BigInt()
		chainConfig.CancunBlock = maxInt.BigInt()
	}
	err := config.NewEVMConfigurator().
		WithChainConfig(chainConfig).
		Configure()
	suite.Require().NoError(err)

	anteHandler := ante.NewAnteHandler(ante.HandlerOptions{
		Cdc:                    suite.network.App.AppCodec(),
		AccountKeeper:          suite.network.App.AccountKeeper,
		BankKeeper:             suite.network.App.BankKeeper,
		DistributionKeeper:     suite.network.App.DistrKeeper,
		EvmKeeper:              suite.network.App.EvmKeeper,
		FeegrantKeeper:         suite.network.App.FeeGrantKeeper,
		IBCKeeper:              suite.network.App.IBCKeeper,
		StakingKeeper:          suite.network.App.StakingKeeper,
		FeeMarketKeeper:        suite.network.App.FeeMarketKeeper,
		SignModeHandler:        encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:         ante.SigVerificationGasConsumer,
		ExtensionOptionChecker: types.HasDynamicFeeExtensionOption,
		TxFeeChecker:           evmante.NewDynamicFeeChecker(suite.network.App.EvmKeeper),
	})

	suite.anteHandler = anteHandler
}

func (suite *AnteTestSuite) WithFeemarketEnabled(enabled bool) {
	suite.enableFeemarket = enabled
}

func (suite *AnteTestSuite) WithLondonHardForkEnabled(enabled bool) {
	suite.enableLondonHF = enabled
}

func (suite *AnteTestSuite) WithBaseFee(baseFee *sdkmath.Int) {
	suite.baseFee = baseFee
}

func (suite *AnteTestSuite) WithEvmParamsOptions(evmParamsOpts func(*evmtypes.Params)) {
	suite.evmParamsOption = evmParamsOpts
}

func (suite *AnteTestSuite) ResetEvmParamsOptions() {
	suite.evmParamsOption = nil
}

func (suite *AnteTestSuite) GetKeyring() keyring.Keyring {
	return suite.keyring
}

func (suite *AnteTestSuite) GetTxFactory() factory.TxFactory {
	return suite.factory
}

func (suite *AnteTestSuite) GetNetwork() *network.UnitTestNetwork {
	return suite.network
}

func (suite *AnteTestSuite) GetClientCtx() client.Context {
	return suite.clientCtx
}

func (suite *AnteTestSuite) GetAnteHandler() sdk.AnteHandler {
	return suite.anteHandler
}
