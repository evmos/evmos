// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"errors"
	"math/big"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	"github.com/evmos/evmos/v20/x/evm/types"
)

var _ statedb.Keeper = &Keeper{}

// ----------------------------------------------------------------------------
// StateDB Keeper implementation
// ----------------------------------------------------------------------------

// GetAccount returns nil if account is not exist
func (k *Keeper) GetAccount(ctx sdk.Context, addr common.Address) *statedb.Account {
	acct := k.GetAccountWithoutBalance(ctx, addr)
	if acct == nil {
		return nil
	}

	acct.Balance = k.GetBalance(ctx, addr)
	return acct
}

// GetState loads contract state from database.
func (k *Keeper) GetState(ctx sdk.Context, addr common.Address, key common.Hash) common.Hash {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))

	value := store.Get(key.Bytes())
	if len(value) == 0 {
		return common.Hash{}
	}

	return common.BytesToHash(value)
}

// GetFastState loads contract state from database.
func (k *Keeper) GetFastState(ctx sdk.Context, addr common.Address, key common.Hash) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))

	return store.Get(key.Bytes())
}

// GetCodeHash loads the code hash from the database for the given contract address.
func (k *Keeper) GetCodeHash(ctx sdk.Context, addr common.Address) common.Hash {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCodeHash)
	bz := store.Get(addr.Bytes())
	if len(bz) == 0 {
		return common.BytesToHash(types.EmptyCodeHash)
	}

	return common.BytesToHash(bz)
}

// IterateContracts iterates over all smart contract addresses in the EVM keeper and
// performs a callback function.
//
// The iteration is stopped when the callback function returns true.
func (k Keeper) IterateContracts(ctx sdk.Context, cb func(addr common.Address, codeHash common.Hash) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixCodeHash)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := common.BytesToAddress(iterator.Key())
		codeHash := common.BytesToHash(iterator.Value())

		if cb(addr, codeHash) {
			break
		}
	}
}

// GetCode loads contract code from database, implements `statedb.Keeper` interface.
func (k *Keeper) GetCode(ctx sdk.Context, codeHash common.Hash) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	return store.Get(codeHash.Bytes())
}

// ForEachStorage iterate contract storage, callback return false to break early
func (k *Keeper) ForEachStorage(ctx sdk.Context, addr common.Address, cb func(key, value common.Hash) bool) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.AddressStoragePrefix(addr)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := common.BytesToHash(iterator.Key())
		value := common.BytesToHash(iterator.Value())

		// check if iteration stops
		if !cb(key, value) {
			return
		}
	}
}

// SetBalance update account's balance, compare with current balance first, then decide to mint or burn.
func (k *Keeper) SetBalance(ctx sdk.Context, addr common.Address, amount *big.Int) error {
	cosmosAddr := sdk.AccAddress(addr.Bytes())

	coin := k.bankWrapper.GetBalance(ctx, cosmosAddr, types.GetEVMCoinDenom())

	delta := new(big.Int).Sub(amount, coin.Amount.BigInt())
	switch delta.Sign() {
	case 1:
		// mint
		if err := k.bankWrapper.MintAmountToAccount(ctx, cosmosAddr, delta); err != nil {
			return err
		}
	case -1:
		// burn
		if err := k.bankWrapper.BurnAmountFromAccount(ctx, cosmosAddr, new(big.Int).Neg(delta)); err != nil {
			return err
		}
	default:
		// not changed
	}
	return nil
}

// SetAccount updates nonce/balance/codeHash together.
func (k *Keeper) SetAccount(ctx sdk.Context, addr common.Address, account statedb.Account) error {
	// update account
	acct := k.accountKeeper.GetAccount(ctx, addr.Bytes())
	if acct == nil {
		acct = k.accountKeeper.NewAccountWithAddress(ctx, addr.Bytes())
	}

	if err := acct.SetSequence(account.Nonce); err != nil {
		return err
	}

	if types.IsEmptyCodeHash(account.CodeHash) {
		k.DeleteCodeHash(ctx, addr)
	} else {
		k.SetCodeHash(ctx, addr.Bytes(), account.CodeHash)
	}
	k.accountKeeper.SetAccount(ctx, acct)

	if err := k.SetBalance(ctx, addr, account.Balance); err != nil {
		return err
	}

	k.Logger(ctx).Debug(
		"account updated",
		"ethereum-address", addr.Hex(),
		"nonce", account.Nonce,
		"codeHash", common.BytesToHash(account.CodeHash).Hex(),
		"balance", account.Balance,
	)
	return nil
}

// SetState update contract storage.
func (k *Keeper) SetState(ctx sdk.Context, addr common.Address, key common.Hash, value []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	store.Set(key.Bytes(), value)

	k.Logger(ctx).Debug(
		"state updated",
		"ethereum-address", addr.Hex(),
		"key", key.Hex(),
	)
}

// DeleteState deletes the entry for the given key in the contract storage
// at the defined contract address.
func (k *Keeper) DeleteState(ctx sdk.Context, addr common.Address, key common.Hash) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressStoragePrefix(addr))
	store.Delete(key.Bytes())

	k.Logger(ctx).Debug(
		"state deleted",
		"ethereum-address", addr.Hex(),
		"key", key.Hex(),
	)
}

// SetCodeHash sets the code hash for the given contract address.
func (k *Keeper) SetCodeHash(ctx sdk.Context, addrBytes, hashBytes []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCodeHash)
	store.Set(addrBytes, hashBytes)

	k.Logger(ctx).Debug(
		"code hash updated",
		"address", common.BytesToAddress(addrBytes).Hex(),
		"code hash", common.BytesToHash(hashBytes).Hex(),
	)
}

// DeleteCodeHash deletes the code hash for the given contract address from the store.
func (k *Keeper) DeleteCodeHash(ctx sdk.Context, addr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCodeHash)
	store.Delete(addr.Bytes())

	k.Logger(ctx).Debug(
		"code hash deleted",
		"address", addr.Hex(),
	)
}

// SetCode sets the given contract code bytes for the corresponding code hash bytes key
// in the code store.
func (k *Keeper) SetCode(ctx sdk.Context, codeHash, code []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	store.Set(codeHash, code)

	k.Logger(ctx).Debug(
		"code updated",
		"code-hash", common.BytesToHash(codeHash).Hex(),
	)
}

// DeleteCode deletes the contract code for the given code hash bytes in
// the corresponding store.
func (k *Keeper) DeleteCode(ctx sdk.Context, codeHash []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCode)
	store.Delete(codeHash)

	k.Logger(ctx).Debug(
		"code deleted",
		"code-hash", common.BytesToHash(codeHash).Hex(),
	)
}

// DeleteAccount handles contract's suicide call:
// - clear balance
// - remove code
// - remove states
// - remove the code hash
// - remove auth account
func (k *Keeper) DeleteAccount(ctx sdk.Context, addr common.Address) error {
	cosmosAddr := sdk.AccAddress(addr.Bytes())
	acct := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	if acct == nil {
		return nil
	}

	// NOTE: only Ethereum contracts can be self-destructed
	if !k.IsContract(ctx, addr) {
		return errors.New("only smart contracts can be self-destructed")
	}

	// clear balance
	if err := k.SetBalance(ctx, addr, new(big.Int)); err != nil {
		return err
	}

	// clear storage
	k.ForEachStorage(ctx, addr, func(key, _ common.Hash) bool {
		k.DeleteState(ctx, addr, key)
		return true
	})

	// clear code hash
	k.DeleteCodeHash(ctx, addr)

	// remove auth account
	k.accountKeeper.RemoveAccount(ctx, acct)

	k.Logger(ctx).Debug(
		"account suicided",
		"ethereum-address", addr.Hex(),
		"cosmos-address", cosmosAddr.String(),
	)

	return nil
}
