// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/vesting/types"
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
	ctx.KVStore(k.storeKey).Set(key, []byte{})
}

// DeleteGovClawbackDisabled enables the given vesting account address to be clawed back
// via governance by deleting the address from the disabled accounts list.
func (k Keeper) DeleteGovClawbackDisabled(ctx sdk.Context, addr sdk.AccAddress) {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackDisabledKey, addr.Bytes()...)
	ctx.KVStore(k.storeKey).Delete(key)
}

// HasActiveClawbackProposal checks if there is an active clawback proposal for the given
// vesting account address.
func (k Keeper) HasActiveClawbackProposal(ctx sdk.Context, addr sdk.AccAddress) bool {
	key := buildActiveAccountClawbackProposalKey(addr)

	return ctx.KVStore(k.storeKey).Has(key)
}

// SetActiveClawbackProposal sets the given vesting account address as subject to an active governance clawback
// proposal by writing it to store under the corresponding key.
func (k Keeper) SetActiveClawbackProposal(ctx sdk.Context, addr sdk.AccAddress) {
	key := buildActiveAccountClawbackProposalKey(addr)
	ctx.KVStore(k.storeKey).Set(key, []byte{})
}

// DeleteActiveClawbackProposal deletes the entry for the given vesting account address
// from the store, indicating that there is no active governance clawback proposal for it.
func (k Keeper) DeleteActiveClawbackProposal(ctx sdk.Context, addr sdk.AccAddress) {
	key := buildActiveAccountClawbackProposalKey(addr)
	ctx.KVStore(k.storeKey).Delete(key)
}

// buildActiveAccountClawbackProposalKey builds the key for the given account address prefixed with the governance clawback proposal key
func buildActiveAccountClawbackProposalKey(addr sdk.AccAddress) []byte {
	key := make([]byte, 0, len(types.KeyPrefixGovClawbackProposalKey)+len(addr.Bytes()))
	key = append(key, types.KeyPrefixGovClawbackProposalKey...)
	key = append(key, addr.Bytes()...)

	return key
}
