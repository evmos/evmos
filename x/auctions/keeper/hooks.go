// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/evmos/evmos/v18/x/epochs/types"
)

// BeforeEpochStart starts a new auction at the beginning of the epoch
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return
	}

	fmt.Println("AUCTIONS: epoch start", epochIdentifier, epochNumber)
	// TODO: Start new auction
}

// AfterEpochEnd ends the current auction and distributes the rewards to the winner
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// End auction and distribute rewards
	fmt.Println("AUCTIONS: epoch end", epochIdentifier, epochNumber)
	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return
	}

	currentRound := k.GetRound(ctx)
	nextRound := currentRound + 1
	k.SetRound(ctx, nextRound)
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
