package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	devShares := sdk.NewDecWithPrec(60, 2)
	validatorShares := sdk.NewDecWithPrec(40, 2)
	derivCostCreate := uint64(50)

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid: enabled",
			NewParams(true, devShares, validatorShares, derivCostCreate),
			false,
		},
		{
			"valid: disabled",
			NewParams(false, devShares, validatorShares, derivCostCreate),
			false,
		},
		{
			"valid: 100% devs",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(1)), sdk.NewDecFromInt(sdk.NewInt(0)), derivCostCreate},
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"invalid: share > 1",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(2)), sdk.NewDecFromInt(sdk.NewInt(0)), derivCostCreate},
			true,
		},
		{
			"invalid: share < 0",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(-1)), sdk.NewDecFromInt(sdk.NewInt(0)), derivCostCreate},
			true,
		},
		{
			"invalid: sum shares > 1 ",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(1)), sdk.NewDecFromInt(sdk.NewInt(1)), derivCostCreate},
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

func TestParamsValidateBool(t *testing.T) {
	err := validateBool(true)
	require.NoError(t, err)
	err = validateBool(false)
	require.NoError(t, err)
	err = validateBool("")
	require.Error(t, err)
	err = validateBool(int64(123))
	require.Error(t, err)
}

func TestParamsValidateUint64(t *testing.T) {
	err := validateUint64(uint64(0))
	require.NoError(t, err)
	err = validateUint64(uint64(1))
	require.NoError(t, err)
	err = validateUint64("")
	require.Error(t, err)
	err = validateUint64(int64(123))
	require.Error(t, err)
	err = validateUint64(int64(-1))
	require.Error(t, err)
}
