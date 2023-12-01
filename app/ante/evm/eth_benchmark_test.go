package evm_test

import (
	"fmt"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethante "github.com/evmos/evmos/v16/app/ante/evm"
	"github.com/evmos/evmos/v16/server/config"
	"github.com/evmos/evmos/v16/testutil"
	testutiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func BenchmarkEthGasConsumeDecorator(b *testing.B) {
	s := new(AnteTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()

	dec := ethante.NewEthGasConsumeDecorator(s.app.BankKeeper, s.app.DistrKeeper, s.app.EvmKeeper, s.app.StakingKeeper, config.DefaultMaxTxGasWanted)

	args := &evmtypes.EvmTxArgs{
		ChainID:  s.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: uint64(1000000),
		GasPrice: big.NewInt(1000000),
	}

	var vmdb *statedb.StateDB

	testCases := []struct {
		name    string
		balance sdkmath.Int
		rewards sdkmath.Int
	}{
		{
			"legacy tx - enough funds to pay for fees",
			sdkmath.NewInt(1e16),
			sdkmath.ZeroInt(),
		},
		{
			"legacy tx - insufficient funds but enough staking rewards to pay for fees",
			sdkmath.ZeroInt(),
			sdkmath.NewInt(1e16),
		},
	}
	b.ResetTimer()

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("Case %s", tc.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				// Stop the timer to perform expensive test setup
				b.StopTimer()
				addr := testutiltx.GenerateAddress()
				args.Accesses = &ethtypes.AccessList{{Address: addr, StorageKeys: nil}}
				tx := evmtypes.NewTx(args)
				tx.From = addr.Hex()

				cacheCtx, _ := s.ctx.CacheContext()
				// Create new stateDB for each test case from the cached context
				vmdb = testutil.NewStateDB(cacheCtx, s.app.EvmKeeper)
				cacheCtx = s.prepareAccount(cacheCtx, addr.Bytes(), tc.balance, tc.rewards)
				s.Require().NoError(vmdb.Commit())

				// Benchmark only the ante handler logic - start the timer
				b.StartTimer()
				_, err := dec.AnteHandle(cacheCtx.WithIsCheckTx(true).WithGasMeter(sdk.NewInfiniteGasMeter()), tx, false, testutil.NextFn)
				s.Require().NoError(err)
			}
		})
	}
}
