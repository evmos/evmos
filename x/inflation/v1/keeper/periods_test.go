package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/stretchr/testify/require"
)

func TestSetGetPeriod(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	expPeriod := uint64(9)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default period",
			func() {},
			false,
		},
		{
			"period set",
			func() {
				nw.App.InflationKeeper.SetPeriod(ctx, expPeriod)
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			period := nw.App.InflationKeeper.GetPeriod(ctx)
			if tc.ok {
				require.Equal(t, expPeriod, period, tc.name)
			} else {
				require.Zero(t, period, tc.name)
			}
		})
	}
}
