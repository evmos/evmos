// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

// import (
// 	"errors"

// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/evmos/evmos/v18/x/erc20/types"
// 	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
// )

// func (suite *KeeperTestSuite) TestGetDynamicPrecompilesInstances() {
// 	testcases := []struct {
// 		name                         string
// 		dynamicPrecompiles           []string
// 		wrappedNativeCoinPrecompiles []string
// 		expected                     []common.Address
// 		expectPanic                  bool
// 		expectErrorMessage           string
// 	}{
// 		{
// 			name:                         "pass - empty precompiles",
// 			dynamicPrecompiles:           []string{},
// 			wrappedNativeCoinPrecompiles: []string{},
// 			expected:                     []common.Address{},
// 			expectPanic:                  false,
// 			expectErrorMessage:           "",
// 		},
// 		{
// 			name:                         "fail - unavailable precompile",
// 			dynamicPrecompiles:           []string{"0x0000000000000000000000000000000000099999"},
// 			wrappedNativeCoinPrecompiles: []string{},
// 			expected:                     []common.Address{common.HexToAddress("0x0000000000000000000000000000000000099999")},
// 			expectPanic:                  true,
// 			expectErrorMessage:           "precompiled contract not initialized: 0x0000000000000000000000000000000000099999",
// 		},
// 		{
// 			name:                         "pass - precompile",
// 			dynamicPrecompiles:           []string{types.WEVMOSContractMainnet},
// 			wrappedNativeCoinPrecompiles: []string{},
// 			expected:                     []common.Address{common.HexToAddress(types.WEVMOSContractMainnet)},
// 			expectPanic:                  false,
// 			expectErrorMessage:           "",
// 		},
// 	}

// 	for _, tc := range testcases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			defer func() {
// 				var err error
// 				if r := recover(); r != nil {
// 					switch x := r.(type) {
// 					case string:
// 						err = errors.New(x)
// 					case error:
// 						err = x
// 					default:
// 						// Fallback err (per specs, error strings should be lowercase w/o punctuation
// 						err = errors.New("unknown panic")
// 					}
// 					suite.Require().True(tc.expectPanic)
// 					suite.Require().Contains(err.Error(), tc.expectErrorMessage)
// 				}
// 			}()

// 			params := types.DefaultParams()
// 			params.DynamicPrecompiles = tc.dynamicPrecompiles
// 			params.NativePrecompiles = tc.wrappedNativeCoinPrecompiles

// 			pair := erc20types.NewTokenPair(common.HexToAddress(types.WEVMOSContractMainnet), "aevmos", erc20types.OWNER_MODULE)
// 			suite.app.Erc20Keeper.SetToken(s.ctx, pair)

// 			addresses, _ := suite.app.EvmKeeper.GetDynamicPrecompilesInstances(s.ctx, &params)
// 			suite.Require().Equal(tc.expected, addresses)
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestIsAvailableDynamicPrecompile() {
// 	testcases := []struct {
// 		name         string
// 		address      common.Address
// 		expAvailable bool
// 	}{
// 		{
// 			name:         "pass - available precompile",
// 			address:      common.HexToAddress(types.WEVMOSContractMainnet),
// 			expAvailable: true,
// 		},
// 		{
// 			name:         "fail - unavailable precompile",
// 			address:      common.HexToAddress("0x0000000000000000000000000000000000099999"),
// 			expAvailable: false,
// 		},
// 	}

// 	for _, tc := range testcases {
// 		tc := tc

// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			params := types.DefaultParams()
// 			params.DynamicPrecompiles = []string{types.WEVMOSContractMainnet}
// 			err := suite.app.EvmKeeper.SetParams(s.ctx, params)
// 			suite.Require().NoError(err)

// 			available := suite.app.EvmKeeper.IsAvailableDynamicPrecompile(&params, tc.address)
// 			suite.Require().Equal(tc.expAvailable, available)
// 		})
// 	}
// }
