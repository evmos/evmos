package keeper_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

func (suite *KeeperTestSuite) TestGetERC20PrecompileInstance() {
	newTokenHexAddr := "0x205CF44075E77A3543abC690437F3b2819bc450a"         //nolint:gosec
	nonExistendTokenHexAddr := "0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963" //nolint:gosec
	newTokenDenom := "test"
	params := suite.app.Erc20Keeper.GetParams(suite.ctx)
	tokenPair := types.NewTokenPair(common.HexToAddress(newTokenHexAddr), newTokenDenom, types.OWNER_MODULE)
	suite.app.Erc20Keeper.SetToken(suite.ctx, tokenPair)
	tokenPairs := suite.app.Erc20Keeper.GetTokenPairs(suite.ctx)
	suite.Require().True(len(tokenPairs) > 1)

	testCases := []struct {
		name          string
		paramsFun     func()
		precompile    common.Address
		expectedFound bool
		expectedError bool
		err           string
	}{
		{
			"fail - precompile not on params",
			func() {
				params = types.DefaultParams()
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress(nonExistendTokenHexAddr),
			false,
			false,
			"",
		},
		{
			"fail - precompile on params, but token pair doesn't exist",
			func() {
				params.NativePrecompiles = []string{newTokenHexAddr, nonExistendTokenHexAddr}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress(nonExistendTokenHexAddr),
			false,
			true,
			"precompiled contract not initialized",
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
			false,
			"",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.paramsFun()

			_, found, err := suite.app.Erc20Keeper.GetERC20PrecompileInstance(suite.ctx, tc.precompile)
			suite.Require().Equal(found, tc.expectedFound)
			if tc.expectedError {
				suite.Require().ErrorContains(err, tc.err)
			}
		})
	}
}
