package keeper

import (
	"fmt"

	"github.com/ArableProtocol/acrechain/x/mint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)
	blockTime := ctx.BlockTime().Unix()

	// skip minting if it's before minting rewards distribution time
	if blockTime < params.MintingRewardsDistributionStartTime {
		return
	}

	// fetch stored minter & params
	minter := k.GetMinter(ctx)

	// if it's the first block after minting rewards distribution start time,
	// skip minting and just set last mint time
	if minter.LastMintTime == 0 {
		minter.LastMintTime = ctx.BlockTime().Unix()
		k.SetMinter(ctx, minter)
		return
	}

	// reduce minting amount when reduction time come
	if blockTime >= params.NextRewardsReductionTime {
		minter.DailyProvisions = minter.DailyProvisions.Mul(params.ReductionFactor)
		k.SetMinter(ctx, minter)
		k.SetNextReductionTime(ctx, blockTime+params.ReductionPeriodInSeconds)
	}

	// mint coins
	mintedCoin := minter.BlockProvision(ctx.BlockTime().Unix(), params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	// update last mint time
	minter.LastMintTime = ctx.BlockTime().Unix()
	k.SetMinter(ctx, minter)

	if mintedCoins.IsAllPositive() {
		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.DistributeMintedCoin(ctx, mintedCoin)
		if err != nil {
			panic(err)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute(types.AttributeBlockNumber, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockProvisions, mintedCoins.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}
