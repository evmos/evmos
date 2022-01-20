package keeper_test

// import (
// 	gocontext "context"
// 	"testing"

// 	"github.com/cosmos/cosmos-sdk/baseapp"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/suite"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// 	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
// 	simapp "github.com/tharsis/evmos/app"

// 	"github.com/tharsis/evmos/x/inflation/types"
// )

// type MintTestSuite struct {
// 	suite.Suite

// 	app         *simapp.Evmos
// 	ctx         sdk.Context
// 	queryClient types.QueryClient
// }

// func (suite *MintTestSuite) SetupTest() {
// 	// setup feemarketGenesis params
// 	feemarketGenesis := feemarkettypes.DefaultGenesisState()
// 	feemarketGenesis.Params.EnableHeight = 1
// 	feemarketGenesis.Params.NoBaseFee = false
// 	feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)
// 	app := simapp.Setup(false, feemarketGenesis)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

// 	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
// 	types.RegisterQueryServer(queryHelper, app.InflationKeeper)
// 	queryClient := types.NewQueryClient(queryHelper)

// 	suite.app = app
// 	suite.ctx = ctx

// 	suite.queryClient = queryClient
// }

// func (suite *MintTestSuite) TestGRPCParams() {
// 	_, _, queryClient := suite.app, suite.ctx, suite.queryClient

// 	_, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
// 	suite.Require().NoError(err)

// 	_, err = queryClient.EpochProvisions(gocontext.Background(), &types.QueryEpochProvisionsRequest{})
// 	suite.Require().NoError(err)
// }

// func TestMintTestSuite(t *testing.T) {
// 	suite.Run(t, new(MintTestSuite))
// }
