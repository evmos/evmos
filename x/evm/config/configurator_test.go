// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package config_test

import (
	"testing"

	"github.com/evmos/evmos/v18/x/evm/config"
	"github.com/evmos/evmos/v18/x/evm/core/vm"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v18/x/evm/types"
)

func TestEVMConfigurator(t *testing.T) {
	err := config.NewEVMConfigurator().Apply()

	require.NoError(t, err)
}

func TestExtendedEips(t *testing.T) {
	testCases := []struct {
		name        string
		malleate    func() *config.EVMConfigurator
		expPass     bool
		errContains string
	}{
		{
			"fail - eip already present in activators return an error",
			func() *config.EVMConfigurator {
				extended_eips := map[int]func(*vm.JumpTable){
					3855: func(jt *vm.JumpTable) {},
				}
				ec := config.NewEVMConfigurator().WithExtendedEips(extended_eips)
				return ec
			},
			false,
			"duplicate activation",
		},
		{
			"success - new default extra eips without duplication added",
			func() *config.EVMConfigurator {
				extended_eips := map[int]func(*vm.JumpTable){
					0o000: func(jt *vm.JumpTable) {},
				}
				ec := config.NewEVMConfigurator().WithExtendedEips(extended_eips)
				return ec
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		ec := tc.malleate()
		err := ec.Apply()

		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains, "expected different error")
		}
	}
}

func TestExtendedDefaultExtraEips(t *testing.T) {
	default_extra_eips_snapshot := types.DefaultExtraEIPs
	testCases := []struct {
		name      string
		malleate  func() *config.EVMConfigurator
		postCheck func()
	}{
		{
			"success - empty default extra eip",
			func() *config.EVMConfigurator {
				var extra_default_eips []int64
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extra_default_eips)
				return ec
			},
			func() {
				require.ElementsMatch(t, default_extra_eips_snapshot, types.DefaultExtraEIPs)
			},
		},
		{
			"success - extra default eip added",
			func() *config.EVMConfigurator {
				extra_default_eips := []int64{1000}
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extra_default_eips)
				return ec
			},
			func() {
				require.ElementsMatch(t, append(default_extra_eips_snapshot, 1000), types.DefaultExtraEIPs)
			},
		},
		{
			"success - extra default eip added removing duplicates",
			func() *config.EVMConfigurator {
				extra_default_eips := []int64{1000, 1001}
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extra_default_eips)
				return ec
			},
			func() {
				require.ElementsMatch(t, append(default_extra_eips_snapshot, 1000, 1001), types.DefaultExtraEIPs)
			},
		},
	}

	for _, tc := range testCases {
		ec := tc.malleate()
		err := ec.Apply()

		require.NoError(t, err)

		tc.postCheck()
	}
}
