// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"slices"

	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	"github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *KeeperTestSuite) TestRegisterERC20Extensions() {
	ibcDenom := utils.ComputeIBCDenom("transfer", "channel-0", "uosmo")
	ibcTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), ibcDenom, types.OWNER_MODULE)
	tokenPair := types.NewTokenPair(utiltx.GenerateAddress(), "test", types.OWNER_MODULE)
	otherTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), "other", types.OWNER_MODULE)
	externalTokenPair := types.NewTokenPair(utiltx.GenerateAddress(), "uext", types.OWNER_EXTERNAL)

	testcases := []struct {
		name        string
		malleate    func()
		expPass     bool
		errContains string
		postCheck   func()
	}{
		{
			name:    "pass - no token pairs in ERC20 keeper",
			expPass: true,
			postCheck: func() {
				s.requireActivePrecompiles(evmtypes.AvailableEVMExtensions)
			},
		},
		{
			name: "pass - native token pair",
			malleate: func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, tokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that the precompile was registered
				available := suite.app.EvmKeeper.IsAvailablePrecompile(tokenPair.GetERC20Contract())
				suite.Require().True(available, "expected precompile to be registered")

				// Check that the precompile is set as active
				suite.requireActivePrecompiles(
					append(evmtypes.AvailableEVMExtensions, tokenPair.Erc20Address),
				)
			},
		},
		{
			name: "pass - IBC token pair",
			malleate: func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, ibcTokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that the precompile was registered
				available := suite.app.EvmKeeper.IsAvailablePrecompile(ibcTokenPair.GetERC20Contract())
				suite.Require().True(available, "expected precompile to be registered")

				// Check that the precompile is set as active
				suite.requireActivePrecompiles(
					append(evmtypes.AvailableEVMExtensions, ibcTokenPair.Erc20Address),
				)
			},
		},
		{
			name: "pass - external token pair is skipped",
			malleate: func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, externalTokenPair)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, otherTokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that active precompiles are unchanged
				suite.requireActivePrecompiles(
					append(evmtypes.AvailableEVMExtensions, otherTokenPair.Erc20Address),
				)
			},
		},
		{
			name: "pass - already registered precompile token pair is skipped",
			malleate: func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, tokenPair)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, otherTokenPair)

				tokenPrecompile, err := erc20precompile.NewPrecompile(tokenPair, suite.app.BankKeeper, suite.app.AuthzKeeper, suite.app.TransferKeeper)
				suite.Require().NoError(err, "expected no error creating precompile")

				err = suite.app.EvmKeeper.AddEVMExtensions(suite.ctx, tokenPrecompile)
				suite.Require().NoError(err, "expected no error adding precompile to EVM keeper")
			},
			expPass: true,
			postCheck: func() {
				// Check that active precompiles contain the already registered precompile
				// as well as the other token pair
				expPrecompiles := append(evmtypes.AvailableEVMExtensions, tokenPair.Erc20Address, otherTokenPair.Erc20Address) //nolint:gocritic // Okay not to store to same slice here after appending
				slices.Sort(expPrecompiles)                                                                                    // NOTE: the precompiles are sorted so we need to sort the expected slice as well

				suite.requireActivePrecompiles(expPrecompiles)
			},
		},
		{
			name: "pass - evm denomination deploys werc20 contract",
			malleate: func() {
				params := suite.app.EvmKeeper.GetParams(suite.ctx)
				params.EvmDenom = tokenPair.Denom
				err := suite.app.EvmKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err, "expected no error setting EVM params")

				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, tokenPair)
			},
			expPass: true,
			postCheck: func() {
				// Check that the precompile was not registered
				available := suite.app.EvmKeeper.IsAvailablePrecompile(tokenPair.GetERC20Contract())
				suite.Require().True(available, "expected precompile to be registered")
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			if tc.malleate != nil {
				tc.malleate()
			}

			err := suite.app.Erc20Keeper.RegisterERC20Extensions(suite.ctx)

			if tc.expPass {
				suite.Require().NoError(err, "expected no error registering ERC20 extensions")
			} else {
				suite.Require().Error(err, "expected an error registering ERC20 extensions")
				suite.Require().ErrorContains(err, tc.errContains, "expected different error message")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
