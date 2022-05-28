package keeper

import (
	"math/big"
	"github.com/Canto-Network/canto/v3/contracts"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"fmt"
	
	"github.com/Canto-Network/canto/v3/x/unigov/types"
	
	erc20types "github.com/Canto-Network/canto/v3/x/erc20/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)


//method for appending UniGov proposal types to the Unigov Map contract
func (k *Keeper) AppendLendingMarketProposal(ctx sdk.Context, lm *types.LendingMarketProposal) (*types.LendingMarketProposal, error) {
	m := lm.GetMetadata()
	var err error
	m.PropId, err = k.govKeeper.GetProposalID(ctx)

	if err != nil {
		return nil, sdkerrors.Wrap(err, "Error obtaining Proposal ID")
	}
	
	nonce, err := k.accKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if nonce == 0 {

		if err != nil {
			return nil, sdkerrors.Wrap(err, "error obtaining account nonce")
		}

		*k.mapContractAddr, err = k.DeployMapContract(ctx, lm)
		if err != nil {
			return nil, err
		}
		fmt.Println("\n\n\n")
		fmt.Println(*k.mapContractAddr)
		fmt.Println("\n\n\n")
		return lm, nil
	}

	_, err = k.erc20Keeper.CallEVM(ctx, contracts.ProposalStoreContract.ABI, types.ModuleAddress, *k.mapContractAddr, true,
	    "AddProposal", sdk.NewIntFromUint64(m.GetPropId()).BigInt(), lm.GetTitle(), lm.GetDescription(), ToAddress(m.GetAccount()),
	    ToBigInt(m.GetValues()), m.GetSignatures(), ToBytes(m.GetCalldatas()))


	if err != nil {
		return nil, sdkerrors.Wrap(err, "Error in EVM Call")
	}
	
	return lm, nil
}

func (k Keeper) DeployMapContract(ctx sdk.Context, lm *types.LendingMarketProposal) (common.Address, error) {

	m := lm.GetMetadata()

	ctorArgs, err := contracts.ProposalStoreContract.ABI.Pack("", sdk.NewIntFromUint64(m.GetPropId()).BigInt(), lm.GetTitle(), lm.GetDescription(), ToAddress(m.GetAccount()),
		ToBigInt(m.GetValues()), m.GetSignatures(), ToBytes(m.GetCalldatas())) //Call empty constructor of Proposal-Store

	if err != nil {
		return common.Address{}, sdkerrors.Wrapf(erc20types.ErrABIPack, "Contract deployment failure: %s", err.Error())
	}

	data := make([]byte, len(contracts.ProposalStoreContract.Bin)+len(ctorArgs))
	copy(data[:len(contracts.ProposalStoreContract.Bin)], contracts.ProposalStoreContract.Bin)
	copy(data[len(contracts.ProposalStoreContract.Bin):], ctorArgs)

	nonce, err := k.accKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())

	if err != nil {
		return common.Address{}, sdkerrors.Wrap(err, "failure in obtaining account nonce")
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddress, nonce)
	_, err = k.erc20Keeper.CallEVMWithData(ctx, types.ModuleAddress, nil, data, true)
	
	if err != nil {
		return common.Address{}, sdkerrors.Wrap(err, "failed to deploy contract")
	}

	return contractAddr, nil
}


func ToAddress(addrs []string) []common.Address {
	if addrs == nil {
		return make([]common.Address, 0)
	}

	arr := make([]common.Address, len(addrs))

	for i,v := range addrs {
		arr[i] = common.HexToAddress(v)
	}

	return arr
}

func ToBytes(strs []string) [][]byte {
	if strs == nil {
		return make([][]byte, 0)
	}
	
	arr := make([][]byte, len(strs))

	for i, v := range strs {
		arr[i] = common.Hex2Bytes(v)
	}
	return arr
}

func ToBigInt(ints []uint64) []*big.Int {
	if ints == nil {
		return make([]*big.Int, 0)
	}
	
	arr := make([]*big.Int, len(ints))

	for i, a := range ints {
		arr[i] = sdk.NewIntFromUint64(a).BigInt()
	}

	return arr
}
	
