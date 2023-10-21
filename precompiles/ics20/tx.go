// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
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

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided delegator address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err = CheckOriginAndSender(contract, origin, sender)
	if err != nil {
		return nil, err
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

	return method.Outputs.Pack(res.Sequence)
}
