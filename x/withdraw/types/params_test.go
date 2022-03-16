package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"empty params",
			Params{},
			false,
		},
		{
			"default params",
			DefaultParams(),
			false,
		},
		{
			"custom params",
			NewParams(true, time.Hour),
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
