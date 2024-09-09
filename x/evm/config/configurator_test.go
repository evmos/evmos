// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package config_test

import (
	"testing"

	"github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v20/x/evm/types"
)

func TestEVMConfigurator(t *testing.T) {
	evmConfigurator := config.NewEVMConfigurator()
	err := evmConfigurator.Configure()
	require.NoError(t, err)

	err = evmConfigurator.Configure()
	require.Error(t, err)
	require.Contains(t, err.Error(), "sealed", "expected different error")
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
				extendedEIPs := map[string]func(*vm.JumpTable){
					"ethereum_3855": func(_ *vm.JumpTable) {},
				}
				ec := config.NewEVMConfigurator().WithExtendedEips(extendedEIPs)
				return ec
			},
			false,
			"duplicate activation",
		},
		{
			"success - new default extra eips without duplication added",
			func() *config.EVMConfigurator {
				extendedEIPs := map[string]func(*vm.JumpTable){
					"evmos_0": func(_ *vm.JumpTable) {},
				}
				ec := config.NewEVMConfigurator().WithExtendedEips(extendedEIPs)
				return ec
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		ec := tc.malleate()
		err := ec.Configure()

		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains, "expected different error")
		}
	}
}

func TestExtendedDefaultExtraEips(t *testing.T) {
	defaultExtraEIPsSnapshot := types.DefaultExtraEIPs
	testCases := []struct {
		name        string
		malleate    func() *config.EVMConfigurator
		postCheck   func()
		expPass     bool
		errContains string
	}{
		{
			"fail - invalid eip name",
			func() *config.EVMConfigurator {
				extendedDefaultExtraEIPs := []string{"os_1_000"}
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extendedDefaultExtraEIPs...)
				return ec
			},
			func() {
				require.ElementsMatch(t, defaultExtraEIPsSnapshot, types.DefaultExtraEIPs)
				types.DefaultExtraEIPs = defaultExtraEIPsSnapshot
			},
			false,
			"eip name does not conform to structure",
		},
		{
			"fail - duplicate default EIP entiries",
			func() *config.EVMConfigurator {
				extendedDefaultExtraEIPs := []string{"os_1000"}
				types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, "os_1000")
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extendedDefaultExtraEIPs...)
				return ec
			},
			func() {
				require.ElementsMatch(t, append(defaultExtraEIPsSnapshot, "os_1000"), types.DefaultExtraEIPs)
				types.DefaultExtraEIPs = defaultExtraEIPsSnapshot
			},
			false,
			"EIP os_1000 is already present",
		},
		{
			"success - empty default extra eip",
			func() *config.EVMConfigurator {
				var extendedDefaultExtraEIPs []string
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extendedDefaultExtraEIPs...)
				return ec
			},
			func() {
				require.ElementsMatch(t, defaultExtraEIPsSnapshot, types.DefaultExtraEIPs)
			},
			true,
			"",
		},
		{
			"success - extra default eip added",
			func() *config.EVMConfigurator {
				extendedDefaultExtraEIPs := []string{"os_1001"}
				ec := config.NewEVMConfigurator().WithExtendedDefaultExtraEIPs(extendedDefaultExtraEIPs...)
				return ec
			},
			func() {
				require.ElementsMatch(t, append(defaultExtraEIPsSnapshot, "os_1001"), types.DefaultExtraEIPs)
				types.DefaultExtraEIPs = defaultExtraEIPsSnapshot
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ec := tc.malleate()
			err := ec.Configure()

			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}

			tc.postCheck()
		})
	}
}
