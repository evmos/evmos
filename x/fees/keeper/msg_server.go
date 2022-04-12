package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

var _ types.MsgServer = &Keeper{}

// RegisterDevFeeInfo registers a contract to receive transaction fees
func (k Keeper) RegisterDevFeeInfo(
	goCtx context.Context,
	msg *types.MsgRegisterDevFeeInfo,
) (*types.MsgRegisterDevFeeInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.isEnabled(ctx) {
		return nil, sdkerrors.Wrapf(types.ErrInternalFee, "fees module is not enabled")
	}

	contract := common.HexToAddress(msg.ContractAddress)
	if k.IsFeeRegistered(ctx, contract) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract is already registered %s", contract)
	}

	deployer, _ := sdk.AccAddressFromBech32(msg.DeployerAddress)
	deployerAccount := k.evmKeeper.GetAccountWithoutBalance(ctx, common.BytesToAddress(deployer.Bytes()))
	if deployerAccount == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deployer account not found %s", msg.DeployerAddress)
	}
	if deployerAccount.IsContract() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "deployer cannot be a contract %s", msg.DeployerAddress)
	}

	var withdrawal sdk.AccAddress
	if msg.WithdrawAddress != "" {
		withdrawal, _ = sdk.AccAddressFromBech32(msg.WithdrawAddress)
	}

	// the contract can be directly deployed by an EOA or created through one
	// or more factory contracts. If it was deployed by an EOA account, then
	// msg.Nonces contains the EOA nonce for the deployment transaction.
	// If it was deployed by one or more factories, msg.Nonces contains the EOA
	// nonce for the origin factory contract, then the nonce of the factory
	// for the creation of the next factory/contract.
	addrDerivationCostCreate := k.GetParams(ctx).AddrDerivationCostCreate
	derivedContractAddr := common.BytesToAddress(deployer)
	for _, nonce := range msg.Nonces {
		ctx.GasMeter().ConsumeGas(
			addrDerivationCostCreate,
			"fees registration: address derivation CREATE opcode",
		)
		derivedContractAddr = crypto.CreateAddress(derivedContractAddr, nonce)
	}

	if contract != derivedContractAddr {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrorInvalidSigner,
			"not contract deployer or wrong nonce: expected %s instead of %s", derivedContractAddr.String(),
			msg.ContractAddress,
		)
	}

	// contract must already be deployed, to avoid spam registrations
	contractAccount := k.evmKeeper.GetAccountWithoutBalance(ctx, contract)
	if contractAccount == nil || !contractAccount.IsContract() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract has no code %s", msg.ContractAddress)
	}

	k.SetFee(ctx, contract, deployer, withdrawal)
	k.SetFeeInverse(ctx, deployer, contract)
	k.Logger(ctx).Debug(
		"registering contract for transaction fees",
		"contract", msg.ContractAddress, "deployer", msg.DeployerAddress,
		"withdraw", msg.WithdrawAddress,
	)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeRegisterDevFeeInfo,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawAddress, msg.WithdrawAddress),
			),
		},
	)

	return &types.MsgRegisterDevFeeInfoResponse{}, nil
}

// CancelDevFeeInfo deletes the fee for a contract
func (k Keeper) CancelDevFeeInfo(
	goCtx context.Context,
	msg *types.MsgCancelDevFeeInfo,
) (*types.MsgCancelDevFeeInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.isEnabled(ctx) {
		return nil, sdkerrors.Wrapf(types.ErrInternalFee, "fees module is not enabled")
	}

	deployerAddress, found := k.GetDeployer(ctx, common.HexToAddress(msg.ContractAddress))
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract %s is not registered", msg.ContractAddress)
	}

	if msg.DeployerAddress != deployerAddress.String() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the contract deployer", msg.DeployerAddress)
	}

	k.DeleteFee(ctx, common.HexToAddress(msg.ContractAddress))
	k.DeleteFeeInverse(ctx, deployerAddress)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeCancelDevFeeInfo,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
			),
		},
	)

	return &types.MsgCancelDevFeeInfoResponse{}, nil
}

// UpdateDevFeeInfo updates the withdraw address for a contract
func (k Keeper) UpdateDevFeeInfo(
	goCtx context.Context,
	msg *types.MsgUpdateDevFeeInfo,
) (*types.MsgUpdateDevFeeInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.isEnabled(ctx) {
		return nil, sdkerrors.Wrapf(types.ErrInternalFee, "fees module is not enabled")
	}

	contractAddress := common.HexToAddress(msg.ContractAddress)
	deployerAddress, found := k.GetDeployer(ctx, contractAddress)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract %s is not registered", msg.ContractAddress)
	}

	if msg.DeployerAddress != deployerAddress.String() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the contract deployer", msg.DeployerAddress)
	}

	withdrawalAddress, _ := sdk.AccAddressFromBech32(msg.WithdrawAddress)
	k.SetWithdrawal(ctx, contractAddress, withdrawalAddress)

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeUpdateDevFeeInfo,
				sdk.NewAttribute(types.AttributeKeyContract, msg.ContractAddress),
				sdk.NewAttribute(sdk.AttributeKeySender, msg.DeployerAddress),
				sdk.NewAttribute(types.AttributeKeyWithdrawAddress, msg.WithdrawAddress),
			),
		},
	)

	return &types.MsgUpdateDevFeeInfoResponse{}, nil
}
