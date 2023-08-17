// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v14/x/vesting/types"
)

// HasGovClawbackDisabled checks if the given account has governance clawback disabled.
//
// If an entry exists in the KV store for the given account, the account is NOT subject
// to governance clawback.
func (k Keeper) HasGovClawbackDisabled(ctx sdk.Context, addr sdk.AccAddress) bool {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackDisabledKey, addr.Bytes()...)
	return ctx.KVStore(k.storeKey).Has(key)
}

// SetGovClawbackDisabled disables the given vesting account address to be clawed back
// via governance.
func (k Keeper) SetGovClawbackDisabled(ctx sdk.Context, addr sdk.AccAddress) {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackDisabledKey, addr.Bytes()...)
	ctx.KVStore(k.storeKey).Set(key, []byte{0x01})
}

// DeleteGovClawbackDisabled enables the given vesting account address to be clawed back
// via governance by deleting the address from the disabled accounts list.
func (k Keeper) DeleteGovClawbackDisabled(ctx sdk.Context, addr sdk.AccAddress) {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackDisabledKey, addr.Bytes()...)
	ctx.KVStore(k.storeKey).Delete(key)
}
