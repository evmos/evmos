package keeper_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

func (suite *KeeperTestSuite) TestGetERC20PrecompileInstance() {
	params := suite.app.Erc20Keeper.GetParams(suite.ctx)
	tokePair := types.NewTokenPair(common.HexToAddress("0x205CF44075E77A3543abC690437F3b2819bc450a"), "test", types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetToken(suite.ctx, tokePair)
	tokenPairs := suite.app.Erc20Keeper.GetTokenPairs(suite.ctx)
	suite.Require().True(len(tokenPairs) > 1)

	testCases := []struct {
		name       string
		paramsFun  func()
		precompile common.Address
		expected   bool
	}{
		{
			"fail - precompile not on params",
			func() {
				params = types.DefaultParams()
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress("0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963"),
			false,
		},
		{
			"fail - precompile on params, but token pair doesnt exist",
			func() {
				params.NativePrecompiles = []string{"0x205CF44075E77A3543abC690437F3b2819bc450a", "0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963"}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress("0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963"),
			false,
		},
		{
			"success - precompile on params, and token pair exist",
			func() {
				params.NativePrecompiles = []string{tokenPairs[0].Erc20Address}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress(tokenPairs[0].Erc20Address),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.paramsFun()

			_, found, _ := suite.app.Erc20Keeper.GetERC20PrecompileInstance(suite.ctx, tc.precompile)
			suite.Require().Equal(found, tc.expected)
		})
	}
}
