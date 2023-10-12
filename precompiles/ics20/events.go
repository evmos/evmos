// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

const (
	// EventTypeIBCTransfer defines the event type for the ICS20 Transfer transaction.
	EventTypeIBCTransfer = "IBCTransfer"
	// EventTypeIBCTransferAuthorization defines the event type for the ICS20 TransferAuthorization transaction.
	EventTypeIBCTransferAuthorization = "IBCTransferAuthorization" //#nosec G101 -- no hardcoded credentials here
	// EventTypeRevokeIBCTransferAuthorization defines the event type for the ICS20 Authorization Revocation transaction.
	EventTypeRevokeIBCTransferAuthorization = "RevokeIBCTransferAuthorization"
)

// EmitIBCTransferAuthorizationEvent creates a new IBC transfer authorization event emitted on a TransferAuthorization transaction.
func (p Precompile) EmitIBCTransferAuthorizationEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	granteeAddr, granterAddr common.Address,
	allocations []cmn.ICS20Allocation,
) error {
	event := p.ABI.Events[EventTypeIBCTransferAuthorization]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(granteeAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(granterAddr)
	if err != nil {
		return err
	}

	// Prepare the event data: sourcePort, sourceChannel, denom, amount
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(allocations)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitIBCRevokeAuthorizationEvent creates a new IBC transfer authorization event emitted on a TransferAuthorization transaction.
func (p Precompile) EmitIBCRevokeAuthorizationEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	granteeAddr, granterAddr common.Address,
) error {
	event := p.ABI.Events[EventTypeRevokeIBCTransferAuthorization]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(granteeAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(granterAddr)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        nil,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitIBCTransferEvent creates a new IBC transfer event emitted on a Transfer transaction.
func (p Precompile) EmitIBCTransferEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	senderAddr common.Address,
	receiver string,
	sourcePort, sourceChannel string,
	token sdk.Coin,
	memo string,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeIBCTransfer]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and receiver are indexed
	topics[1], err = cmn.MakeTopic(senderAddr)
	if err != nil {
		return err
	}
	topics[2], err = cmn.MakeTopic(receiver)
	if err != nil {
		return err
	}

	// Prepare the event data: denom, amount, memo
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3], event.Inputs[4], event.Inputs[5], event.Inputs[6]}
	packed, err := arguments.Pack(sourcePort, sourceChannel, token.Denom, token.Amount.BigInt(), memo)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
