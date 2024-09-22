// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/evmos/evmos/v20/utils"
	"github.com/evmos/evmos/v20/x/inflation/v1/types"
)

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(
	ctx sdk.Context,
	coin sdk.Coin,
	params types.Params,
) (
	staking, communityPool sdk.Coins,
	err error,
) {
	// skip as no coins need to be minted
	if coin.Amount.IsNil() || !coin.Amount.IsPositive() {
		return nil, nil, nil
	}

	// Mint coins for distribution
	if err := k.MintCoins(ctx, coin); err != nil {
		return nil, nil, err
	}

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	return k.AllocateExponentialInflation(ctx, coin, params)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, coin sdk.Coin) error {
	coins := sdk.Coins{coin}
	return k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
}

// AllocateExponentialInflation allocates coins from the inflation to external
// modules according to allocation proportions:
//   - staking rewards -> sdk `auth` module fee collector
//   - usage incentives -> `x/incentives` module
//   - community pool -> `sdk `distr` module community pool
func (k Keeper) AllocateExponentialInflation(
	ctx sdk.Context,
	mintedCoin sdk.Coin,
	params types.Params,
) (
	staking, communityPool sdk.Coins,
	err error,
) {
	distribution := params.InflationDistribution

	// Allocate staking rewards into fee collector account
	staking = sdk.Coins{k.GetProportions(ctx, mintedCoin, distribution.StakingRewards)}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		staking,
	); err != nil {
		return nil, nil, err
	}

	// Allocate community pool amount (remaining module balance) to community
	// pool address
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	communityPool = k.bankKeeper.GetAllBalances(ctx, moduleAddr)

	err = k.distrKeeper.FundCommunityPool(
		ctx,
		communityPool,
		moduleAddr,
	)
	if err != nil {
		return nil, nil, err
	}

	return staking, communityPool, nil
}

// GetProportions calculates the proportion of coins that is to be
// allocated during inflation for a given distribution.
func (k Keeper) GetProportions(
	_ sdk.Context,
	coin sdk.Coin,
	distribution math.LegacyDec,
) sdk.Coin {
	return sdk.Coin{
		Denom:  coin.Denom,
		Amount: math.LegacyNewDecFromInt(coin.Amount).Mul(distribution).TruncateInt(),
	}
}

// BondedRatio the fraction of the staking tokens which are currently bonded
// It doesn't consider team allocation for inflation
func (k Keeper) BondedRatio(ctx sdk.Context) (math.LegacyDec, error) {
	stakeSupply, err := k.stakingKeeper.StakingTokenSupply(ctx)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	isMainnet := utils.IsMainnet(ctx.ChainID())

	mintDenom := k.GetParams(ctx).MintDenom
	teamAlloc := math.LegacyZeroDec()
	if isMainnet {
		teamAlloc = k.computeTeamAllocation(ctx, mintDenom)
	}

	legacyStakeSupply := math.LegacyNewDecFromInt(stakeSupply)
	if !stakeSupply.IsPositive() || (isMainnet && legacyStakeSupply.LTE(teamAlloc)) {
		return math.LegacyZeroDec(), nil
	}

	// Don't count team allocation in bonded ratio's stake supply
	if isMainnet {
		stakeSupply = legacyStakeSupply.Sub(teamAlloc).RoundInt()
	}

	totalBondedTokens, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return math.LegacyZeroDec(), err
	}
	return math.LegacyNewDecFromInt(totalBondedTokens).QuoInt(stakeSupply), nil
}

// computeTeamAllocation computes the allocation associated with the Evmos team.
// The team allocation is defined as the total mint tokens in the wallets of the
// foundation.
func (k Keeper) computeTeamAllocation(ctx sdk.Context, mintDenom string) math.LegacyDec {
	teamAllocation := math.LegacyZeroDec()
	for _, wallet := range types.FoundationWallets {
		walletAddress := utils.EthHexToCosmosAddr(wallet)
		walletBalance := k.bankKeeper.GetBalance(ctx, walletAddress, mintDenom)
		teamAllocation = teamAllocation.Add(math.LegacyNewDecFromInt(walletBalance.Amount))
	}
	return teamAllocation
}

// GetCirculatingSupply returns the bank supply of the mintDenom excluding the
// team allocation in the first year
func (k Keeper) GetCirculatingSupply(ctx sdk.Context, mintDenom string) math.LegacyDec {
	circulatingSupply := math.LegacyNewDecFromInt(k.bankKeeper.GetSupply(ctx, mintDenom).Amount)

	// Consider team allocation only on mainnet chain id. Team funds are held in
	// the foundation wallets. When the foundation distribute tokens to the
	// team, they will not be in these wallets anymore and will be account for
	// the in the circulating supply.
	if utils.IsMainnet(ctx.ChainID()) {
		foundationAllocation := k.computeTeamAllocation(ctx, mintDenom)
		circulatingSupply = circulatingSupply.Sub(foundationAllocation)
	}

	return circulatingSupply
}

// GetInflationRate returns the inflation rate for the current period.
func (k Keeper) GetInflationRate(ctx sdk.Context, mintDenom string) math.LegacyDec {
	epp := k.GetEpochsPerPeriod(ctx)
	if epp == 0 {
		return math.LegacyZeroDec()
	}

	epochMintProvision := k.GetEpochMintProvision(ctx)
	if epochMintProvision.IsZero() {
		return math.LegacyZeroDec()
	}

	epochsPerPeriod := math.LegacyNewDec(epp)

	circulatingSupply := k.GetCirculatingSupply(ctx, mintDenom)
	if circulatingSupply.IsZero() {
		return math.LegacyZeroDec()
	}

	// EpochMintProvision * 365 / circulatingSupply * 100
	return epochMintProvision.Mul(epochsPerPeriod).Quo(circulatingSupply).Mul(math.LegacyNewDec(100))
}

// GetEpochMintProvision retrieves necessary params KV storage
// and calculate EpochMintProvision
func (k Keeper) GetEpochMintProvision(ctx sdk.Context) math.LegacyDec {
	bondedRadio, err := k.BondedRatio(ctx)
	// ?? Should we bubble up this error instead of returning 0 ??
	if err != nil {
		return math.LegacyZeroDec()
	}
	return types.CalculateEpochMintProvision(
		k.GetParams(ctx),
		k.GetPeriod(ctx),
		k.GetEpochsPerPeriod(ctx),
		bondedRadio,
	).Quo(math.LegacyNewDec(types.ReductionFactor))
}
