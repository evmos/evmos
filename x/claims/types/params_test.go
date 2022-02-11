package types

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"fail - empty",
			Params{},
			true,
		},
		{
			"fail - duration of decay is 0",
			Params{DurationOfDecay: 0},
			true,
		},
		{
			"fail - duration until decay is 0",
			Params{
				DurationOfDecay:    DefaultDurationOfDecay,
				DurationUntilDecay: 0,
			},
			true,
		},
		{
			"fail - invalid claim denom",
			Params{
				DurationOfDecay:    DefaultDurationOfDecay,
				DurationUntilDecay: DefaultDurationUntilDecay,
				ClaimsDenom:        "",
			},
			true,
		},
		{
			"success - default params",
			DefaultParams(),
			false,
		},
		{
			"success - valid params",
			Params{
				DurationOfDecay:    DefaultDurationOfDecay,
				DurationUntilDecay: DefaultDurationUntilDecay,
				ClaimsDenom:        "tevmos",
			},
			false,
		},
		{
			"success - constructor",
			NewParams(true, time.Unix(0, 0), "tevmos", DefaultDurationOfDecay, DefaultDurationUntilDecay),
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
	err = validateDuration(int64(123))
	require.Error(t, err)
	err = validateDenom("")
	require.Error(t, err)
}

func TestParamsDecayStartTime(t *testing.T) {
	startTime := time.Now().UTC()
	params := Params{
		AirdropStartTime:   startTime,
		DurationOfDecay:    time.Hour,
		DurationUntilDecay: time.Hour,
	}

	decayStartTime := params.DecayStartTime()
	require.Equal(t, startTime.Add(time.Hour), decayStartTime)
}

func TestIsClaimsActive(t *testing.T) {
	startTime := time.Now().UTC()
	params := Params{
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
	params := Params{
		AirdropStartTime:   startTime,
		DurationOfDecay:    time.Hour,
		DurationUntilDecay: time.Hour,
	}

	endTime := params.AirdropEndTime()
	require.Equal(t, startTime.Add(time.Hour).Add(time.Hour), endTime)
}
