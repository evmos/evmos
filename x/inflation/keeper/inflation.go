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

	// Allocate minted coins according to allocation proportions
	if err := k.AllocateMintedCoin(ctx, coin); err != nil {
		panic(err)
	}

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
	stakingRewards := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.StakingRewards))
	err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		stakingRewards,
	)
	if err != nil {
		return err
	}

	// Allocate usage incentives to incentives module account
	usageIncentives := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, proportions.UsageIncentives))
	err = k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		incentivestypes.ModuleName,
		usageIncentives,
	)
	if err != nil {
		return err
	}

	// Allocate community pool inflation (remaining module balance) to community
	// pool address
	moduleAddr := sdk.AccAddress(types.ModuleAddress.Bytes())
	communityPool := sdk.NewCoins(k.bankKeeper.GetBalance(ctx, moduleAddr, mintedCoin.Denom))
	err = k.distrKeeper.FundCommunityPool(
		ctx,
		communityPool,
		k.accountKeeper.GetModuleAddress(types.ModuleName),
	)

	return err
}

// AllocateTeamVesting allocates the team vesting proportion from the team
// vesting supply
func (k Keeper) AllocateTeamVesting(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// Check team vesting account balance
	if ok := k.bankKeeper.HasSupply(ctx, params.MintDenom); !ok {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "team vesting account has no supply",
		)
	}

	// Get team vesting provision to allocate from the tharsis account balance. If
	// tharsis account doesn't have sufficient balance to allocate, only allocate
	// remaining coins.
	coin := sdk.NewCoin(params.MintDenom, sdk.NewInt(params.TeamVestingProvision.BigInt().Int64()))
	tharsisAccount := k.accountKeeper.GetModuleAddress(types.TharsisAccount)
	balance := k.bankKeeper.GetBalance(ctx, tharsisAccount, params.MintDenom)

	if balance.IsLT(coin) {
		coin = balance
	}
	coins := sdk.NewCoins(coin)

	// Allocate teamVesting to community pool when rewards address is empty
	if params.TeamAddress == "" {
		if err := k.distrKeeper.FundCommunityPool(ctx, coins, tharsisAccount); err != nil {
			return err
		}
	}

	// Send coins to teamVestingReceiver account
	teamAddress, err := sdk.AccAddressFromHex(params.TeamAddress)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		tharsisAccount.String(),
		teamAddress,
		coins,
	)

	return err
}

// GetProportions gets the balance of the `MintedDenom` from minted coins and
// returns coins according to the `InflationDistribution`
func (k Keeper) GetProportions(ctx sdk.Context, mintedCoin sdk.Coin, ratio sdk.Dec) sdk.Coin {
	params := k.GetParams(ctx)

	// The InflationDistribution param is based on the totalProvision, which
	// includes the teamVestingProvision, that is minted once at genesis. To
	// calculate the proportions based on the MintProvision, the distribution
	// ratio needs to be corrected into the mintRatio.
	//
	// Example:
	// TotalProvision = MintProvision (75%) + TeamVestingProvision (25%)
	// 800M 					= 600M          			+ 200M
	//
	// Adjust ratio to mintedRatio by taking out teamVesting ratio
	// 25% / (1 - 25%) = 0.33
	// 40% / (1 - 25%) = 0.53
	// 10% / (1 - 25%) = 0.13

	// mintRatio = ratio / (1 - teamVestingRatio)
	teamVestingFactor := sdk.OneDec().Sub(params.InflationDistribution.TeamVesting)
	mintRatioBigInt := sdk.OneDec().BigInt().Div(ratio.BigInt(), teamVestingFactor.BigInt())
	mintRatio := sdk.NewDecFromBigInt(mintRatioBigInt)

	// proportion = MintProvision * mintRatio
	proportion := mintedCoin.Amount.ToDec().Mul(mintRatio).TruncateInt()

	return sdk.NewCoin(mintedCoin.Denom, proportion)
}
