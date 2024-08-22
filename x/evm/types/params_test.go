package types

import (
	"testing"

	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	t.Parallel()

	extraEips := []string{"ethereum_2929", "ethereum_1884", "ethereum_1344"}
	testCases := []struct {
		name        string
		params      Params
		expPass     bool
		errContains string
	}{
		{
			name:    "default",
			params:  DefaultParams(),
			expPass: true,
		},
		{
			name:    "valid",
			params:  NewParams(DefaultEVMDenom, false, DefaultChainConfig(), extraEips, nil, nil, DefaultAccessControl, DefaultDenomDecimals),
			expPass: true,
		},
		{
			name:        "empty",
			params:      Params{},
			errContains: "invalid denom: ", // NOTE: this returns the first error that occurs
		},
		{
			name: "invalid evm denom",
			params: Params{
				EvmDenom: "@!#!@$!@5^32",
			},
			errContains: "invalid denom: @!#!@$!@5^32",
		},
		{
			name: "invalid eip",
			params: Params{
				EvmDenom:  DefaultEVMDenom,
				ExtraEIPs: []string{"os_1000000"},
			},
			errContains: "EIP os_1000000 is not activateable, valid EIPs are",
		},
		{
			name: "unsorted precompiles",
			params: Params{
				EvmDenom: DefaultEVMDenom,
				ActiveStaticPrecompiles: []string{
					"0x0000000000000000000000000000000000000801",
					"0x0000000000000000000000000000000000000800",
				},
			},
			errContains: "precompiles need to be sorted",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !tc.expPass {
				// NOTE: check that the necessary information is provided. Otherwise, a false
				// error message could be accepted when checking for an empty string.
				require.NotEmpty(t, tc.errContains, "expected test case to provide expected error message")
			}

			err := tc.params.Validate()

			if tc.expPass {
				require.NoError(t, err, "expected parameters to be valid")
			} else {
				require.Error(t, err, "expected parameters to be invalid")
				require.ErrorContains(t, err, tc.errContains, "expected different error message")
			}
		})
	}
}

func TestParamsEIPs(t *testing.T) {
	extraEips := []string{"ethereum_2929", "ethereum_1884", "ethereum_1344"}
	params := NewParams("ara", false, DefaultChainConfig(), extraEips, nil, nil, DefaultAccessControl, DefaultDenomDecimals)
	actual := params.EIPs()

	require.Equal(t, []string{"ethereum_2929", "ethereum_1884", "ethereum_1344"}, actual)
}

func TestParamsValidatePriv(t *testing.T) {
	require.Error(t, validateEVMDenom(false))
	require.NoError(t, validateEVMDenom("inj"))
	require.Error(t, validateBool(""))
	require.NoError(t, validateBool(true))
	require.Error(t, validateEIPs(""))
	require.Error(t, validateEIPs([]int64{1884}))
	require.NoError(t, validateEIPs([]string{"ethereum_1884"}))
	require.ErrorContains(t, validateEIPs([]string{"ethereum_1884", "ethereum_1884", "ethereum_1885"}), "duplicate EIP: ethereum_1884")
	require.NoError(t, validateChannels([]string{"channel-0"}))
	require.Error(t, validateChannels(false))
	require.Error(t, validateChannels(int64(123)))
	require.Error(t, validateChannels(""))
}

func TestValidateChainConfig(t *testing.T) {
	testCases := []struct {
		name     string
		i        interface{}
		expError bool
	}{
		{
			"invalid chain config type",
			"string",
			true,
		},
		{
			"valid chain config type",
			DefaultChainConfig(),
			false,
		},
	}
	for _, tc := range testCases {
		err := validateChainConfig(tc.i)

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestIsLondon(t *testing.T) {
	testCases := []struct {
		name   string
		height int64
		result bool
	}{
		{
			"Before london block",
			5,
			false,
		},
		{
			"After london block",
			12_965_001,
			true,
		},
		{
			"london block",
			12_965_000,
			true,
		},
	}

	for _, tc := range testCases {
		ethConfig := ethparams.MainnetChainConfig
		require.Equal(t, IsLondon(ethConfig, tc.height), tc.result)
	}
}
