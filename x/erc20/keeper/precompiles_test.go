package keeper_test

import (
<<<<<<< HEAD
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

func (suite *KeeperTestSuite) TestGetERC20PrecompileInstance() {
	var (
		ctx        sdk.Context
		tokenPairs []types.TokenPair
	)
	newTokenHexAddr := "0x205CF44075E77A3543abC690437F3b2819bc450a"         //nolint:gosec
	nonExistendTokenHexAddr := "0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963" //nolint:gosec
	newTokenDenom := "test"
	tokenPair := types.NewTokenPair(common.HexToAddress(newTokenHexAddr), newTokenDenom, types.OWNER_MODULE)
=======
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
>>>>>>> main

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
<<<<<<< HEAD
				params := types.DefaultParams()
				err := suite.network.App.Erc20Keeper.SetParams(ctx, params)
=======
				params = types.DefaultParams()
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
>>>>>>> main
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
<<<<<<< HEAD
				params := types.DefaultParams()
				params.NativePrecompiles = []string{newTokenHexAddr, nonExistendTokenHexAddr}
				err := suite.network.App.Erc20Keeper.SetParams(ctx, params)
=======
				params.NativePrecompiles = []string{newTokenHexAddr, nonExistendTokenHexAddr}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
>>>>>>> main
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
<<<<<<< HEAD
				params := types.DefaultParams()
				params.NativePrecompiles = []string{tokenPair.Erc20Address}
				err := suite.network.App.Erc20Keeper.SetParams(ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress(tokenPair.Erc20Address),
=======
				params.NativePrecompiles = []string{tokenPairs[0].Erc20Address}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
			},
			common.HexToAddress(tokenPairs[0].Erc20Address),
>>>>>>> main
			true,
			false,
			"",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
<<<<<<< HEAD
			suite.SetupTest()
			ctx = suite.network.GetContext()

			suite.network.App.Erc20Keeper.SetToken(ctx, tokenPair)
			tokenPairs = suite.network.App.Erc20Keeper.GetTokenPairs(ctx)
			suite.Require().True(len(tokenPairs) > 1)

			tc.paramsFun()

			_, found, err := suite.network.App.Erc20Keeper.GetERC20PrecompileInstance(ctx, tc.precompile)
=======
			tc.paramsFun()

			_, found, err := suite.app.Erc20Keeper.GetERC20PrecompileInstance(suite.ctx, tc.precompile)
>>>>>>> main
			suite.Require().Equal(found, tc.expectedFound)
			if tc.expectedError {
				suite.Require().ErrorContains(err, tc.err)
			}
		})
	}
}
