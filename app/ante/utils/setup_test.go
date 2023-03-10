package utils_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/app/ante"
	"github.com/evmos/evmos/v12/encoding"
	"github.com/evmos/evmos/v12/ethereum/eip712"
	"github.com/evmos/evmos/v12/testutil"
	"github.com/evmos/evmos/v12/utils"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v12/x/feemarket/types"
)

type AnteTestSuite struct {
	suite.Suite

	ctx             sdk.Context
	app             *app.Evmos
	clientCtx       client.Context
	anteHandler     sdk.AnteHandler
	ethSigner       types.Signer
	enableFeemarket bool
	enableLondonHF  bool
	evmParamsOption func(*evmtypes.Params)
}

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.EthSetup(checkTx, func(app *app.Evmos, genesis simapp.GenesisState) simapp.GenesisState {
		if suite.enableFeemarket {
			// setup feemarketGenesis params
			feemarketGenesis := feemarkettypes.DefaultGenesisState()
			feemarketGenesis.Params.EnableHeight = 1
			feemarketGenesis.Params.NoBaseFee = false
			// Verify feeMarket genesis
			err := feemarketGenesis.Validate()
			suite.Require().NoError(err)
			genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}
		evmGenesis := evmtypes.DefaultGenesisState()
		evmGenesis.Params.AllowUnprotectedTxs = false
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
		genesis[evmtypes.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)
		return genesis
	})

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, ChainID: utils.TestnetChainID + "-1", Time: time.Now().UTC()})
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(evmtypes.DefaultEVMDenom, sdk.OneInt())))
	suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000000000000000))
	suite.app.EvmKeeper.WithChainID(suite.ctx)

	// set staking denomination to Evmos denom
	params := suite.app.StakingKeeper.GetParams(suite.ctx)
	params.BondDenom = utils.BaseDenom
	suite.app.StakingKeeper.SetParams(suite.ctx, params)

	infCtx := suite.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	suite.app.AccountKeeper.SetParams(infCtx, authtypes.DefaultParams())

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	// We're using TestMsg amino encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	eip712.SetEncodingConfig(encodingConfig)

	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	suite.Require().NotNil(suite.app.AppCodec())

	anteHandler := ante.NewAnteHandler(ante.HandlerOptions{
		Cdc:                suite.app.AppCodec(),
		AccountKeeper:      suite.app.AccountKeeper,
		BankKeeper:         suite.app.BankKeeper,
		DistributionKeeper: suite.app.DistrKeeper,
		EvmKeeper:          suite.app.EvmKeeper,
		FeegrantKeeper:     suite.app.FeeGrantKeeper,
		IBCKeeper:          suite.app.IBCKeeper,
		StakingKeeper:      suite.app.StakingKeeper,
		FeeMarketKeeper:    suite.app.FeeMarketKeeper,
		SignModeHandler:    encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:     ante.SigVerificationGasConsumer,
	})

	suite.anteHandler = anteHandler
	suite.ethSigner = types.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	var err error
	suite.ctx, err = testutil.Commit(suite.ctx, suite.app, time.Second*0, nil)
	suite.Require().NoError(err)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{
		enableLondonHF: true,
	})
}
