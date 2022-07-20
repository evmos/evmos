package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v7/x/feesplit/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterFeeSplit registers a contract to receive transaction fees
func (k Keeper) RegisterFeeSplit(
	goCtx context.Context,
	msg *types.MsgRegisterFeeSplit,
) (*types.MsgRegisterFeeSplitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableFeeSplit {
		return nil, types.ErrFeeSplitDisabled
	}

	contract := common.HexToAddress(msg.ContractAddress)

	if k.IsFeeSplitRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(
			types.ErrFeeSplitAlreadyRegistered,
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
			types.ErrFeeSplitDeployerIsNotEOA,
			"deployer cannot be a contract %s", msg.DeployerAddress,
		)
	}

	// contract must already be deployed, to avoid spam registrations
	contractAccount := k.evmKeeper.GetAccountWithoutBalance(ctx, contract)

	if contractAccount == nil || !contractAccount.IsContract() {
		return nil, sdkerrors.Wrapf(
			types.ErrFeeSplitNoContractDeployed,
			"no contract code found at address %s", msg.ContractAddress,
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
			"fee split registration: address derivation CREATE opcode",
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

	// prevent storing the same address for deployer and withdrawer
	feeSplit := types.NewFeeSplit(contract, deployer, withdrawer)
	k.SetFeeSplit(ctx, feeSplit)
	k.SetDeployerMap(ctx, deployer, contract)

	// The effective withdrawer is the withdraw address that is stored after the
	// fee split registration is completed. It defaults to the deployer address if
	// the withdraw address in the msg is omitted. When omitted, the withdraw map
	// dosn't need to be set.
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
				types.EventTypeRegisterFeeSplit,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, effectiveWithdrawer),
			),
		},
	)

	return &types.MsgRegisterFeeSplitResponse{}, nil
}

// UpdateFeeSplit updates the withdraw address of a given FeeSplit. If the given
// withdraw address is empty or the same as the deployer address, the withdraw
// address is removed.
func (k Keeper) UpdateFeeSplit(
	goCtx context.Context,
	msg *types.MsgUpdateFeeSplit,
) (*types.MsgUpdateFeeSplitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableFeeSplit {
		return nil, types.ErrFeeSplitDisabled
	}

	contract := common.HexToAddress(msg.ContractAddress)
	feeSplit, found := k.GetFeeSplit(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrFeeSplitContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	// error if the msg deployer address is not the same as the fee's deployer
	if msg.DeployerAddress != feeSplit.DeployerAddress {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"%s is not the contract deployer", msg.DeployerAddress,
		)
	}

	// check if updating feesplit to default withdrawer
	if msg.WithdrawerAddress == feeSplit.DeployerAddress {
		msg.WithdrawerAddress = ""
	}

	// fee split with the given withdraw address is already registered
	if msg.WithdrawerAddress == feeSplit.WithdrawerAddress {
		return nil, sdkerrors.Wrapf(
			types.ErrFeeSplitAlreadyRegistered,
			"fee split with withdraw address %s", msg.WithdrawerAddress,
		)
	}

	// only delete withdrawer map if is not default
	if feeSplit.WithdrawerAddress != "" {
		k.DeleteWithdrawerMap(ctx, sdk.MustAccAddressFromBech32(feeSplit.WithdrawerAddress), contract)
	}

	// only add withdrawer map if new entry is not default
	if msg.WithdrawerAddress != "" {
		k.SetWithdrawerMap(
			ctx,
			sdk.MustAccAddressFromBech32(msg.WithdrawerAddress),
			contract,
		)
	}
	// update fee split
	feeSplit.WithdrawerAddress = msg.WithdrawerAddress
	k.SetFeeSplit(ctx, feeSplit)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateFeeSplit,
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, msg.WithdrawerAddress),
			),
		},
	)

	return &types.MsgUpdateFeeSplitResponse{}, nil
}

// CancelFeeSplit deletes the FeeSplit for a given contract
func (k Keeper) CancelFeeSplit(
	goCtx context.Context,
	msg *types.MsgCancelFeeSplit,
) (*types.MsgCancelFeeSplitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	if !params.EnableFeeSplit {
		return nil, types.ErrFeeSplitDisabled
	}

	contract := common.HexToAddress(msg.ContractAddress)

	fee, found := k.GetFeeSplit(ctx, contract)
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrFeeSplitContractNotRegistered,
			"contract %s is not registered", msg.ContractAddress,
		)
	}

	if msg.DeployerAddress != fee.DeployerAddress {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"%s is not the contract deployer", msg.DeployerAddress,
		)
	}

	k.DeleteFeeSplit(ctx, fee)
	k.DeleteDeployerMap(
		ctx,
		fee.GetDeployerAddr(),
		contract,
	)

	// delete entry from withdrawer map if not default
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
				types.EventTypeCancelFeeSplit,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
			),
		},
	)

	return &types.MsgCancelFeeSplitResponse{}, nil
}
