// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/x/evm/statedb"
)

const (
	// TransferMethod defines the ABI method name for the ICS20 Transfer
	// transaction.
	TransferMethod = "transfer"
)

// Transfer implements the ICS20 transfer transactions.
func (p Precompile) Transfer(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, sender, err := NewMsgTransfer(method, args)
	if err != nil {
		return nil, err
	}

	// check if channel exists and is open
	if !p.channelKeeper.HasChannel(ctx, msg.SourcePort, msg.SourceChannel) {
		return nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", msg.SourcePort, msg.SourceChannel)
	}

	// isCallerSender is true when the contract caller is the same as the sender
	isCallerSender := contract.CallerAddress == sender

	// If the contract caller is not the same as the sender, the sender must be the origin
	if !isCallerSender && origin != sender {
		return nil, fmt.Errorf(ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	resp, expiration, err := CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	res, err := p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err := UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, resp); err != nil {
		return nil, err
	}

	if err = EmitIBCTransferEvent(
		ctx,
		stateDB,
		p.ABI.Events[EventTypeIBCTransfer],
		p.Address(),
		sender,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		msg.Token,
		msg.Memo,
	); err != nil {
		return nil, err
	}

	// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB.
	// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
	if isCallerSender && msg.Token.Denom == p.stakingKeeper.BondDenom(ctx) {
		stateDB.(*statedb.StateDB).SubBalance(contract.CallerAddress, msg.Token.Amount.BigInt())
	}

	return method.Outputs.Pack(res.Sequence)
}
