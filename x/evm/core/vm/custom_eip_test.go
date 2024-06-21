// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestExtendActivators(t *testing.T) {
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
			func() {},
		},
		{
			"success - single new activator",
			map[int]func(*JumpTable){
				0000: func(jt *JumpTable) {},
			},
			true,
			"",
			func() {
				_, ok := activators[0000]
				require.True(t, ok)
			},
		},
		{
			"success - multiple new activators",
			map[int]func(*JumpTable){
				0001: func(jt *JumpTable) {},
				0002: func(jt *JumpTable) {},
			},
			true,
			"",
			func() {
				_, ok := activators[0000]
				require.True(t, ok)
				_, ok = activators[0001]
				require.True(t, ok)
			},
		},
		{
			"fail - only repeated activator",
			map[int]func(*JumpTable){
				3855: func(jt *JumpTable) {},
			},
			false,
			"",
			func() {},
		},
		{
			"fail - repeated activator with valid activator",
			map[int]func(*JumpTable){
				0000: func(jt *JumpTable) {},
				3855: func(jt *JumpTable) {},
			},
			false,
			"",
			func() {
				_, ok := activators[0000]
				require.False(t, ok)
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
	}
}
