// Copyright Tharsis Labs Ltd.(Evmosii)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post_test

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/evmos/evmos/v15/app/post"
)

func (s *PostTestSuite) TestValidateHandlerOptions() {
	// t.Parallel()

	validBankKeeper := s.network.App.BankKeeper
	validFeeCollector := "fee_collector"

	testCases := []struct {
		name         string
		feeCollector string
		bankKeeper   bankkeeper.Keeper
		expPass      bool
		errContains  string
	}{
		{
			name:         "fail - empty fee collector name",
			feeCollector: "",
			bankKeeper:   validBankKeeper,
			expPass:      false,
			errContains:  "fee collector name cannot be empty",
		},
		{
			name:         "fail - nil bank keeper",
			feeCollector: validFeeCollector,
			bankKeeper:   nil,
			expPass:      false,
			errContains:  "bank keeper cannot be nil",
		},
		{
			name:         "pass - correct inputs",
			feeCollector: validFeeCollector,
			bankKeeper:   validBankKeeper,
			expPass:      true,
		},
	}

	for _, tc := range testCases {
		// Be sure to have a fresh new network before each test. It is not required for following
		// test but it is still a good practice.
		s.SetupTest()
		s.Run(tc.name, func() {
			// start each test with a fresh new block.
			err := s.network.NextBlock()
			s.Require().NoError(err)

			handlerOptions := post.HandlerOptions{
				FeeCollectorName: tc.feeCollector,
				BankKeeper:       tc.bankKeeper,
			}

			err = handlerOptions.Validate()

			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err, "expected error during HandlerOptions validation")
				s.Require().Contains(err.Error(), tc.errContains, "expected a different error")
			}
		})
	}
}
