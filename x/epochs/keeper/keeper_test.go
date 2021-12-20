package keeper_test

// import (
// 	"testing"

// 	"github.com/cosmos/cosmos-sdk/baseapp"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/osmosis-labs/osmosis/app"
// 	"github.com/osmosis-labs/osmosis/x/epochs/types"
// 	"github.com/stretchr/testify/suite"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
// )

// type KeeperTestSuite struct {
// 	suite.Suite

// 	app         *app.OsmosisApp
// 	ctx         sdk.Context
// 	queryClient types.QueryClient
// }

// func (suite *KeeperTestSuite) SetupTest() {
// 	suite.app = app.Setup(false)
// 	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

// 	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
// 	types.RegisterQueryServer(queryHelper, suite.app.EpochsKeeper)
// 	suite.queryClient = types.NewQueryClient(queryHelper)
// }

// func TestKeeperTestSuite(t *testing.T) {
// 	suite.Run(t, new(KeeperTestSuite))
// }
