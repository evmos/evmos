package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/v5/x/epochs/types"
)

var _ types.EpochHooks = MultiEpochHooks{}

// combine multiple epoch hooks, all hook functions are run in array sequence
type MultiEpochHooks []types.EpochHooks

func NewMultiEpochHooks(hooks ...types.EpochHooks) MultiEpochHooks {
	return hooks
}

// AfterEpochEnd is called when epoch is going to be ended, epochNumber is the
// number of epoch that is ending
func (mh MultiEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range mh {
		mh[i].AfterEpochEnd(ctx, epochIdentifier, epochNumber)
	}
}

// BeforeEpochStart is called when epoch is going to be started, epochNumber is
// the number of epoch that is starting
func (mh MultiEpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range mh {
		mh[i].BeforeEpochStart(ctx, epochIdentifier, epochNumber)
	}
}

// AfterEpochEnd executes the indicated hook after epochs ends
func (k Keeper) AfterEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.AfterEpochEnd(ctx, identifier, epochNumber)
}

// BeforeEpochStart executes the indicated hook before the epochs
func (k Keeper) BeforeEpochStart(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.BeforeEpochStart(ctx, identifier, epochNumber)
}
