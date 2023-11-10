// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post_test

import (
	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v15/app/post"
	// "github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	sdk "github.com/cosmos/cosmos-sdk/types"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
)

func (s *PostTestSuite) TestPostHandle() {
	// from := s.keyring.GetAddr(1)
	from := utiltx.GenerateAddress()
	to := utiltx.GenerateAddress()
	feeAmount := sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "aevmos"}}

	testCases := []struct {
		name        string
		tx          func() sdk.Tx
		expPass     bool
		errContains string
	}{
		{
			name: "pass - noop with Ethereum message",
			tx: func() sdk.Tx {
				return s.BuildEthTx(from, to)
			},
			expPass: true,
		},
		{
			name: "pass - burn fee with Cosmos message",
			tx: func() sdk.Tx {
				// Minting tokens for the FeeCollector to simulate fee accrued.
				s.unitNetwork.App.BankKeeper.MintCoins(
					s.unitNetwork.GetContext(),
					inflationtypes.ModuleName,
					feeAmount,
				)
				s.unitNetwork.App.BankKeeper.SendCoinsFromModuleToModule(
					s.unitNetwork.GetContext(),
					inflationtypes.ModuleName,
					authtypes.FeeCollectorName,
					feeAmount,
				)

				return s.BuildCosmosTx(from, to, feeAmount)
			},
			expPass: true,
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
		})
	}
}
