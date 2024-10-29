package erc20

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

// SetContractOwnerAddress sets the contract owner address for the precompile token pair
// and updates the store with the token pair.
func (p *Precompile) SetContractOwnerAddress(ctx sdk.Context, newOwner sdk.AccAddress) error {
	store := prefix.NewStore(ctx.KVStore(p.storeKey), types.KeyPrefixTokenPair)

	p.tokenPair.ContractOwnerAddress = newOwner.String()
	marshaledPair, err := p.tokenPair.Marshal()
	if err != nil {
		return err
	}

	store.Set(p.tokenPair.GetID(), marshaledPair)

	return nil
}

// GetContractOwnerAddress returns the contract owner address for the stored precompile token pair.
// It returns an error if the token pair is not found.
func (p *Precompile) GetContractOwnerAddress(ctx sdk.Context) (sdk.AccAddress, error) {
	store := prefix.NewStore(ctx.KVStore(p.storeKey), types.KeyPrefixTokenPair)

	var tokenPair types.TokenPair
	marshaledPair := store.Get(p.tokenPair.GetID())
	if marshaledPair == nil {
		return nil, fmt.Errorf("token pair not found")
	}

	err := tokenPair.Unmarshal(marshaledPair)
	if err != nil {
		return nil, err
	}

	fmt.Println("tokenPair.ContractOwnerAddress", tokenPair.ContractOwnerAddress)

	return sdk.AccAddressFromBech32(tokenPair.ContractOwnerAddress)
}
