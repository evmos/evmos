package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AfterEpochEnd executes the indicated hook after epochs ends
func (k Keeper) AfterEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.AfterEpochEnd(ctx, identifier, epochNumber)
}

// BeforeEpochStart executes the indicated hook before the epochs
func (k Keeper) BeforeEpochStart(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.BeforeEpochStart(ctx, identifier, epochNumber)
}
