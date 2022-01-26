package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tharsis/ethermint/tests"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *app.Evmos
	ctx         sdk.Context
	queryClient types.QueryClient
}

// Test helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false
	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)

	// init app
	suite.app = app.Setup(checkTx, feemarketGenesis)

	// setup inflation params
	inflationGenesis := inflationtypes.DefaultGenesisState()
	teamAddress := sdk.AccAddress(tests.GenerateAddress().Bytes())
	inflationGenesis.Params.TeamAddress = teamAddress.String()

	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.EpochsKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
