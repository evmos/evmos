// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/evmos/evmos/v18/x/epochs/types"
)

// BeforeEpochStart noop, We don't need to do anything here
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ string, _ int64) {
}

// AfterEpochEnd mints and allocates coins at the end of each epoch end
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)
	if !params.EnableAuction {
		return
	}
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
