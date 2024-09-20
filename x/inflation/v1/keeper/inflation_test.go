package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/utils"
	"github.com/evmos/evmos/v20/x/inflation/v1/types"
	"github.com/stretchr/testify/require"
)

const (
	nAccs = int64(1)
	nVals = int64(3)
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

	foundationAcc := []sdk.AccAddress{
		utils.EthHexToCosmosAddr(types.FoundationWallets[0]),
		utils.EthHexToCosmosAddr(types.FoundationWallets[1]),
	}
	teamAllocation := network.PrefundedAccountInitialBalance.MulRaw(int64(len(foundationAcc)))

	// Genesis available tokens are defined by the testing suite setup:
	//- Validators' self delegation.
	//- Tokens delegated by only one EOA.
	//- Free EOA tokens.
	valBondedAmt := network.DefaultBondedAmount.MulRaw(nVals)
	accsBondAmount := math.OneInt().MulRaw(nVals)
	bondedAmount := valBondedAmt.Add(accsBondAmount)

	testCases := []struct {
		name       string
		bankSupply math.Int
		malleate   func()
	}{
		{
			"no epochs per period",
			sdk.TokensFromConsensusPower(400_000_000, evmostypes.PowerReduction),
			func() {
				nw.App.InflationKeeper.SetEpochsPerPeriod(ctx, 0)
			},
		},
		{
			"high supply",
			sdk.TokensFromConsensusPower(800_000_000, evmostypes.PowerReduction),
			func() {},
		},
		{
			"low supply",
			sdk.TokensFromConsensusPower(400_000_000, evmostypes.PowerReduction),
			func() {},
		},
		{
			"zero circulating supply",
			sdk.TokensFromConsensusPower(200_000_000, evmostypes.PowerReduction),
			func() {},
		},
	}
	for _, isTestnet := range []bool{false, true} {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("Case %s, mainnet = %t", tc.name, !isTestnet), func(t *testing.T) {
				// This variable consider all non bonded tokens during genesis but the team
				// allocation.
				accsFreeAmount := network.PrefundedAccountInitialBalance.MulRaw(nAccs).Sub(accsBondAmount)

				chainID := utils.MainnetChainID + "-1"
				if isTestnet {
					chainID = utils.TestnetChainID + "-1"
					accsFreeAmount = accsFreeAmount.Add(teamAllocation)
				}
				// reset
				keyring := testkeyring.New(int(nAccs))
				nw = network.NewUnitTestNetwork(
					network.WithChainID(chainID),
					network.WithAmountOfValidators(int(nVals)),
					network.WithPreFundedAccounts(append(keyring.GetAllAccAddrs(), foundationAcc...)...),
				)
				ctx = nw.GetContext()

				tc.malleate()

				// Mint coins to increase supply
				coin := sdk.NewCoin(
					denomMint,
					tc.bankSupply,
				)
				err := nw.App.InflationKeeper.MintCoins(ctx, coin)
				require.NoError(t, err)

				circulatingSupply := nw.App.InflationKeeper.GetCirculatingSupply(ctx, denomMint)
				expCirculatingSupply := math.LegacyNewDecFromInt(tc.bankSupply.Add(bondedAmount).Add(accsFreeAmount))
				require.Equal(t, expCirculatingSupply, circulatingSupply)

				epp := nw.App.InflationKeeper.GetEpochsPerPeriod(ctx)
				epochsPerPeriod := math.LegacyNewDec(epp)

				// If epochs per period is equal to zero we have a division by
				// zero in the computai
				epochMintProvision := nw.App.InflationKeeper.GetEpochMintProvision(ctx)

				expInflationRate := epochMintProvision.Mul(epochsPerPeriod).Quo(expCirculatingSupply).Mul(math.LegacyNewDec(100))

				inflationRate := nw.App.InflationKeeper.GetInflationRate(ctx, denomMint)
				require.Equal(t, expInflationRate, inflationRate)
			})
		}
	}
}

func TestBondedRatio(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)

	foundationAcc := []sdk.AccAddress{
		utils.EthHexToCosmosAddr(types.FoundationWallets[0]),
		utils.EthHexToCosmosAddr(types.FoundationWallets[1]),
	}
	teamAllocation := network.PrefundedAccountInitialBalance.MulRaw(int64(len(foundationAcc)))

	valBondedAmt := network.DefaultBondedAmount.MulRaw(nVals)
	accsBondAmount := math.OneInt().MulRaw(nVals)
	bondedAmount := valBondedAmt.Add(accsBondAmount)

	testCases := []struct {
		name      string
		isMainnet bool
	}{
		{
			"is mainnet",
			true,
		},
		{
			"not mainnet",
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			totalSupply := network.PrefundedAccountInitialBalance.MulRaw(nAccs).Sub(accsBondAmount).Add(bondedAmount)

			chainID := utils.MainnetChainID + "-1"
			if !tc.isMainnet {
				chainID = utils.TestnetChainID + "-1"
				totalSupply = totalSupply.Add(teamAllocation)
			}

			// reset
			keyring := testkeyring.New(int(nAccs))
			nw = network.NewUnitTestNetwork(
				network.WithChainID(chainID),
				network.WithAmountOfValidators(int(nVals)),
				network.WithPreFundedAccounts(append(keyring.GetAllAccAddrs(), foundationAcc...)...),
			)
			ctx = nw.GetContext()

			expBondedRatio := math.LegacyNewDecFromInt(bondedAmount).QuoInt(totalSupply)
			bondRatio, err := nw.App.InflationKeeper.BondedRatio(ctx)
			require.NoError(t, (err))
			require.Equal(t, expBondedRatio, bondRatio)
		})
	}
}
