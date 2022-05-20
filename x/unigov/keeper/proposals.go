package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/Canto-Network/canto/v3/contracts"
	
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)


func AppendLendingMarketproposal(ctx sdk.Context, lm *types.LendingMarketProposal) (*types.LendingMarketProposal, error) {
	
}
func (k Keeper) DeployMapContract(ctx sdk.Context) (common.Address, error) {
	ctorArgs, err := contracts.ProposalStoreContract.ABI.Pack("") //Call empty constructor of Proposal-Store

	if err != nil{
		return common.Address{}, sdkerrors.Wrapf(types.ErrABIPack, "Contract deployment failure: %s", err.Error())
	}

	data := make([]byte, len(contracts.ProposalStore.Bin) + len(ctorArgs))
	copy(data[:(contracts.ProposalStore.Bin)], contracts.ProposalStore.Bin)
	copy(data[(contracts.ProposalStore.Bin):], ctorArgs)

	nonce, err := k.accKeeper.GetSequence(ctx, types.ModuleAddres.Bytes())

	if err != nil {
		return common.Address{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddress, nonce)
	_, err := k.CallEVMWithData(ctx, types.ModuleAdress, nil, data, true)

	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(err, "failed to deploy contract ")
	}
	
	return contractAddr, nil
}
