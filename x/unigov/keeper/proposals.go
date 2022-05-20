package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/Canto-Network/canto/v3/contracts"
	
	"github.com/Canto-Network/canto/v3/x/unigov/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)


func (k Keeper)AppendLendingMarketproposal(ctx sdk.Context, lm *types.LendingMarketProposal) (*types.LendingMarketProposal, error) {
	if err := lm.ValidateBasic(); err != nil {
		return &types.LendingMarketProposal{}, err
	}

	if k.mapContractAddr == nil {
		if err := k.DeployMapContract(ctx); err != nil {
			return nil, err
		}
	}
	
	//Any other checks needed for Proposal

	args, err := contracts.ProposalStoreContract.ABI.Pack(
		"AddProposal", lm.PropId, lm.GetTitle(), lm.GetDescription(),
		lm.Account, lm.Values, lm.Signatures, lm.Calldatas
	)
	
	data := make([]byte, len(contracts.ProposalStoreContract.Bin) + len(args))
	copy(data[:len(contracts.ProposalStoreContract.Bin)], contracts.ProposalStore.Bin)
	copy(data[len(contracts.ProposalStoreContract.Bin):], args)
	
	_, err := k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, &k.mapContractAddr, data, true)
	if err != nil {
		return nil, err
	}
	
	return lm, nil
}

func (k Keeper) DeployMapContract(ctx sdk.Context) (error) {
	ctorArgs, err := contracts.ProposalStoreContract.ABI.Pack("") //Call empty constructor of Proposal-Store

	if err != nil{
		return common.Address{}, sdkerrors.Wrapf(types.ErrABIPack, "Contract deployment failure: %s", err.Error())
	}

	data := make([]byte, len(contracts.ProposalStore.Bin) + len(ctorArgs))
	copy(data[:len(contracts.ProposalStore.Bin)], contracts.ProposalStore.Bin)
	copy(data[len(contracts.ProposalStore.Bin):], ctorArgs)

	nonce, err := k.accKeeper.GetSequence(ctx, types.ModuleAddres.Bytes())

	if err != nil {
		return err
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddress, nonce)
	_, err := k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAdress, nil, data, false)

	if err != nil {
		return sdkerrors.Wrapf(err, "failed to deploy contract ")
	}
	
	k.mapContractAddr = contractAddr
	return nil
}
