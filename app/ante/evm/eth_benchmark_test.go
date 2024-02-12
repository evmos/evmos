package evm_test

import (
	"fmt"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethante "github.com/evmos/evmos/v16/app/ante/evm"
	"github.com/evmos/evmos/v16/testutil"
	testutiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func BenchmarkEthGasConsumeDecorator(b *testing.B) {
	s := new(AnteTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	ctx := s.GetNetwork().GetContext()

	args := &evmtypes.EvmTxArgs{
		ChainID:  s.GetNetwork().App.EvmKeeper.ChainID(),
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

				cacheCtx, _ := ctx.CacheContext()
				// Create new stateDB for each test case from the cached context
				vmdb = testutil.NewStateDB(cacheCtx, s.GetNetwork().App.EvmKeeper)
				cacheCtx = s.prepareAccount(cacheCtx, addr.Bytes(), tc.balance, tc.rewards)
				s.Require().NoError(vmdb.Commit())
				keepers := ethante.ConsumeGasKeepers{
					Bank:         s.GetNetwork().App.BankKeeper,
					Distribution: s.GetNetwork().App.DistrKeeper,
					Evm:          s.GetNetwork().App.EvmKeeper,
					Staking:      s.GetNetwork().App.StakingKeeper,
				}

				baseFee := s.GetNetwork().App.FeeMarketKeeper.GetParams(ctx).BaseFee
				fee := tx.GetEffectiveFee(baseFee.BigInt())
				denom := s.GetNetwork().App.EvmKeeper.GetParams(ctx).EvmDenom
				fees := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(fee)))
				bechAddr := sdk.AccAddress(addr.Bytes())

				// Benchmark only the ante handler logic - start the timer
				b.StartTimer()

				err := ethante.ConsumeFeesAndEmitEvent(
					cacheCtx.WithIsCheckTx(true).WithGasMeter(storetypes.NewInfiniteGasMeter()),
					&keepers,
					fees,
					bechAddr,
				)
				s.Require().NoError(err)
			}
		})
	}
}
