// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types_test

import (
	"testing"

	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestDecimalsValidate(t *testing.T) {
	testCases := []struct {
		decimals evmtypes.Decimals
		expPass  bool
	}{
		{
			decimals: evmtypes.EighteenDecimals,
			expPass:  true,
		},
		{
			decimals: evmtypes.SixDecimals,
			expPass:  true,
		},
		{
			decimals: evmtypes.Decimals(0),
			expPass:  false,
		},
	}

	for _, tc := range testCases {
		err := tc.decimals.Validate()

		if tc.expPass {
			require.NoError(t, err, "expected decimals to be valid")
		} else {
			require.Error(t, err, "expected decimals to be not valid")
		}
	}
}
