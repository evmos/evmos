package types

import (
	"testing"

	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	extraEips := []int64{2929, 1884, 1344}
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid",
			NewParams(DefaultEVMDenom, false, true, true, DefaultChainConfig(), extraEips),
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"invalid evm denom",
			Params{
				EvmDenom: "@!#!@$!@5^32",
			},
			true,
		},
		{
			"invalid eip",
			Params{
				EvmDenom:  DefaultEVMDenom,
				ExtraEIPs: []int64{1},
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestParamsEIPs(t *testing.T) {
	extraEips := []int64{2929, 1884, 1344}
	params := NewParams("ara", false, true, true, DefaultChainConfig(), extraEips)
	actual := params.EIPs()

	require.Equal(t, []int{2929, 1884, 1344}, actual)
}

func TestParamsValidatePriv(t *testing.T) {
	require.Error(t, validateEVMDenom(false))
	require.NoError(t, validateEVMDenom("inj"))
	require.Error(t, validateBool(""))
	require.NoError(t, validateBool(true))
	require.Error(t, validateEIPs(""))
	require.NoError(t, validateEIPs([]int64{1884}))
	require.ErrorContains(t, validateEIPs([]int64{1884, 1884, 1885, 1886}), "duplicate EIP: 1884")
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

func TestIsActivePrecompile(t *testing.T) {
	t.Parallel()

	precompileAddr := "0x0000000000000000000000000000000000000800"

	testCases := []struct {
		name      string
		malleate  func() (Params, string)
		expActive bool
	}{
		{
			name: "inactive precompile",
			malleate: func() (Params, string) {
				return Params{}, precompileAddr
			},
			expActive: false,
		},
		{
			name: "active precompile",
			malleate: func() (Params, string) {
				return Params{ActivePrecompiles: []string{precompileAddr}}, precompileAddr
			},
			expActive: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.malleate, "test case must provide malleate function")
			params, precompile := tc.malleate()

			active := params.IsActivePrecompile(precompile)
			require.Equal(t, tc.expActive, active, "expected different active status for precompile: %s", precompile)
		})
	}
}
