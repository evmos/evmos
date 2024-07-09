// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/utils"
	epochstypes "github.com/evmos/evmos/v18/x/epochs/types"
	"github.com/evmos/evmos/v18/x/evm/types"
)

// BeforeEpochStart starts a new auction at the beginning of the epoch
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	fmt.Println("AUCTIONS: epoch start", epochIdentifier, epochNumber)
	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return
	}

	// TODO: Start new auction
}

// AfterEpochEnd ends the current auction and distributes the rewards to the winner
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// If it's not the weekly epoch, no-op
	fmt.Println("AUCTIONS epoch end", epochIdentifier, epochNumber)
	if epochIdentifier != epochstypes.WeekEpochID {
		return
	}

	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return
	}

	lastBid := k.GetHighestBid(ctx)
	lastBidAmount := lastBid.Amount.Amount

	// Distribute the awards from the last auction
	if lastBidAmount.IsPositive() && lastBid.Sender != "" {
		bidWinner, err := sdk.AccAddressFromBech32(lastBid.Sender)
		if err != nil {
			return
		}

		moduleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
		coins := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

		remainingCoins := sdk.NewCoins()
		var evmosCoin sdk.Coin
		for _, coin := range coins {
			if coin.Denom == utils.BaseDenom {
				evmosCoin = coin
			} else {
				remainingCoins = remainingCoins.Add(coin)
			}
		}

		// Burn the Evmos Coins from the module account
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(evmosCoin)); err != nil {
			return
		}

		// Send the remaining Coins from the module account to the auction winner
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidWinner, remainingCoins); err != nil {
			return
		}

		// Clear up the bid in the store
		k.DeleteBid(ctx)
	}

	// Advance the auction round
	currentRound := k.GetRound(ctx)
	nextRound := currentRound + 1
	k.SetRound(ctx, nextRound)

	// TODO: Emit some events here

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
