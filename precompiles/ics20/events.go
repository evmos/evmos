// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

const (
	// EventTypeIBCTransfer defines the event type for the ICS20 Transfer transaction.
	EventTypeIBCTransfer = "IBCTransfer"
)

// EmitIBCTransferEvent creates a new IBC transfer event emitted on a Transfer transaction.
func EmitIBCTransferEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	events map[string]abi.Event,
	senderAddr, precompileAddr common.Address,
	msg *transfertypes.MsgTransfer,
) error {
	// Prepare the event topics
	event := events[EventTypeIBCTransfer]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and receiver are indexed
	topics[1], err = cmn.MakeTopic(senderAddr)
	if err != nil {
		return err
	}
	// TODO: This should be an address ?
	topics[2], err = cmn.MakeTopic(msg.Receiver)
	if err != nil {
		return err
	}

	// Prepare the event data: denom, amount, memo
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3], event.Inputs[4], event.Inputs[5], event.Inputs[6]}
	packed, err := arguments.Pack(msg.SourcePort, msg.SourceChannel, msg.Token.Denom, msg.Token.Amount.BigInt(), msg.Memo)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     precompileAddr,
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
