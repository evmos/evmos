package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
)

// RegisterCoin deploys an erc20 contract and creates the token pair for the existing cosmos coin
func (k Keeper) DeployLendingMarketContract(ctx sdk.Context, propMetaData govtypes.Proposal) (common.Address, error) {
	// check if the conversion is globally enabled
	if k.mapContractAddr == common.HexToAddress("0000000000000000000000000000000000000000") {
		// the contract has never been deployed, so deploy contract here. Make sure to update the address to the deployed contract.

	} else {
		// the contract has already been deployed, so call the contract to add the proposal
	}
	return &pair, nil
}
