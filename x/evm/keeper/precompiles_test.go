// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	"github.com/ethereum/go-ethereum/common"
	stakingprecompile "github.com/evmos/evmos/v15/precompiles/staking"
)

func (s *KeeperTestSuite) TestIsAvailablePrecompile() {
	testcases := []struct {
		name         string
		address      common.Address
		expAvailable bool
	}{
		{
			name:         "pass - available precompile",
			address:      common.HexToAddress(stakingprecompile.PrecompileAddress),
			expAvailable: true,
		},
		{
			name:         "fail - unavailable precompile",
			address:      common.HexToAddress("0x0000000000000000000000000000000000099999"),
			expAvailable: false,
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			available := s.app.EvmKeeper.IsAvailablePrecompile(tc.address)
			s.Require().Equal(tc.expAvailable, available)
		})
	}
}
