package cosmos_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/testutil"
	testutiltx "github.com/evmos/evmos/v12/testutil/tx"
)

// This tests setup contains expensive operations.
// Make sure to run this benchmark tests with a limited number of iterations
// To do so, specify the iteration num with the -benchtime flag
// e.g.: go test -bench=DeductFeeDecorator -benchtime=1000x
func BenchmarkDeductFeeDecorator(b *testing.B) {
	s := new(AnteTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	testCases := []deductFeeDecoratorTestCase{
		{
			name:     "sufficient balance to pay fees",
			balance:  sdk.NewInt(1e18),
			rewards:  sdk.ZeroInt(),
			simulate: true,
		},
		{
			name:    "insufficient funds but sufficient staking rewards",
			balance: sdk.ZeroInt(),
			rewards: sdk.NewInt(1e18),
			gas:     10_000_000,
		},
	}

	b.ResetTimer()

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("Case: %s", tc.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				// Stop the timer to perform expensive test setup
				b.StopTimer()
				addr, priv := testutiltx.NewAccAddressAndKey()

				// Create a new DeductFeeDecorator
				dfd, args := s.setupDeductFeeDecoratorTestCase(addr, priv, tc)

				s.ctx = s.ctx.WithIsCheckTx(tc.checkTx)

				// Create a transaction out of the message
				tx, _ := testutiltx.PrepareCosmosTx(s.ctx, s.app, args)

				// Benchmark only the ante handler logic - start the timer
				b.StartTimer()
				_, err := dfd.AnteHandle(s.ctx, tx, tc.simulate, testutil.NextFn)
				s.Require().NoError(err)
			}
		})
	}
}
