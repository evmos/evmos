package upgrade_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v17/app"
	"github.com/evmos/evmos/v17/encoding"
	"github.com/evmos/evmos/v17/tests/e2e/upgrade"
	"github.com/stretchr/testify/require"
)

func TestUnpackBalancesResponse(t *testing.T) {
	t.Parallel()

	expAmount, ok := sdk.NewIntFromString("1000000000000000000000")
	require.True(t, ok, "failed to convert amount to int")

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	protoCodec, ok := encodingConfig.Codec.(*codec.ProtoCodec)
	require.True(t, ok, "failed to cast codec to proto codec")

	testcases := []struct {
		name        string
		output      string
		want        sdk.Coins
		expPass     bool
		errContains string
	}{
		{
			name: "success",
			output: fmt.Sprintf(
				`{"balances":[{"denom":"aevmos","amount":"%s"}],`+
					`"pagination":{"next_key":null,"total":"0"}}`,
				expAmount,
			),
			want:    sdk.Coins{sdk.NewCoin("aevmos", expAmount)},
			expPass: true,
		},
		{
			name:    "pass - empty balances",
			output:  `{"balances":[],"pagination":{"next_key":null,"total":"0"}}`,
			want:    sdk.Coins{},
			expPass: true,
		},
		{
			name:        "fail - invalid output",
			output:      `invalid`,
			errContains: "failed to unmarshal balances",
		},
	}

	for _, tc := range testcases {
		tc := tc // Added for parallel testing, check https://pkg.go.dev/testing#hdr-Subtests_and_Sub_benchmarks
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := upgrade.UnpackBalancesResponse(protoCodec, tc.output)
			if tc.expPass {
				require.NoError(t, err, "unexpected error")
				require.Equal(t, tc.want, got, "expected different balances")
			} else {
				require.Error(t, err, "expected error but got none")
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}
