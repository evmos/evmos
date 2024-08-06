// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendActivators(t *testing.T) {
	eips_snapshot := GetActivatorsEipNumbers()

	testCases := []struct {
		name           string
		new_activators map[int]func(*JumpTable)
		expPass        bool
		errContains    string
		postCheck      func()
	}{
		{
			"success - nil new activators",
			nil,
			true,
			"",
			func() {
				eips := GetActivatorsEipNumbers()
				require.ElementsMatch(t, eips_snapshot, eips, "expected eips number to be equal")
			},
		},
		{
			"success - single new activator",
			map[int]func(*JumpTable){
				0o000: func(jt *JumpTable) {},
			},
			true,
			"",
			func() {
				eips := GetActivatorsEipNumbers()
				require.ElementsMatch(t, append(eips_snapshot, 0o000), eips, "expected eips number to be equal")
			},
		},
		{
			"success - multiple new activators",
			map[int]func(*JumpTable){
				0o001: func(jt *JumpTable) {},
				0o002: func(jt *JumpTable) {},
			},
			true,
			"",
			func() {
				eips := GetActivatorsEipNumbers()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, 0o000, 0o001, 0o002), eips, "expected eips number to be equal")
			},
		},
		{
			"fail - repeated activator",
			map[int]func(*JumpTable){
				3855: func(jt *JumpTable) {},
			},
			false,
			"",
			func() {
				eips := GetActivatorsEipNumbers()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, 0o000, 0o001, 0o002), eips, "expected eips number to be equal")
			},
		},
		{
			"fail - valid activator is not stored if a repeated is present",
			map[int]func(*JumpTable){
				0o003: func(jt *JumpTable) {},
				3855:  func(jt *JumpTable) {},
			},
			false,
			"",
			func() {
				eips := GetActivatorsEipNumbers()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, 0o000, 0o001, 0o002), eips, "expected eips number to be equal")
			},
		},
	}

	for _, tc := range testCases {
		err := ExtendActivators(tc.new_activators)
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains, "expected different error")
		}

		tc.postCheck()
	}
}
