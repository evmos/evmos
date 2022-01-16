package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/tharsis/evmos/x/fees/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) RegisterContract(c context.Context, msg *types.MsgRegisterContract) (*types.MsgRegisterContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	contract := common.HexToAddress(msg.ContractAddress)
	deployer := common.HexToAddress(msg.DeployerAddress)

	if k.HasContractWithdrawAddress(ctx, contract) {
		return nil, sdkerrors.Wrap(types.ErrContractAlreadyRegistered, msg.ContractAddress)
	}

	derivedContractAddr := crypto.CreateAddress(deployer, msg.Nonce)

	if contract != derivedContractAddr {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrorInvalidSigner,
			"signer address %s must match deployer address", msg.DeployerAddress,
		)
	}

	k.SetContractWithdrawAddress(ctx, contract, deployer)
	// store reverse index to be able to query all the contracts registered by an address
	k.SetContractWithdrawAddressInverse(ctx, contract, deployer)

	k.Logger(ctx).Info(
		"contract registered for fee distribution",
		"contract-address", msg.ContractAddress,
		"withdraw-address", msg.DeployerAddress,
	)

	return &types.MsgRegisterContractResponse{}, nil
}

func (k Keeper) UpdateWithdawAddress(c context.Context, msg *types.MsgUpdateWithdawAddress) (*types.MsgUpdateWithdawAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	contract := common.HexToAddress(msg.ContractAddress)
	signer := common.HexToAddress(msg.WithdrawAddress)

	withdrawAddr, found := k.GetContractWithdrawAddress(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrContractWithdrawAddrNotFound, msg.ContractAddress)
	}

	if signer != withdrawAddr {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"only txs signed by the registered withdraw address %s are allowed to update records", withdrawAddr,
		)
	}

	newWithdrawAddr := common.HexToAddress(msg.NewWithdrawAddress)

	k.SetContractWithdrawAddress(ctx, contract, newWithdrawAddr)

	// update the inverse index
	k.DeleteContractWithdrawAddressInverse(ctx, contract, withdrawAddr)
	k.SetContractWithdrawAddressInverse(ctx, contract, newWithdrawAddr)

	k.Logger(ctx).Info(
		"contract record updated",
		"contract-address", msg.ContractAddress,
		"withdraw-address", msg.WithdrawAddress,
		"new-withdraw-address", msg.NewWithdrawAddress,
	)

	return &types.MsgUpdateWithdawAddressResponse{}, nil
}
