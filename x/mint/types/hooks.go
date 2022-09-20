package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MintHooks defines an interface for mint module's hooks.
type MintHooks interface {
	AfterDistributeMintedCoin(ctx sdk.Context)
}

var _ MintHooks = MultiMintHooks{}

// MultiMintHooks is a container for mint hooks.
// All hooks are run in sequence.
type MultiMintHooks []MintHooks

// NewMultiMintHooks returns new MultiMintHooks given hooks.
func NewMultiMintHooks(hooks ...MintHooks) MultiMintHooks {
	return hooks
}

// AfterDistributeMintedCoin is a hook that runs after minter mints and distributes coins
// at the beginning of each block.
func (h MultiMintHooks) AfterDistributeMintedCoin(ctx sdk.Context) {
	for i := range h {
		h[i].AfterDistributeMintedCoin(ctx)
	}
}
