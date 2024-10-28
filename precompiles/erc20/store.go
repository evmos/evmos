package erc20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetContractOwnerAddress sets the contract owner address for the precompile token pair
// and updates the store with the token pair.
func (p *Precompile) SetContractOwnerAddress(ctx sdk.Context, newOwner sdk.AccAddress) error {
	store := ctx.KVStore(p.storeKey)


	p.tokenPair.ContractOwnerAddress = newOwner.String()
	marshaledPair, err := p.tokenPair.Marshal()
	if err != nil {
		return err
	}
	
	store.Set(p.tokenPair.GetID(), marshaledPair)

	return nil
}