package cosmos_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v15/testutil"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
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
			balance:  math.NewInt(1e18),
			rewards:  []math.Int{math.ZeroInt()},
			simulate: true,
		},
		{
			name:    "insufficient funds but sufficient staking rewards",
			balance: math.ZeroInt(),
			rewards: []math.Int{math.NewInt(1e18)},
			gas:     10_000_000,
		},
		{
			name:     "sufficient balance to pay fees with 10.000 users staking",
			balance:  math.NewInt(1e18),
			rewards:  []math.Int{math.ZeroInt()},
			simulate: true,
			setup: func() {
				var err error
				usersCount := 10_000
				// setup other users rewards
				for i := 0; i < usersCount; i++ {
					userAddr, _ := testutiltx.NewAccAddressAndKey()
					_, err = testutil.PrepareAccountsForDelegationRewards(s.T(), s.network.GetContext(), s.network.App, userAddr, math.ZeroInt(), math.NewInt(1e18))
					s.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				}
				_, err = testutil.Commit(s.network.GetContext(), s.network.App, time.Second*0, nil)
				s.Require().NoError(err)
			},
		},
		{
			name:    "insufficient funds but sufficient staking rewards with 10.000 users staking",
			balance: math.ZeroInt(),
			rewards: []math.Int{math.NewInt(1e18)},
			gas:     10_000_000,
			setup: func() {
				var err error
				usersCount := 10_000
				// setup other users rewards
				for i := 0; i < usersCount; i++ {
					userAddr, _ := testutiltx.NewAccAddressAndKey()
					_, err = testutil.PrepareAccountsForDelegationRewards(s.T(), s.network.GetContext(), s.network.App, userAddr, math.ZeroInt(), math.NewInt(1e18))
					s.Require().NoError(err, "failed to prepare accounts for delegation rewards")
				}
				_, err = testutil.Commit(s.network.GetContext(), s.network.App, time.Second*0, nil)
				s.Require().NoError(err)
			},
		},
		{
			name:    "insufficient funds but sufficient staking rewards - 110 delegations",
			balance: math.ZeroInt(),
			rewards: intSlice(110, math.NewInt(1e14)),
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
				// Stop the timer to perform expensive test setup
				b.StopTimer()
				addr, priv := testutiltx.NewAccAddressAndKey()

				// Create a new DeductFeeDecorator
				dfd, args := s.setupDeductFeeDecoratorTestCase(addr, tc)

				if tc.checkTx {
					s.network.WithCheckTxContext()
				}

				// Create a transaction out of the message
				tx, _ := s.factory.BuildCosmosTx(priv, args)

				// Benchmark only the ante handler logic - start the timer
				b.StartTimer()
				_, err := dfd.AnteHandle(s.network.GetContext(), tx, tc.simulate, testutil.NextFn)
				s.Require().NoError(err)
			}
		})
	}
}
