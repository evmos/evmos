// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidEIPName(t *testing.T) {
	testCases := []struct {
		name        string
		eipName     string
		expPass     bool
		errContains string
	}{
		{
			"fail - invalid number",
			"os_OS",
			false,
			"eip number should be convertible to int",
		},
		{
			"fail - invalid structure, only chain name",
			"os",
			false,
			"eip name does not conform to structure 'chainName_Number'",
		},
		{
			"fail - invalid structure, only number",
			"0000",
			false,
			"eip name does not conform to structure 'chainName_Number'",
		},
		{
			"fail - invalid structure, only delimiter",
			"_",
			false,
			"eip number should be convertible to int",
		},
		{
			"success - valid eip name",
			"os_0000",
			true,
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEIPName(tc.eipName)
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}

func TestExtendActivators(t *testing.T) {
	eips_snapshot := GetActivatorsEipNames()

	testCases := []struct {
		name          string
		newActivators map[string]func(*JumpTable)
		expPass       bool
		errContains   string
		postCheck     func()
	}{
		{
			"success - nil new activators",
			nil,
			true,
			"",
			func() {
				eips := GetActivatorsEipNames()
				require.ElementsMatch(t, eips_snapshot, eips, "expected eips number to be equal")
			},
		},
		{
			"success - single new activator",
			map[string]func(*JumpTable){
				"evmos_0": func(jt *JumpTable) {},
			},
			true,
			"",
			func() {
				eips := GetActivatorsEipNames()
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0"), eips, "expected eips number to be equal")
			},
		},
		{
			"success - multiple new activators",
			map[string]func(*JumpTable){
				"evmos_1": func(jt *JumpTable) {},
				"evmos_2": func(jt *JumpTable) {},
			},
			true,
			"",
			func() {
				eips := GetActivatorsEipNames()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0", "evmos_1", "evmos_2"), eips, "expected eips number to be equal")
			},
		},
		{
			"fail - repeated activator",
			map[string]func(*JumpTable){
				"ethereum_3855": func(jt *JumpTable) {},
			},
			false,
			"",
			func() {
				eips := GetActivatorsEipNames()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0", "evmos_1", "evmos_2"), eips, "expected eips number to be equal")
			},
		},
		{
			"fail - valid activator is not stored if a repeated is present",
			map[string]func(*JumpTable){
				"evmos_3":       func(jt *JumpTable) {},
				"ethereum_3855": func(jt *JumpTable) {},
			},
			false,
			"",
			func() {
				eips := GetActivatorsEipNames()
				// since we are working with a global function, tests are not independent
				require.ElementsMatch(t, append(eips_snapshot, "evmos_0", "evmos_1", "evmos_2"), eips, "expected eips number to be equal")
			},
		},
	}

	for _, tc := range testCases {
		err := ExtendActivators(tc.newActivators)
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains, "expected different error")
		}

		tc.postCheck()
	}
}
