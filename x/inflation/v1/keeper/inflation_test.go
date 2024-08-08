package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v19/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/inflation/v1/types"
	"github.com/stretchr/testify/require"
)

func TestMintAndAllocateInflation(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	testCases := []struct {
		name                string
		mintCoin            sdk.Coin
		malleate            func()
		expStakingRewardAmt sdk.Coin
		expCommunityPoolAmt sdk.DecCoins
		expPass             bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, math.NewInt(1_000_000)),
			func() {},
			sdk.NewCoin(denomMint, math.NewInt(533_333)),
			sdk.NewDecCoins(sdk.NewDecCoin(denomMint, math.NewInt(466_667))),
			true,
		},
		{
			"pass - no coins minted ",
			sdk.NewCoin(denomMint, math.ZeroInt()),
			func() {},
			sdk.NewCoin(denomMint, math.ZeroInt()),
			sdk.DecCoins(nil),
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			_, _, err := nw.App.InflationKeeper.MintAndAllocateInflation(ctx, tc.mintCoin, types.DefaultParams())
			require.NoError(t, err, tc.name)

			// Get balances
			balanceModule := nw.App.BankKeeper.GetBalance(
				ctx,
				nw.App.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			feeCollector := nw.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			balanceStakingRewards := nw.App.BankKeeper.GetBalance(
				ctx,
				feeCollector,
				denomMint,
			)

			pool, err := nw.App.DistrKeeper.FeePool.Get(ctx)
			require.NoError(t, err)
			balanceCommunityPool := pool.CommunityPool

			if tc.expPass {
				require.NoError(t, err, tc.name)
				require.True(t, balanceModule.IsZero())
				require.Equal(t, tc.expStakingRewardAmt, balanceStakingRewards)
				require.Equal(t, tc.expCommunityPoolAmt, balanceCommunityPool)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetCirculatingSupplyAndInflationRate(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)

	nAccs := int64(1)
	nVals := int64(3)

	// the total bonded tokens for the 4 accounts initialized on the setup (3 validators, 1 EOA)
	bondedAmount := network.DefaultBondedAmount.MulRaw(nVals)                             // Add the allocation for the validators
	bondedAmount = bondedAmount.Add(network.PrefundedAccountInitialBalance.MulRaw(nAccs)) // Add the allocation for the EOA
	bondedCoins := sdk.NewDecCoin(evmostypes.AttoEvmos, bondedAmount)

	testCases := []struct {
		name             string
		bankSupply       math.Int
		malleate         func()
		expInflationRate math.LegacyDec
	}{
		{
			"no epochs per period",
			sdk.TokensFromConsensusPower(400_000_000, evmostypes.PowerReduction).Sub(bondedAmount),
			func() {
				nw.App.InflationKeeper.SetEpochsPerPeriod(ctx, 0)
			},
			math.LegacyZeroDec(),
		},
		{
			"high supply",
			sdk.TokensFromConsensusPower(800_000_000, evmostypes.PowerReduction).Sub(bondedAmount),
			func() {},
			math.LegacyMustNewDecFromStr("5.729166666666666700"),
		},
		{
			"low supply",
			sdk.TokensFromConsensusPower(400_000_000, evmostypes.PowerReduction).Sub(bondedAmount),
			func() {},
			math.LegacyMustNewDecFromStr("17.187500000000000000"),
		},
		{
			"zero circulating supply",
			sdk.TokensFromConsensusPower(200_000_000, evmostypes.PowerReduction).Sub(bondedAmount),
			func() {},
			math.LegacyZeroDec(),
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			keyring := testkeyring.New(int(nAccs))
			nw = network.NewUnitTestNetwork(
				network.WithAmountOfValidators(int(nVals)),
				network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)
			ctx = nw.GetContext()

			// Team allocation is only set on mainnet
			ctx = ctx.WithChainID("evmos_9001-1")
			tc.malleate()

			// Mint coins to increase supply
			coin := sdk.NewCoin(
				types.DefaultInflationDenom,
				tc.bankSupply,
			)
			decCoin := sdk.NewDecCoinFromCoin(coin)
			err := nw.App.InflationKeeper.MintCoins(ctx, coin)
			require.NoError(t, err)

			teamAlloc := sdk.NewDecCoin(
				types.DefaultInflationDenom,
				sdk.TokensFromConsensusPower(int64(200_000_000), evmostypes.PowerReduction),
			)

			circulatingSupply := nw.App.InflationKeeper.GetCirculatingSupply(ctx, types.DefaultInflationDenom)
			require.Equal(t, decCoin.Add(bondedCoins).Sub(teamAlloc).Amount, circulatingSupply)

			inflationRate := nw.App.InflationKeeper.GetInflationRate(ctx, types.DefaultInflationDenom)
			require.Equal(t, tc.expInflationRate, inflationRate)
		})
	}
}

func TestBondedRatio(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	testCases := []struct {
		name         string
		isMainnet    bool
		malleate     func()
		expBondRatio math.LegacyDec
	}{
		{
			"is mainnet",
			true,
			func() {},
			math.LegacyZeroDec(),
		},
		{
			"not mainnet",
			false,
			func() {},
			math.LegacyMustNewDecFromStr("0.000029999100026999"),
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			chainID := utils.MainnetChainID + "-1"
			if !tc.isMainnet {
				chainID = utils.TestnetChainID + "-1"
			}
			// reset
			nw = network.NewUnitTestNetwork(network.WithChainID(chainID))
			ctx = nw.GetContext()

			tc.malleate()

			bondRatio, err := nw.App.InflationKeeper.BondedRatio(ctx)
			require.NoError(t, (err))
			require.Equal(t, tc.expBondRatio, bondRatio)
		})
	}
}
