package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

// Keeper defines the IBC fungible transfer keeper
type Keeper struct {
	storeKey sdk.StoreKey

	transferKeeper types.TransferKeeper
	hooks          types.TransferHooks
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(key sdk.StoreKey, tk types.TransferKeeper) Keeper {
	return Keeper{
		storeKey:       key,
		transferKeeper: tk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"-"+types.ModuleName)
}

func (k *Keeper) SetHooks(th types.TransferHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set hooks twice")
	}

	k.hooks = th

	return k
}

// IsTransferHooks returns whether transfer hooks logic should be run.
func (k Keeper) IsTransferHooksEnabled(ctx sdk.Context) bool {
	return k.hooks != nil
	// if k.hooks == nil {
	// 	return false
	// }

	// // TODO: check params
	// return true
}

func (k Keeper) DenomPathFromHash(ctx sdk.Context, denom string) (string, error) {
	return k.transferKeeper.DenomPathFromHash(ctx, denom)
}
