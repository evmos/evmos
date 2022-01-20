package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/types"
	// poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
)

// TODO Refactor into one function that performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, coin sdk.Coin) error {

	return nil
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// DistributeMintedCoins implements distribution of minted coins from mint to external modules.
func (k Keeper) DistributeMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	// allocate staking incentives into fee collector account to be moved to on next begin blocker by staking module
	stakingIncentivesCoins := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.Staking))
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, stakingIncentivesCoins)
	if err != nil {
		return err
	}

	// TODO replace poolIncentives with usageIncentives
	// // allocate pool allocation ratio to pool-incentives module account account
	// poolIncentivesCoins := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.PoolIncentives))
	// err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, poolincentivestypes.ModuleName, poolIncentivesCoins)
	// if err != nil {
	// 	return err
	// }

	devRewardCoin := k.GetProportions(ctx, mintedCoin, proportions.DeveloperRewards)
	devRewardCoins := sdk.NewCoins(devRewardCoin)
	// This is supposed to come from the developer vesting module address, not the mint module address
	// we over-allocated to the mint module address earlier though, so we burn it right here.
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, devRewardCoins)
	if err != nil {
		return err
	}
	if len(params.WeightedDeveloperRewardsReceivers) == 0 {
		// fund community pool when rewards address is empty
		err = k.distrKeeper.FundCommunityPool(ctx, devRewardCoins, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
		if err != nil {
			return err
		}
	} else {
		// allocate developer rewards to addresses by weight
		for _, w := range params.WeightedDeveloperRewardsReceivers {
			devRewardPortionCoins := sdk.NewCoins(k.GetProportions(ctx, devRewardCoin, w.Weight))
			if w.Address == "" {
				err = k.distrKeeper.FundCommunityPool(ctx, devRewardPortionCoins,
					k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
				if err != nil {
					return err
				}
			} else {
				devRewardsAddr, err := sdk.AccAddressFromBech32(w.Address)
				if err != nil {
					return err
				}
				// If recipient is vesting account, pay to account according to its vesting condition
				err = k.bankKeeper.SendCoinsFromModuleToAccount(
					ctx, types.DeveloperVestingModuleAcctName, devRewardsAddr, devRewardPortionCoins)
				if err != nil {
					return err
				}
			}
		}
	}

	// TODO substraction logic needed?
	// // subtract from original provision to ensure no coins left over after the allocations
	// communityPoolCoins := sdk.NewCoins(mintedCoin).Sub(stakingIncentivesCoins).Sub(poolIncentivesCoins).Sub(devRewardCoins)
	// err = k.distrKeeper.FundCommunityPool(ctx, communityPoolCoins, k.accountKeeper.GetModuleAddress(types.ModuleName))
	// if err != nil {
	// 	return err
	// }

	// call an hook after the minting and distribution of new coins
	k.hooks.AfterDistributeMintedCoin(ctx, mintedCoin)

	return err
}

// GetProportions gets the balance of the `MintedDenom` from minted coins and returns coins according to the `AllocationRatio`
func (k Keeper) GetProportions(ctx sdk.Context, mintedCoin sdk.Coin, ratio sdk.Dec) sdk.Coin {
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToDec().Mul(ratio).TruncateInt())
}
