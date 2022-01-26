package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	incentivestypes "github.com/tharsis/evmos/x/incentives/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, coin sdk.Coin) error {
	// Mint coins for distribution
	if err := k.MintCoins(ctx, coin); err != nil {
		return err
	}

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	if err := k.AllocateExponentialInflation(ctx, coin); err != nil {
		return err
	}

	// Transfer coins (allocated at genesis) from unvested_team_account to team
	// address
	if err := k.AllocateTeamVesting(ctx); err != nil {
		return err
	}

	return nil
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoin sdk.Coin) error {
	newCoins := sdk.NewCoins(newCoin)

	// skip as no coins need to be minted
	if newCoins.Empty() {
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AllocateExponentialInflation allocates coins from the inflation to external
// modules according to allocation proportions:
//   - staking rewards -> sdk `auth` module fee collector
//   - usage incentives -> `x/incentives` module
//   - community pool -> `sdk `distr` module community pool
func (k Keeper) AllocateExponentialInflation(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	proportions := params.InflationDistribution

	// Allocate staking rewards into fee collector account
	stakingRewardsAmt := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.StakingRewards))
	err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		stakingRewardsAmt,
	)
	if err != nil {
		return err
	}

	// Allocate usage incentives to incentives module account
	usageIncentivesAmt := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.UsageIncentives))
	err = k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		incentivestypes.ModuleName,
		usageIncentivesAmt,
	)
	if err != nil {
		return err
	}

	// Allocate community pool amount (remaining module balance) to community
	// pool address
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	communityPoolAmt := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	return k.distrKeeper.FundCommunityPool(
		ctx,
		communityPoolAmt,
		moduleAddr,
	)
}

// AllocateTeamVesting allocates the team vesting proportion from the team
// vesting supply.
func (k Keeper) AllocateTeamVesting(ctx sdk.Context) error {
	logger := k.Logger(ctx)
	params := k.GetParams(ctx)

	// Check unvested team account balances
	unvestedTeamAccount := k.accountKeeper.GetModuleAddress(types.UnvestedTeamAccount)
	balances := k.bankKeeper.GetAllBalances(ctx, unvestedTeamAccount)
	if balances.IsZero() {
		logger.Debug(
			"unvested_team_account account has no supply",
			"balances", balances,
		)
		return nil
	}

	// Get team vesting provision to allocate from the unvested team account
	mintProvision := sdk.NewCoin(params.MintDenom, params.TeamVestingProvision)

	// Create team vesting amount from balances. Any non-mint balances are fully
	// allocated, whereas the mint balance is limited to the mintProvision.
	teamVestingAmt := balances
	mintBalance := balances.AmountOfNoDenomValidation(params.MintDenom)

	// Check if unvested team account has sufficient balance to allocate.
	if mintBalance.GTE(mintProvision.Amount) {
		// Limit allocation to mintProvision
		teamVestingAmt = teamVestingAmt.Sub(sdk.NewCoins(sdk.NewCoin(params.MintDenom, mintBalance)))
		teamVestingAmt = teamVestingAmt.Add(mintProvision)
	} else {
		// Log that only the remaining balance will be allocated
		logger.Debug(
			"insufficient funds",
			"unvested_team_account balance is lower than team provision",
			"balance", mintBalance, "<", "provision", mintProvision,
		)
	}

	// Allocate teamVestingAmt to community pool when rewards address is empty
	if params.TeamAddress == "" {
		logger.Debug(
			"team address not set, transfer allocation to community pool",
			"address", params.TeamAddress,
		)
		return k.distrKeeper.FundCommunityPool(ctx, teamVestingAmt, unvestedTeamAccount)
	}

	// Allocate teamVestingAmt to team address
	teamAddress, err := sdk.AccAddressFromBech32(params.TeamAddress)
	if err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.UnvestedTeamAccount,
		teamAddress,
		teamVestingAmt,
	)
}

// GetAllocationProportion calculates the proportion of coins that is to be
// allocated during inflation for a given distribution.
func (k Keeper) GetProportions(
	ctx sdk.Context,
	coin sdk.Coin,
	distribution sdk.Dec,
) sdk.Coin {
	return sdk.NewCoin(
		coin.Denom,
		coin.Amount.ToDec().Mul(distribution).TruncateInt(),
	)
}
