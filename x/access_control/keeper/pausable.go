package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v18/x/access_control/types"
)

func (k Keeper) GetPauser(ctx sdk.Context, contract common.Address) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPauser)
	pauser := store.Get(contract.Bytes())
	if len(pauser) == 0 {
		return common.Address{}, false
	}

	return common.BytesToAddress(pauser), true
}

func (k Keeper) SetPauser(ctx sdk.Context, contract, pauser common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPauser)
	store.Set(contract.Bytes(), pauser.Bytes())
}

func (k Keeper) Paused(ctx sdk.Context, contract common.Address) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPaused)
	return store.Has(contract.Bytes())
}

func (k Keeper) Pause(ctx sdk.Context, contract common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPaused)
	store.Set(contract.Bytes(), []byte{0x01})
}

func (k Keeper) UnPause(ctx sdk.Context, contract common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixPaused)
	store.Delete(contract.Bytes())
}

func (k Keeper) GetPausedContracts(ctx sdk.Context) []common.Address {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixPaused)

	var contracts []common.Address
	for ; iterator.Valid(); iterator.Next() {
		contract := common.BytesToAddress(iterator.Key())
		contracts = append(contracts, contract)
	}

	return contracts
}
