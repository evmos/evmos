package v2_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tharsis/evmos/v2/app"
	erc20Migrations "github.com/tharsis/evmos/v2/x/erc20/migrations/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.Evmos
}

func (suite *UpgradeTestSuite) SetupTest() {
	// setup feemarketGenesis params
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	suite.app = app.Setup(false, feemarketGenesis)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: "evmos_9001-2",
	})
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestUpdateEVMHooks() {
	testCases := []struct {
		msg        string
		preUpdate  func()
		update     func()
		postUpdate func()
		expPass    bool
	}{
		{
			"Test ERC20 module migrations",
			func() {
				erc20Params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				erc20Params.EnableEVMHook = false
				erc20Params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, erc20Params)

				suite.Require().False(suite.app.Erc20Keeper.GetParams(suite.ctx).EnableErc20)
				suite.Require().False(suite.app.Erc20Keeper.GetParams(suite.ctx).EnableEVMHook)
			},
			func() {
				erc20Migrations.UpdateParams(suite.ctx, suite.app.Erc20Keeper)
			},
			func() {
				erc20Params := suite.app.Erc20Keeper.GetParams(suite.ctx)
				suite.Require().True(erc20Params.EnableErc20)
				suite.Require().True(erc20Params.EnableEVMHook)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.preUpdate()
			tc.update()
			tc.postUpdate()
		})
	}
}
