// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/precompiles/authorization"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	vestingtypes "github.com/evmos/evmos/v18/x/vesting/types"
)

const (
	// EventTypeCreateClawbackVestingAccount defines the event type for the vesting
	// CreateClawbackVestingAccount transaction.
	EventTypeCreateClawbackVestingAccount = "CreateClawbackVestingAccount"
	// EventTypeFundVestingAccount defines the event type for the vesting
	// FundVestingAccount transaction.
	EventTypeFundVestingAccount = "FundVestingAccount"
	// EventTypeClawback defines the event type for the vesting Clawback transaction.
	EventTypeClawback = "Clawback"
	// EventTypeUpdateVestingFunder defines the event type for the vesting UpdateVestingFunder transaction.
	EventTypeUpdateVestingFunder = "UpdateVestingFunder"
	// EventTypeConvertVestingAccount defines the event type for the vesting ConvertVestingAccount transaction.
	EventTypeConvertVestingAccount = "ConvertVestingAccount"
)

// EmitApprovalEvent creates a new approval event emitted on an Approve, IncreaseAllowance and DecreaseAllowance transactions.
func (p Precompile) EmitApprovalEvent(ctx sdk.Context, stateDB vm.StateDB, grantee, granter common.Address, typeURL string) error {
	// Prepare the event topics
	event := p.ABI.Events[authorization.EventTypeApproval]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(grantee)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(granter)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(typeURL)
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

// EmitCreateClawbackVestingAccountEvent creates a new create clawback vesting account event emitted
// on a CreateClawbackVestingAccount transaction.
func (p Precompile) EmitCreateClawbackVestingAccountEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	funderAddr, vestingAddr common.Address,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeCreateClawbackVestingAccount]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(funderAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(vestingAddr)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	// Create the event
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        nil,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitFundVestingAccountEvent creates a new fund vesting account event emitted
// on a FundVestingAccount transaction.
func (p Precompile) EmitFundVestingAccountEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	msg *vestingtypes.MsgFundVestingAccount,
	funderAddr, vestingAddr common.Address,
	lockupPeriods *LockupPeriods,
	vestingPeriods *VestingPeriods,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeFundVestingAccount]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(funderAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(vestingAddr)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3], event.Inputs[4]}
	packed, err := arguments.Pack(uint64(msg.StartTime.Unix()), lockupPeriods.LockupPeriods, vestingPeriods.VestingPeriods)
	if err != nil {
		return err
	}

	// Create the event
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitClawbackEvent creates a new clawback event emitted on a Clawback transaction.
//
//nolint:dupl
func (p Precompile) EmitClawbackEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	funderAddr, accountAddr, destAddr common.Address,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeClawback]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(funderAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(accountAddr)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(destAddr)
	if err != nil {
		return err
	}

	// Create the event
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitUpdateVestingFunderEvent creates a new update vesting funder event emitted on a UpdateVestingFunder transaction.
//
//nolint:dupl
func (p Precompile) EmitUpdateVestingFunderEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	funderAddr, newFunderAddr, vestingAddr common.Address,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeUpdateVestingFunder]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(funderAddr)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(vestingAddr)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(newFunderAddr)
	if err != nil {
		return err
	}

	// Create the event
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitConvertVestingAccountEvent creates a new convert vesting account event emitted on a ConvertVestingAccount transaction.
func (p Precompile) EmitConvertVestingAccountEvent(ctx sdk.Context, stateDB vm.StateDB, vestingAddr common.Address) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeConvertVestingAccount]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(vestingAddr)
	if err != nil {
		return err
	}

	// Create the event
	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        nil,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
