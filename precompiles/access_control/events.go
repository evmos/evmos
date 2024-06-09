// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	"math/big"
)

const (
	EventRoleAdminChanged = "RoleAdminChanged"
	EventRoleGranted      = "RoleGranted"
	EventRoleRevoked      = "RoleRevoked"

	EventMint = "Mint"
	EventBurn = "Burn"
)

// EmitEventMint creates a new Transfer event emitted on transfer and transferFrom transactions.
func (p Precompile) EmitEventMint(ctx sdk.Context, stateDB vm.StateDB, to common.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventMint]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(to)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[1]}
	packed, err := arguments.Pack(amount)
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

// EmitEventBurn creates a new Transfer event emitted on transfer and transferFrom transactions.
func (p Precompile) EmitEventBurn(ctx sdk.Context, stateDB vm.StateDB, burner common.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventBurn]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(burner)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[1]}
	packed, err := arguments.Pack(amount)
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

// EmitEventRoleAdminChanged emits an event when the admin of a role is changed.
func (p Precompile) EmitEventRoleAdminChanged(ctx sdk.Context, stateDB vm.StateDB, role common.Hash, previousAdmin, newAdmin common.Address) error {
	// Prepare the event topics
	events := p.ABI.Events[EventRoleAdminChanged]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = events.ID

	var err error
	topics[1], err = cmn.MakeTopic(role)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(previousAdmin)
	if err != nil {
		return err
	}

	topics[3], err = cmn.MakeTopic(newAdmin)
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

// EmitEventRoleGranted emits an event when a role is granted to an account.
func (p Precompile) EmitEventRoleGranted(ctx sdk.Context, stateDB vm.StateDB, role common.Hash, account, sender common.Address) error {
	// Prepare the event topics
	events := p.ABI.Events[EventRoleGranted]
	topics := make([]common.Hash, 4)

	// The first topic is always the signature of the event.
	topics[0] = events.ID

	var err error
	topics[1], err = cmn.MakeTopic(role)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(account)
	if err != nil {
		return err
	}

	topics[3], err = cmn.MakeTopic(sender)
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

// EmitEventRoleRevoked emits an event when a role is revoked from an account.
func (p Precompile) EmitEventRoleRevoked(ctx sdk.Context, stateDB vm.StateDB, role common.Hash, account, sender common.Address) error {
	// Prepare the event topics
	events := p.ABI.Events[EventRoleRevoked]
	topics := make([]common.Hash, 4)

	// The first topic is always the signature of the event.
	topics[0] = events.ID

	var err error
	topics[1], err = cmn.MakeTopic(role)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(account)
	if err != nil {
		return err
	}

	topics[3], err = cmn.MakeTopic(sender)
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
