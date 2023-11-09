// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// Osmosis package contains the logic of the Osmosis outpost on the Evmos chain.
// This outpost uses the ics20 precompile to relay IBC packets to the Osmosis
// chain, targeting the Cross-Chain Swap Contract V2 (XCS V2).
package osmosis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/ics20"

	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// SwapMethod is the name of the swap method
	SwapMethod = "swap"
	// SwapAction is the action name needed in the memo field
	SwapAction = "Swap"
)

const (
	// NextMemo is the memo to use after the swap of the token in the IBC packet
	// built on the Osmosis chain. In the alpha version of the outpost this is
	// an empty string that will not be included in the XCS V2 contract payload.
	NextMemo = ""
)

// Swap is a transaction that swap tokens on the Osmosis chain using
// an ICS20 transfer with a custom memo field to trigger the XCS V2 contract.
func (p Precompile) Swap(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, input, output, amount, slippagePercentage, windowSeconds, swapReceiver, err := ParseSwapPacketData(args)
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

	inputDenom, err := p.erc20Keeper.GetTokenDenom(ctx, input)
	if err != nil {
		return nil, err
	}
	outputDenom, err := p.erc20Keeper.GetTokenDenom(ctx, output)
	if err != nil {
		return nil, err
	}

	// We need the bonded denom just for the outpost alpha version where the
	// the only two inputs allowed are aevmos and uosmo.
	bondDenom := p.stakingKeeper.GetParams(ctx).BondDenom

	err = ValidateInputOutput(inputDenom, outputDenom, bondDenom, p.portID, p.channelID)
	if err != nil {
		return nil, err
	}

	// If the receiver doesn't have the prefix "osmo", we should compute its address
	// in the Osmosis chain as a recovery address for the contract.
	onFailedDelivery := CreateOnFailedDeliveryField(sender.String())
	packet := CreatePacketWithMemo(
		outputDenom, swapReceiver, XCSContract, slippagePercentage, windowSeconds, onFailedDelivery, NextMemo,
	)

	err = packet.Memo.Validate()
	if err != nil {
		return nil, err
	}
	packetString := packet.String()

	coin := sdk.Coin{Denom: inputDenom, Amount: sdk.NewIntFromBigInt(amount)}
	msg, err := ics20.CreateAndValidateMsgTransfer(
		p.portID,
		p.channelID,
		coin,
		sdk.AccAddress(sender.Bytes()).String(),
		XCSContract,
		p.timeoutHeight,
		p.timeoutTimestamp,
		packetString,
	)
	if err != nil {
		return nil, err
	}

	// No need to have authorization when the contract caller is the same as
	// origin (owner of funds) and the sender is the origin
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
		packetString,
	); err != nil {
		return nil, err
	}

	// Emit the custom Swap Event
	if err := p.EmitSwapEvent(ctx, stateDB, sender, input, output, amount, swapReceiver); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Sequence, true)
}
