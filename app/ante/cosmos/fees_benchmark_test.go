package cosmos_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	cosmosante "github.com/evmos/evmos/v16/app/ante/cosmos"
	"github.com/evmos/evmos/v16/testutil"
	"github.com/evmos/evmos/v16/testutil/integration/common/factory"
	testutiltx "github.com/evmos/evmos/v16/testutil/tx"
)

// This tests setup contains expensive operations.
// Make sure to run this benchmark tests with a limited number of iterations
// To do so, specify the iteration num with the -benchtime flag
// e.g.: go test -bench=DeductFeeDecorator -benchtime=1000x
func BenchmarkDeductFeeDecorator(b *testing.B) {
	s := new(AnteTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	nw := s.GetNetwork()
	ctx := nw.GetContext()

	testCases := []deductFeeDecoratorTestCase{
		{
			name:     "sufficient balance to pay fees",
			balance:  sdkmath.NewInt(1e18),
			rewards:  []sdkmath.Int{sdkmath.ZeroInt()},
			simulate: true,
		},
		{
			name:    "insufficient funds but sufficient staking rewards",
			balance: sdkmath.ZeroInt(),
			rewards: []sdkmath.Int{sdkmath.NewInt(1e18)},
			gas:     10_000_000,
		},
		{
			name:     "sufficient balance to pay fees with 10.000 users staking",
			balance:  sdkmath.NewInt(1e18),
			rewards:  []sdkmath.Int{sdkmath.ZeroInt()},
			simulate: true,
			setup: func() {
				var err error
				usersCount := 10_000
				// setup other users rewards
				for i := 0; i < usersCount; i++ {
					userAddr, _ := testutiltx.NewAccAddressAndKey()
					ctx, err = testutil.PrepareAccountsForDelegationRewards(s.T(), ctx, nw.App, userAddr, sdkmath.ZeroInt(), sdkmath.NewInt(1e18))
					s.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				}
			},
		},
		{
			name:    "insufficient funds but sufficient staking rewards with 10.000 users staking",
			balance: sdkmath.ZeroInt(),
			rewards: []sdkmath.Int{sdkmath.NewInt(1e18)},
			gas:     10_000_000,
			setup: func() {
				var err error
				usersCount := 10_000
				// setup other users rewards
				for i := 0; i < usersCount; i++ {
					userAddr, _ := testutiltx.NewAccAddressAndKey()
					ctx, err = testutil.PrepareAccountsForDelegationRewards(s.T(), ctx, nw.App, userAddr, sdkmath.ZeroInt(), sdkmath.NewInt(1e18))
					s.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				}
			},
		},
		{
			name:    "insufficient funds but sufficient staking rewards - 110 delegations",
			balance: sdkmath.ZeroInt(),
			rewards: intSlice(110, sdkmath.NewInt(1e14)),
			gas:     10_000_000,
		},
	}

	b.ResetTimer()

	for _, tc := range testCases {
		if tc.setup != nil {
			tc.setup()
		}
		b.Run(fmt.Sprintf("Case: %s", tc.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var (
					dfd  cosmosante.DeductFeeDecorator
					args factory.CosmosTxArgs
				)
				// Stop the timer to perform expensive test setup
				b.StopTimer()
				addr, priv := testutiltx.NewAccAddressAndKey()

				// Create a new DeductFeeDecorator
				ctx, dfd, args = s.setupDeductFeeDecoratorTestCase(addr, priv, tc)

				ctx = ctx.WithIsCheckTx(tc.checkTx)

				// Create a transaction out of the message
				tx, _ := s.GetTxFactory().BuildCosmosTx(priv, args)

				// Benchmark only the ante handler logic - start the timer
				b.StartTimer()
				_, err := dfd.AnteHandle(ctx, tx, tc.simulate, testutil.NextFn)
				s.Require().NoError(err)
			}
		})
	}
}
