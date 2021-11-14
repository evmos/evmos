package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// Keeper of this module maintains collections of intrarelayer.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	govKeeper     types.GovKeeper
	evmKeeper     *evmkeeper.Keeper // TODO: use interface
}

// NewKeeper creates new instances of the intrarelayer Keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	govKeeper types.GovKeeper,
	evmKeeper *evmkeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramstore:    ps,
		accountKeeper: ak,
		bankKeeper:    bk,
		govKeeper:     govKeeper,
		evmKeeper:     evmKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// initHeightQueueCount counts the height transactions (if needed) and store the amount in a cache
func (k *Keeper) getModuleAccountNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixModuleNonce)
	if len(bz) == 0 {
		store.Set(types.KeyPrefixModuleNonce, sdk.Uint64ToBigEndian(uint64(0)))
		return uint64(0)

	}
	nonce := sdk.BigEndianToUint64(bz)
	return nonce
}

func (k *Keeper) increaseModuleAccountNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixModuleNonce, sdk.Uint64ToBigEndian(uint64(nonce+1)))
}
