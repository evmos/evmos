// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

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
	// Create a cached context that is committed only
	// if not errors happen.
	ctxCache, writeFn := ctx.CacheContext()

	// If it's not the weekly epoch, no-op
	if epochIdentifier != epochstypes.WeekEpochID {
		return
	}

	params := k.GetParams(ctxCache)
	if !params.EnableAuction {
		return
	}

	lastBid := k.GetHighestBid(ctxCache)

	// Distribute the awards from the last auction

	// lastBid.Sender can be "", "invalidBech32" or "validBech32".
	bidWinner, err := sdk.AccAddressFromBech32(lastBid.Sender)

	// A valid bid is one in which lastBid.Sender is "validBech32" and the
	// bid.Amount.Amount is positvie.
	if err == nil && lastBid.Amount.Amount.IsPositive() {
		moduleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
		coins := k.bankKeeper.GetAllBalances(ctxCache, moduleAddress)

		remainingCoins := removeBaseCoinFromCoins(coins)

		// Burn the Evmos Coins from the module account.
		if err := k.bankKeeper.BurnCoins(ctxCache, types.ModuleName, sdk.NewCoins(lastBid.Amount)); err != nil {
			return
		}

		// Send the remaining Coins from the module account to the auction winner.
		fmt.Println("Try to send")
		fmt.Println(len(remainingCoins))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctxCache, types.ModuleName, bidWinner, remainingCoins); err != nil {
			fmt.Println("Error in sending coins")
			return
		}

		// Clear up the bid in the store
		k.deleteBid(ctxCache)
	}

	// If the bid is not valid, we still have to advance round and send funds between the modules.
	currentRound := k.GetRound(ctx)
	nextRound := currentRound + 1
	k.SetRound(ctx, nextRound)

	// Send the entire balance from the Auctions Collector module account to the current Auctions account
	accumulatedCoins := k.bankKeeper.GetAllBalances(ctx, k.accountKeeper.GetModuleAddress(types.AuctionCollectorName))
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.AuctionCollectorName, types.ModuleName, accumulatedCoins); err != nil {
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
