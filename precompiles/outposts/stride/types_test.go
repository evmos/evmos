package stride_test

import (
	"testing"

	strideoutpost "github.com/evmos/evmos/v15/precompiles/outposts/stride"
	"github.com/stretchr/testify/require"
)

func TestCreateMemo(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		action      string
		receiver    string
		expPass     bool
		errContains string
	}{
		{
			name:     "success - liquid stake",
			action:   strideoutpost.LiquidStakeAction,
			receiver: "cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
			expPass:  true,
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			memo, err := strideoutpost.CreateMemo(tc.action, tc.receiver)
			if tc.expPass {
				require.NoError(t, err, "expected no error while creating memo")
				require.NotEmpty(t, memo, "expected memo not to be empty")
			} else {
				require.Error(t, err, "expected error while creating memo")
				require.Contains(t, err.Error(), tc.errContains, "expected different error")
			}
		})
	}
}
