// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	"github.com/evmos/evmos/v19/x/staterent/types"
)

// Keeper of this module maintains collections of staterent and hooks.
type Keeper struct {
	cdc       codec.Codec
	storeKey  storetypes.StoreKey
	evmKeeper *evmkeeper.Keeper
}

// NewKeeper returns a new instance of staterent Keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, evmKeeper *evmkeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		evmKeeper: evmKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
