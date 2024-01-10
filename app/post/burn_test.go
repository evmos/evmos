// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post_test

import (
	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v16/app/post"

	// "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *PostTestSuite) TestPostHandle() {
	testCases := []struct {
		name        string
		tx          func() sdk.Tx
		expPass     bool
		errContains string
		postChecks  func()
	}{
		{
			name: "pass - noop with Ethereum message",
			tx: func() sdk.Tx {
				return s.BuildEthTx()
			},
			expPass:    true,
			postChecks: func() {},
		},
		{
			name: "pass - burn fees of a single token with empty end balance",
			tx: func() sdk.Tx {
				feeAmount := sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "btc"}}
				amount := feeAmount
				s.MintCoinsForFeeCollector(amount)

				return s.BuildCosmosTxWithNSendMsg(1, feeAmount)
			},
			expPass: true,
			postChecks: func() {
				expected := sdk.Coins{}
				balance := s.GetFeeCollectorBalance()
				s.Require().Equal(expected, balance)
			},
		},
		{
			name: "pass - burn fees of a single token with non-empty end balance",
			tx: func() sdk.Tx {
				feeAmount := sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"}}
				amount := sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(20), Denom: "evmos"}}
				s.MintCoinsForFeeCollector(amount)

				return s.BuildCosmosTxWithNSendMsg(1, feeAmount)
			},
			expPass: true,
			postChecks: func() {
				expected := sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"}}
				balance := s.GetFeeCollectorBalance()
				s.Require().Equal(expected, balance)
			},
		},
		{
			name: "pass - burn fees of multiple tokens with empty end balance",
			tx: func() sdk.Tx {
				feeAmount := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "eth"},
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"},
				}
				amount := feeAmount
				s.MintCoinsForFeeCollector(amount)

				return s.BuildCosmosTxWithNSendMsg(1, feeAmount)
			},
			expPass: true,
			postChecks: func() {
				balance := s.GetFeeCollectorBalance()
				s.Require().Equal(sdk.Coins{}, balance)
			},
		},
		{ //nolint:dupl
			name: "pass - burn fees of multiple tokens with non-empty end balance",
			tx: func() sdk.Tx {
				feeAmount := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "btc"},
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"},
				}
				amount := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(20), Denom: "btc"},
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"},
					sdk.Coin{Amount: sdkmath.NewInt(3), Denom: "osmo"},
				}
				s.MintCoinsForFeeCollector(amount)

				return s.BuildCosmosTxWithNSendMsg(1, feeAmount)
			},
			expPass: true,
			postChecks: func() {
				expected := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "btc"},
					sdk.Coin{Amount: sdkmath.NewInt(3), Denom: "osmo"},
				}
				balance := s.GetFeeCollectorBalance()
				s.Require().Equal(expected, balance)
			},
		},
		{ //nolint:dupl
			name: "pass - burn fees of multiple tokens, non-empty end balance, and multiple messages",
			tx: func() sdk.Tx {
				feeAmount := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "btc"},
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"},
				}
				amount := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(20), Denom: "btc"},
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "evmos"},
					sdk.Coin{Amount: sdkmath.NewInt(3), Denom: "osmo"},
				}
				s.MintCoinsForFeeCollector(amount)

				return s.BuildCosmosTxWithNSendMsg(100, feeAmount)
			},
			expPass: true,
			postChecks: func() {
				expected := sdk.Coins{
					sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "btc"},
					sdk.Coin{Amount: sdkmath.NewInt(3), Denom: "osmo"},
				}
				balance := s.GetFeeCollectorBalance()
				s.Require().Equal(expected, balance)
			},
		},
		{
			name: "pass - fees exceeds MaxUint64 (~18 EVMOS). Should not panic",
			tx: func() sdk.Tx {
				amt, ok := sdkmath.NewIntFromString("10000000000000000000000000000000000")
				s.Require().True(ok)
				feeAmount := sdk.Coins{sdk.Coin{Amount: amt, Denom: "evmos"}}
				amount := sdk.Coins{sdk.Coin{Amount: amt, Denom: "evmos"}}
				s.MintCoinsForFeeCollector(amount)

				return s.BuildCosmosTxWithNSendMsg(1, feeAmount)
			},
			expPass: true,
			postChecks: func() {
				expected := sdk.Coins{}
				balance := s.GetFeeCollectorBalance()
				s.Require().Equal(expected, balance)
			},
		},
	}

	for _, tc := range testCases {
		// Be sure to have a fresh new network before each test. It is not required for following
		// test but it is still a good practice.
		s.SetupTest()
		s.Run(tc.name, func() {
			// start each test with a fresh new block.
			err := s.unitNetwork.NextBlock()
			s.Require().NoError(err)

			burnDecorator := post.NewBurnDecorator(
				authtypes.FeeCollectorName,
				s.unitNetwork.App.BankKeeper,
			)

			// In the execution of the PostHandle method, simulate, success, and next have been
			// hard-coded because they are not influencing the behavior of the BurnDecorator.
			terminator := sdk.ChainPostDecorators(sdk.Terminator{})
			_, err = burnDecorator.PostHandle(
				s.unitNetwork.GetContext(),
				tc.tx(),
				false,
				false,
				terminator,
			)

			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err, "expected error during HandlerOptions validation")
				s.Require().Contains(err.Error(), tc.errContains, "expected a different error")
			}

			tc.postChecks()
		})
	}
}
