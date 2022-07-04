package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v6/x/fees/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterFee registers a contract to receive transaction fees
func (k Keeper) RegisterFee(
	goCtx context.Context,
	msg *types.MsgRegisterFee,
) (*types.MsgRegisterFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableFees {
		return nil, types.ErrFeesDisabled
	}

	contract := common.HexToAddress(msg.ContractAddress)

	if k.IsFeeRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(
			types.ErrFeesAlreadyRegistered,
			"contract is already registered %s", contract,
		)
	}

	deployer := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	deployerAccount := k.evmKeeper.GetAccountWithoutBalance(ctx, common.BytesToAddress(deployer))
	if deployerAccount == nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrNotFound,
			"deployer account not found %s", msg.DeployerAddress,
		)
	}

	if deployerAccount.IsContract() {
		return nil, sdkerrors.Wrapf(
			types.ErrFeesDeployerIsNotEOA,
			"deployer cannot be a contract %s", msg.DeployerAddress,
		)
	}

	var withdrawer sdk.AccAddress
	if msg.WithdrawerAddress != "" && msg.WithdrawerAddress != msg.DeployerAddress {
		withdrawer = sdk.MustAccAddressFromBech32(msg.WithdrawerAddress)
	}

	derivedContract := common.BytesToAddress(deployer)

	// the contract can be directly deployed by an EOA or created through one
	// or more factory contracts. If it was deployed by an EOA account, then
	// msg.Nonces contains the EOA nonce for the deployment transaction.
	// If it was deployed by one or more factories, msg.Nonces contains the EOA
	// nonce for the origin factory contract, then the nonce of the factory
	// for the creation of the next factory/contract.
	for _, nonce := range msg.Nonces {
		ctx.GasMeter().ConsumeGas(
			params.AddrDerivationCostCreate,
			"fees registration: address derivation CREATE opcode",
		)

		derivedContract = crypto.CreateAddress(derivedContract, nonce)
	}

	if contract != derivedContract {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrorInvalidSigner,
			"not contract deployer or wrong nonce: expected %s instead of %s",
			derivedContract, msg.ContractAddress,
		)
	}

	// contract must already be deployed, to avoid spam registrations
	contractAccount := k.evmKeeper.GetAccountWithoutBalance(ctx, contract)

	if contractAccount == nil || !contractAccount.IsContract() {
		return nil, sdkerrors.Wrapf(
			types.ErrFeesNoContractDeployed,
			"no contract code found at address %s", msg.ContractAddress,
		)
	}

	// prevent storing the same address for deployer and withdrawer
	fee := types.NewFee(contract, deployer, withdrawer)
	k.SetFee(ctx, fee)
	k.SetDeployerMap(ctx, deployer, contract)

	// NOTE: only set withdraw map if address is not empty
	effectiveWithdrawer := msg.DeployerAddress

	if len(withdrawer) != 0 {
		k.SetWithdrawerMap(ctx, withdrawer, contract)
		effectiveWithdrawer = msg.WithdrawerAddress
	}

	k.Logger(ctx).Debug(
		"registering contract for transaction fees",
		"contract", msg.ContractAddress, "deployer", msg.DeployerAddress,
		"withdraw", effectiveWithdrawer,
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeRegisterFee,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, effectiveWithdrawer),
			),
		},
	)

	return &types.MsgRegisterFeeResponse{}, nil
}

// UpdateFee updates the withdraw address of a given Fee. If the given withdraw
// address is empty or the same as the deployer address, the withdraw address is
// removed.
func (k Keeper) UpdateFee(
	goCtx context.Context,
	msg *types.MsgUpdateFee,
) (*types.MsgUpdateFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableFees {
		return nil, types.ErrFeesDisabled
	}

	contract := common.HexToAddress(msg.ContractAddress)
	fee, found := k.GetFee(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrFeesContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	// error if the msg deployer address is not the same as the fee's deployer
	if msg.DeployerAddress != fee.DeployerAddress {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"%s is not the contract deployer", msg.DeployerAddress,
		)
	}

	// fees with the given withdraw address is already registered
	if msg.WithdrawerAddress == fee.WithdrawerAddress {
		return nil, sdkerrors.Wrapf(
			types.ErrFeesAlreadyRegistered,
			"fee with withdraw address %s", msg.WithdrawerAddress,
		)
	}

	// NOTE: withdraw address cannot be empty due to msg stateless validation
	fee.WithdrawerAddress = msg.WithdrawerAddress
	k.SetFee(ctx, fee)
	k.SetWithdrawerMap(
		ctx,
		fee.GetWithdrawerAddr(),
		contract,
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateFee,
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, msg.WithdrawerAddress),
			),
		},
	)

	return &types.MsgUpdateFeeResponse{}, nil
}

// CancelFee deletes the fee for a given contract
func (k Keeper) CancelFee(
	goCtx context.Context,
	msg *types.MsgCancelFee,
) (*types.MsgCancelFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableFees {
		return nil, types.ErrFeesDisabled
	}

	contract := common.HexToAddress(msg.ContractAddress)

	fee, found := k.GetFee(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrFeesContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	if msg.DeployerAddress != fee.DeployerAddress {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"%s is not the contract deployer", msg.DeployerAddress,
		)
	}

	k.DeleteFee(ctx, fee)
	k.DeleteDeployerMap(
		ctx,
		fee.GetDeployerAddr(),
		contract,
	)

	if fee.WithdrawerAddress != "" {
		k.DeleteWithdrawerMap(
			ctx,
			fee.GetWithdrawerAddr(),
			contract,
		)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeCancelFee,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
			),
		},
	)

	return &types.MsgCancelFeeResponse{}, nil
}
