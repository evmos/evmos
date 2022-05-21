package keeper

import (
	"github.com/Canto-Network/canto/v3/contracts"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Canto-Network/canto/v3/x/unigov/types"

	erc20types "github.com/Canto-Network/canto/v3/x/erc20/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (k Keeper) AppendLendingMarketProposal(ctx sdk.Context, lm *types.LendingMarketProposal) (*types.LendingMarketProposal, error) {
	if err := lm.ValidateBasic(); err != nil {
		return &types.LendingMarketProposal{}, err
	}

	if k.mapContractAddr == common.HexToAddress("0000000000000000000000000000000000000000") {
		if err := k.DeployMapContract(ctx); err != nil {
			return nil, err
		}
	}

	//Any other checks needed for Proposal

	m := lm.GetMetadata()
	
	args, err := contracts.ProposalStoreContract.ABI.Pack(
		"AddProposal", m.GetPropId(), lm.GetTitle(), lm.GetDescription(),
		m.GetAccount(), m.GetValues(), m.GetSignatures(), m.GetCalldatas(),
	)

	data := make([]byte, len(contracts.ProposalStoreContract.Bin)+len(args))
	copy(data[:len(contracts.ProposalStoreContract.Bin)], contracts.ProposalStoreContract.Bin)
	copy(data[len(contracts.ProposalStoreContract.Bin):], args)

	_, err = k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, &k.mapContractAddr, data, true)
	if err != nil {
		return nil, err
	}

	return lm, nil
}

func (k Keeper) DeployMapContract(ctx sdk.Context) error {
	ctorArgs, err := contracts.ProposalStoreContract.ABI.Pack("") //Call empty constructor of Proposal-Store

	if err != nil {
		return sdkerrors.Wrapf(erc20types.ErrABIPack, "Contract deployment failure: %s", err.Error())
	}

	data := make([]byte, len(contracts.ProposalStoreContract.Bin)+len(ctorArgs))
	copy(data[:len(contracts.ProposalStoreContract.Bin)], contracts.ProposalStoreContract.Bin)
	copy(data[len(contracts.ProposalStoreContract.Bin):], ctorArgs)

	nonce, err := k.accKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())

	if err != nil {
		return err
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddress, nonce)
	_, err = k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, nil, data, false)

	if err != nil {
		return sdkerrors.Wrapf(err, "failed to deploy contract ")
	}

	k.mapContractAddr = contractAddr
	return nil
}
