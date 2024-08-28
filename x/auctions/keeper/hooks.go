// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v19/x/auctions/types"
	epochstypes "github.com/evmos/evmos/v19/x/epochs/types"
)

// BeforeEpochStart starts a new auction at the beginning of the epoch
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
	// no-op
}

// AfterEpochEnd ends the current auction and distributes the rewards to the winner
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) {
	// If it's not the weekly epoch, no-op
	if epochIdentifier != epochstypes.WeekEpochID {
		return
	}

	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return
	}

	lastBid := k.GetHighestBid(ctx)

	// Distribute the awards from the last auction

	// lastBid.Sender can be "", "invalidBech32" or "validBech32".
	bidWinner, err := sdk.AccAddressFromBech32(lastBid.Sender)

	// Create a cached context that is committed only
	// if not errors happen handling a valid bid.
	ctxCache, writeFn := ctx.CacheContext()

	// A valid bid is one in which lastBid.Sender is "validBech32" and the
	// bid.Amount.Amount is positvie.
	if err == nil && lastBid.BidValue.Amount.IsPositive() {
		moduleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
		coins := k.bankKeeper.GetAllBalances(ctxCache, moduleAddress)

		remainingCoins := removeBaseCoinFromCoins(coins)

		// Burn the Evmos Coins from the module account
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(lastBid.BidValue)); err != nil {
			k.Logger(ctx).Error("failed to burn coins from Auctions module account", "error", err.Error())
			return
		}

		// Send the remaining Coins from the module account to the auction winner
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidWinner, remainingCoins); err != nil {
			k.Logger(ctx).Error("failed to send coins from Auctions module account to the winner", "error", err.Error())
			return
		}

		// Clear up the bid in the store
		k.deleteBid(ctx)
		currentRound := k.GetRound(ctx)
		if err := EmitAuctionEndEvent(ctx, bidWinner, currentRound, remainingCoins, lastBid.BidValue.Amount); err != nil {
			k.Logger(ctx).Error("failed to emit AuctionEnd event", "error", err.Error())
		}
	}

	// Advance the auction round
	currentRound := k.GetRound(ctx)
	nextRound := currentRound + 1
	k.SetRound(ctx, nextRound)

	// Send the entire balance from the Auctions Collector module account to the current Auctions account
	accumulatedCoins := k.bankKeeper.GetAllBalances(ctx, k.accountKeeper.GetModuleAddress(types.AuctionCollectorName))
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.AuctionCollectorName, types.ModuleName, accumulatedCoins); err != nil {
		k.Logger(ctx).Error("failed to send coins from Auctions Collector to Auctions module account", "error", err)
		return
	}
	writeFn()
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Hooks Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
