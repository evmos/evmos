// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	"fmt"

	"github.com/evmos/evmos/v15/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/precompiles/ics20"
)

const (
	// LiquidStakeMethod is the name of the liquidStake method
	LiquidStakeMethod = "liquidStake"
	// RedeemStakeMethod is the name of the redeem method
	RedeemStakeMethod = "redeemStake"
	// LiquidStakeAction is the action name needed in the memo field
	LiquidStakeAction = "LiquidStake"
	// RedeemStakeAction is the action name needed in the memo field
	RedeemStakeAction = "RedeemStake"
	// NoReceiver is the string used in the memo field when the receiver is not needed
	NoReceiver = ""
)

// LiquidStake is a transaction that liquid stakes tokens using
// a ICS20 transfer with a custom memo field that will trigger Stride's Autopilot middleware
func (p Precompile) LiquidStake(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, token, amount, receiver, err := parseLiquidStakeArgs(args)
	if err != nil {
		return nil, err
	}

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err = ics20.CheckOriginAndSender(contract, origin, sender)
	if err != nil {
		return nil, err
	}

	bondDenom := p.stakingKeeper.BondDenom(ctx)

	tokenPairID := p.erc20Keeper.GetDenomMap(ctx, bondDenom)

	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	// NOTE this should always exist
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, tokenPairID)
	}

	// NOTE: for v1 we only support the native EVM (and staking) denomination (WEVMOS/WTEVMOS).
	if token != tokenPair.GetERC20Contract() {
		return nil, fmt.Errorf(ErrUnsupportedToken, token, tokenPair.Erc20Address)
	}

	coin := sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}

	// Create the memo for the ICS20 transfer packet
	memo, err := CreateMemo(LiquidStakeAction, receiver, NoReceiver)
	if err != nil {
		return nil, err
	}

	// Build the MsgTransfer with the memo and coin
	msg, err := ics20.CreateAndValidateMsgTransfer(
		p.portID,
		p.channelID,
		coin,
		sdk.AccAddress(sender.Bytes()).String(),
		receiver,
		p.timeoutHeight,
		0,
		memo,
	)
	if err != nil {
		return nil, err
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	accept, expiration, err := ics20.CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	// Execute the ICS20 Transfer
	res, err := p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed
	if err := ics20.UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, accept); err != nil {
		return nil, err
	}

	// Emit the IBC transfer Event
	if err := ics20.EmitIBCTransferEvent(
		ctx,
		stateDB,
		p.ABI.Events[ics20.EventTypeIBCTransfer],
		p.Address(),
		sender,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		coin,
		memo,
	); err != nil {
		return nil, err
	}

	// Emit the custom LiquidStake Event
	if err := p.EmitLiquidStakeEvent(ctx, stateDB, sender, token, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Sequence, true)
}

// RedeemStake is a transaction that redeems the native tokens using the liquid stake
// tokens. It executes a ICS20 transfer with a custom memo field that will
// trigger Stride's Autopilot middleware
func (p Precompile) RedeemStake(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, receiver, token, strideForwarder, amount, err := parseRedeemStakeArgs(args)
	if err != nil {
		return nil, err
	}

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err = ics20.CheckOriginAndSender(contract, origin, sender)
	if err != nil {
		return nil, err
	}

	bondDenom := p.stakingKeeper.BondDenom(ctx)
	stToken := "st" + bondDenom

	ibcDenom := utils.ComputeIBCDenom(p.portID, p.channelID, stToken)

	tokenPairID := p.erc20Keeper.GetDenomMap(ctx, ibcDenom)
	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, ibcDenom)
	}

	if token != tokenPair.GetERC20Contract() {
		return nil, fmt.Errorf(ErrUnsupportedToken, token, tokenPair.Erc20Address)
	}

	coin := sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}

	// Create the memo for the ICS20 transfer
	memo, err := CreateMemo(RedeemStakeAction, strideForwarder, sdk.AccAddress(receiver.Bytes()).String())
	if err != nil {
		return nil, err
	}

	// Build the MsgTransfer with the memo and coin
	msg, err := ics20.CreateAndValidateMsgTransfer(
		p.portID,
		p.channelID,
		coin,
		sdk.AccAddress(sender.Bytes()).String(),
		strideForwarder,
		p.timeoutHeight,
		0,
		memo,
	)
	if err != nil {
		return nil, err
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	accept, expiration, err := ics20.CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	// Execute the ICS20 Transfer
	res, err := p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed
	if err := ics20.UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, accept); err != nil {
		return nil, err
	}

	// Emit the IBC transfer Event
	if err := ics20.EmitIBCTransferEvent(
		ctx,
		stateDB,
		p.ABI.Events[ics20.EventTypeIBCTransfer],
		p.Address(),
		sender,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		coin,
		memo,
	); err != nil {
		return nil, err
	}

	// Emit the custom RedeemStake Event
	if err := p.EmitRedeemStakeEvent(ctx, stateDB, sender, token, receiver, strideForwarder, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Sequence, true)
}
