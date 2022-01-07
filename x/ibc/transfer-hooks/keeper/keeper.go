package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/prefix"
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

// SetTransferHooksEnabled sets a flag to determine if transfer hooks logic should run for the given channel
// identified by channel and port identifiers.
func (k Keeper) SetTransferHooksEnabled(ctx sdk.Context, portID, channelID string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PrefixKeyTransferHooksEnabled)
	store.Set(types.KeyTransferHooksEnabled(portID, channelID), []byte{1})
}

// DeleteTransferHooks deletes the transfer hooks enabled flag for a given portID and channelID
func (k Keeper) DeleteTransferHooksEnabled(ctx sdk.Context, portID, channelID string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PrefixKeyTransferHooksEnabled)
	store.Delete(types.KeyTransferHooksEnabled(portID, channelID))
}

// IsTransferHooks returns whether transfer hooks logic should be run for the given port. It will check the
// transfer hooks enabled flag for the given port and channel identifiers
func (k Keeper) IsTransferHooksEnabled(ctx sdk.Context, portID, channelID string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PrefixKeyTransferHooksEnabled)
	return store.Has(types.KeyTransferHooksEnabled(portID, channelID))
}

func (k Keeper) DenomPathFromHash(ctx sdk.Context, denom string) (string, error) {
	return k.transferKeeper.DenomPathFromHash(ctx, denom)
}
