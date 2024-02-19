// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"github.com/ethereum/go-ethereum/common"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	"github.com/evmos/evmos/v16/x/evm/types"
)

func (suite *KeeperTestSuite) TestGetDynamicPrecompilesInstances() {

	params := types.DefaultParams()
	params.ActiveDynamicPrecompiles = []string{erc20precompile.WEVMOSContractMainnet}

	testcases := []struct {
		name     string
		params   types.Params
		expected []common.Address
	}{
		{
			name:     "pass - empty precompiles",
			params:   types.DefaultParams(),
			expected: []common.Address{},
		},
		// {
		// 	TODO: test panic
		// 	name:     "fail - unavailable precompile",
		// 	params:   params,
		// 	expected: []common.Address{common.HexToAddress("0x0000000000000000000000000000000000099999")},
		// },
		{
			name:     "pass - precompile",
			params:   params,
			expected: []common.Address{common.HexToAddress(erc20precompile.WEVMOSContractMainnet)},
		},
	}

	for _, tc := range testcases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			pair := erc20types.NewTokenPair(common.HexToAddress(erc20precompile.WEVMOSContractMainnet), "aevmos", erc20types.OWNER_MODULE)
			suite.app.Erc20Keeper.SetToken(s.ctx, pair)

			addresses, _ := suite.app.EvmKeeper.GetDynamicPrecompilesInstances(s.ctx, &tc.params)
			suite.Require().Equal(tc.expected, addresses)

		})
	}
}

func (suite *KeeperTestSuite) TestIsAvailableDynamicPrecompile() {
	testcases := []struct {
		name         string
		address      string
		expAvailable bool
	}{
		{
			name:         "pass - available precompile",
			address:      erc20precompile.WEVMOSContractMainnet,
			expAvailable: true,
		},
		{
			name:         "fail - unavailable precompile",
			address:      "0x0000000000000000000000000000000000099999",
			expAvailable: false,
		},
	}

	for _, tc := range testcases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			params := types.DefaultParams()
			params.ActiveDynamicPrecompiles = []string{erc20precompile.WEVMOSContractMainnet}
			suite.app.EvmKeeper.SetParams(s.ctx, params)

			available := suite.app.EvmKeeper.IsAvailableDynamicPrecompile(s.ctx, tc.address)
			suite.Require().Equal(tc.expAvailable, available)
		})
	}
}
