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
