package types_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v11/x/claims/types"
	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   types.Params
		expError bool
	}{
		{
			"fail - empty",
			types.Params{},
			true,
		},
		{
			"fail - duration of decay is 0",
			types.Params{DurationOfDecay: 0},
			true,
		},
		{
			"fail - duration until decay is 0",
			types.Params{
				DurationOfDecay:    types.DefaultDurationOfDecay,
				DurationUntilDecay: 0,
			},
			true,
		},
		{
			"fail - invalid claim denom",
			types.Params{
				DurationOfDecay:    types.DefaultDurationOfDecay,
				DurationUntilDecay: types.DefaultDurationUntilDecay,
				ClaimsDenom:        "",
			},
			true,
		},
		{
			"fail - invalid authorized channel",
			types.Params{
				DurationOfDecay:    types.DefaultDurationOfDecay,
				DurationUntilDecay: types.DefaultDurationUntilDecay,
				ClaimsDenom:        types.DefaultClaimsDenom,
				AuthorizedChannels: []string{""},
			},
			true,
		},
		{
			"fail - invalid EVM channel",
			types.Params{
				DurationOfDecay:    types.DefaultDurationOfDecay,
				DurationUntilDecay: types.DefaultDurationUntilDecay,
				ClaimsDenom:        types.DefaultClaimsDenom,
				AuthorizedChannels: types.DefaultAuthorizedChannels,
				EVMChannels:        []string{""},
			},
			true,
		},
		{
			"success - default params",
			types.DefaultParams(),
			false,
		},
		{
			"success - valid params",
			types.Params{
				DurationOfDecay:    types.DefaultDurationOfDecay,
				DurationUntilDecay: types.DefaultDurationUntilDecay,
				ClaimsDenom:        "tevmos",
				AuthorizedChannels: types.DefaultAuthorizedChannels,
				EVMChannels:        types.DefaultEVMChannels,
			},
			false,
		},
		{
			"success - constructor",
			types.NewParams(
				true,
				"tevmos",
				time.Unix(0, 0),
				types.DefaultDurationOfDecay,
				types.DefaultDurationUntilDecay,
				types.DefaultAuthorizedChannels,
				types.DefaultEVMChannels,
			),
			false,
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

func TestParamsvalidateStartDate(t *testing.T) {
	err := validateStartDate(time.Time{})
	require.NoError(t, err)
	err = validateStartDate(time.Now())
	require.NoError(t, err)
	err = validateStartDate("")
	require.Error(t, err)
	err = validateStartDate(int64(123))
	require.Error(t, err)
}

func TestParamsvalidateDuration(t *testing.T) {
	err := validateDuration(time.Hour)
	require.NoError(t, err)
	err = validateDuration(time.Hour * -1)
	require.Error(t, err)
	err = validateDuration("")
	require.Error(t, err)
	err = validateDuration(int64(123))
	require.Error(t, err)
}

func TestParamsValidateDenom(t *testing.T) {
	err := validateDenom("aevmos")
	require.NoError(t, err)
	err = validateDenom(false)
	require.Error(t, err)
	err = validateDenom(int64(123))
	require.Error(t, err)
	err = validateDenom("")
	require.Error(t, err)
}

func TestParamsValidateChannels(t *testing.T) {
	err := types.ValidateChannels([]string{"channel-0"})
	require.NoError(t, err)
	err = types.ValidateChannels(false)
	require.Error(t, err)
	err = types.ValidateChannels(int64(123))
	require.Error(t, err)
	err = types.ValidateChannels("")
	require.Error(t, err)
}

func TestParamsDecayStartTime(t *testing.T) {
	startTime := time.Now().UTC()
	params := types.Params{
		AirdropStartTime:   startTime,
		DurationOfDecay:    time.Hour,
		DurationUntilDecay: time.Hour,
	}

	decayStartTime := params.DecayStartTime()
	require.Equal(t, startTime.Add(time.Hour), decayStartTime)
}

func TestIsClaimsActive(t *testing.T) {
	startTime := time.Now().UTC()
	params := types.Params{
		EnableClaims:       false,
		AirdropStartTime:   startTime,
		DurationOfDecay:    time.Hour,
		DurationUntilDecay: time.Hour,
	}

	res := params.IsClaimsActive(time.Now().UTC())
	require.False(t, res)

	params.EnableClaims = true
	blockTime := startTime.Add(-time.Hour)
	res = params.IsClaimsActive(blockTime)
	require.False(t, res)

	blockTime = startTime.Add(time.Hour)
	res = params.IsClaimsActive(blockTime)
	require.True(t, res)
}

func TestParamsAirdropEndTime(t *testing.T) {
	startTime := time.Now().UTC()
	params := types.Params{
		AirdropStartTime:   startTime,
		DurationOfDecay:    time.Hour,
		DurationUntilDecay: time.Hour,
	}

	endTime := params.AirdropEndTime()
	require.Equal(t, startTime.Add(time.Hour).Add(time.Hour), endTime)
}

func TestIsAuthorizedChannel(t *testing.T) {
	params := types.DefaultParams()
	res := params.IsAuthorizedChannel("")
	require.False(t, res)
	res = params.IsAuthorizedChannel(types.DefaultAuthorizedChannels[0])
	require.True(t, res)
}

func TestIsEVMChannel(t *testing.T) {
	params := types.DefaultParams()
	res := params.IsEVMChannel("")
	require.False(t, res)
	res = params.IsEVMChannel(types.DefaultEVMChannels[0])
	require.True(t, res)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateStartDate(i interface{}) error {
	_, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("duration must be positive: %s", v)
	}

	return nil
}

func validateDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}
