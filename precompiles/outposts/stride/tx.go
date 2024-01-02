// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v16/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/precompiles/ics20"
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
	autopilotArgs, err := parseAutopilotArgs(method, args)
	if err != nil {
		return nil, err
	}

	sender := autopilotArgs.Sender
	receiver := autopilotArgs.Receiver
	token := autopilotArgs.Token
	amount := autopilotArgs.Amount
	strideForwarder := autopilotArgs.StrideForwarder

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err = ics20.CheckOriginAndSender(contract, origin, sender)
	if err != nil {
		return nil, err
	}

	// WEVMOS address is the only supported token for liquid staking
	if token != p.wevmosAddress {
		return nil, fmt.Errorf(ErrUnsupportedToken, token, p.wevmosAddress)
	}

	bondDenom := p.stakingKeeper.BondDenom(ctx)
	coin := sdk.Coin{Denom: bondDenom, Amount: math.NewIntFromBigInt(amount)}

	// Create the memo for the ICS20 transfer packet
	memo, err := CreateMemo(LiquidStakeAction, strideForwarder, sdk.AccAddress(receiver.Bytes()).String())
	if err != nil {
		return nil, err
	}

	// Build the MsgTransfer with the memo and coin
	msg, err := ics20.CreateAndValidateMsgTransfer(
		transfertypes.PortID,
		autopilotArgs.ChannelID,
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
	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
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

	return method.Outputs.Pack(true)
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
	autopilotArgs, err := parseAutopilotArgs(method, args)
	if err != nil {
		return nil, err
	}

	sender := autopilotArgs.Sender
	receiver := autopilotArgs.Receiver
	token := autopilotArgs.Token
	amount := autopilotArgs.Amount
	strideForwarder := autopilotArgs.StrideForwarder
	channelID := autopilotArgs.ChannelID

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

	ibcDenom := utils.ComputeIBCDenom(transfertypes.PortID, channelID, stToken)

	tokenPairID := p.erc20Keeper.GetDenomMap(ctx, ibcDenom)
	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	if !found {
		return nil, fmt.Errorf(ErrTokenPairNotFound, ibcDenom)
	}

	if token != tokenPair.GetERC20Contract() {
		return nil, fmt.Errorf(ErrUnsupportedToken, token, tokenPair.Erc20Address)
	}

	coin := sdk.Coin{Denom: tokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}

	// Create the memo for the ICS20 transfer
	memo, err := CreateMemo(RedeemStakeAction, strideForwarder, sdk.AccAddress(receiver.Bytes()).String())
	if err != nil {
		return nil, err
	}

	// Build the MsgTransfer with the memo and coin
	msg, err := ics20.CreateAndValidateMsgTransfer(
		transfertypes.PortID,
		channelID,
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
	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
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

	return method.Outputs.Pack(true)
}
