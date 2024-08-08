package keeper_test

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

func (suite *KeeperTestSuite) TestParams() {
	var ctx sdk.Context

	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				return types.DefaultParams()
			},
			func() interface{} {
				return suite.network.App.Erc20Keeper.GetParams(ctx)
			},
			true,
		},
		{
			"success - Checks if dynamic precompiles are set correctly",
			func() interface{} {
				params := types.DefaultParams()
				params.DynamicPrecompiles = []string{"0xB5124FA2b2cF92B2D469b249433BA1c96BDF536D", "0xC4CcDf91b810a61cCB48b35ccCc066C63bf94B4F"}
				err := suite.network.App.Erc20Keeper.SetParams(ctx, params)
				suite.Require().NoError(err)
				return params.DynamicPrecompiles
			},
			func() interface{} {
				return suite.network.App.Erc20Keeper.GetParams(ctx).DynamicPrecompiles
			},
			true,
		},
		{
			"success - Checks if native precompiles are set correctly",
			func() interface{} {
				params := types.DefaultParams()
				params.NativePrecompiles = []string{"0x205CF44075E77A3543abC690437F3b2819bc450a", "0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963"}
				err := suite.network.App.Erc20Keeper.SetParams(ctx, params)
				suite.Require().NoError(err)
				return params.NativePrecompiles
			},
			func() interface{} {
				return suite.network.App.Erc20Keeper.GetParams(ctx).NativePrecompiles
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx = suite.network.GetContext()

			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}
