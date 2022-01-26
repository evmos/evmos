package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	incentivestypes "github.com/tharsis/evmos/x/incentives/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, coin sdk.Coin) error {
	// Mint coins for distribution
	if err := k.MintCoins(ctx, coin); err != nil {
		panic(err)
	}

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	if err := k.AllocateMintedCoin(ctx, coin); err != nil {
		panic(err)
	}

	// Transfer coins (allocated at genesis) from unvested_team_account to team
	// address
	if err := k.AllocateTeamVesting(ctx); err != nil {
		panic(err)
	}

	return nil
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoin sdk.Coin) error {
	newCoins := sdk.NewCoins(newCoin)

	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AllocateMintedCoins allocates minted coins from the inflation to external
// modules
func (k Keeper) AllocateMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
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
// vesting supply
func (k Keeper) AllocateTeamVesting(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// TODO log instead of error
	// Check unvested team account balance
	unvestedTeamAccount := k.accountKeeper.GetModuleAddress(types.UnvestedTeamAccount)
	balances := k.bankKeeper.GetAllBalances(ctx, unvestedTeamAccount)
	if balances.IsZero() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "%s account has no supply",
			types.UnvestedTeamAccount,
		)
	}

	// Get team vesting provision to allocate from the unvested team account
	// balance.
	mintCoin := sdk.NewCoin(params.MintDenom, params.TeamVestingProvision)

	// Create teamVesitingAmt from balances
	var teamVestingAmt sdk.Coins
	for _, coin := range balances {
		if coin.Denom == mintCoin.Denom {
			// Check if unvested team account has sufficient balance to allocate. If
			// yes, only allocate mintCoin, else allocate remaining balance.
			if coin.IsGTE(mintCoin) {
				coin = mintCoin
			}
		}
		// add coin to teamVestingAmt
		teamVestingAmt = append(teamVestingAmt, coin)
	}

	// Allocate teamVesting to community pool when rewards address is empty
	if params.TeamAddress == "" {
		return k.distrKeeper.FundCommunityPool(ctx, teamVestingAmt, unvestedTeamAccount)
	}

	// Send coins to team address
	teamAddress, err := sdk.AccAddressFromBech32(params.TeamAddress)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.UnvestedTeamAccount,
		teamAddress,
		teamVestingAmt,
	)

	return err
}

// GetProportions gets the balance of the `MintedDenom` from minted coins and
// returns coins according to the `InflationDistribution`.
func (k Keeper) GetProportions(ctx sdk.Context, mintedCoin sdk.Coin, ratio sdk.Dec) sdk.Coin {
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToDec().Mul(ratio).TruncateInt())
}
