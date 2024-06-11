// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v18/x/access_control/types"
)

func (k Keeper) GetOwner(ctx sdk.Context, contract common.Address) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOwner)
	owner := store.Get(contract.Bytes())
	if len(owner) == 0 {
		return common.Address{}, false
	}

	return common.BytesToAddress(owner), true
}

func (k Keeper) SetOwner(ctx sdk.Context, contract common.Address, owner common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOwner)
	store.Set(contract.Bytes(), owner.Bytes())
}

// FIXME: set proto file for the ContractAccount
func (k Keeper) GetOwners(ctx sdk.Context) []types.ContractAccount {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOwner)

	var owners []types.ContractAccount
	for ; iterator.Valid(); iterator.Next() {
		contract := common.BytesToAddress(iterator.Key())
		owner := common.BytesToAddress(iterator.Value())
		owners = append(owners, types.ContractAccount{
			Contract: contract.String(),
			Account:  owner.String(),
		})
	}

	return owners
}
