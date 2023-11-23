// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v15_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v15 "github.com/evmos/evmos/v15/app/upgrades/v15"
	evmkeeper "github.com/evmos/evmos/v15/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

func (s *UpgradesTestSuite) TestEnableEIPs() {
	testcases := []struct {
		name        string
		extraEIPs   []int64
		malleate    func(sdk.Context, *evmkeeper.Keeper)
		expEIPs     []int64
		expPass     bool
		errContains string
	}{
		{
			name:      "success - empty EIPs",
			extraEIPs: []int64{},
			expPass:   true,
		},
		{
			name:      "success - single EIP",
			extraEIPs: []int64{3855},
			expEIPs:   []int64{3855},
			expPass:   true,
		},
		{
			name:      "success - multiple EIPs",
			extraEIPs: []int64{3855, 2200, 1884, 1344},
			expEIPs:   []int64{3855, 2200, 1884, 1344},
			expPass:   true,
		},
		{
			name:      "fail - duplicate EIP",
			extraEIPs: []int64{3855, 1344, 2200},
			malleate: func(ctx sdk.Context, ek *evmkeeper.Keeper) {
				params := evmtypes.DefaultParams()
				params.ExtraEIPs = []int64{2200}
				err := ek.SetParams(ctx, params)
				s.Require().NoError(err, "expected no error setting params")
			},
			expEIPs:     []int64{2200}, // NOTE: since the function is failing, we expect the EIPs to remain the same
			expPass:     false,
			errContains: "found duplicate EIP: 2200",
		},
		{
			name:        "fail - invalid EIP",
			extraEIPs:   []int64{3860},
			expEIPs:     []int64{},
			expPass:     false,
			errContains: "is not activateable, valid EIPs are",
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.malleate != nil {
				tc.malleate(s.ctx, s.app.EvmKeeper)
			}

			err := v15.EnableEIPs(s.ctx, s.app.EvmKeeper, tc.extraEIPs...)

			if tc.expPass {
				s.Require().NoError(err, "expected no error enabling EIPs")
			} else {
				s.Require().Error(err, "expected error enabling EIPs")
				s.Require().ErrorContains(err, tc.errContains, "expected different error")
			}

			evmParams := s.app.EvmKeeper.GetParams(s.ctx)
			s.Require().ElementsMatch(tc.expEIPs, evmParams.ExtraEIPs, "expected different EIPs after test")
		})
	}
}
