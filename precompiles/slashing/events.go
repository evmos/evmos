// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// EventTypeValidatorUnjailed defines the event type for validator unjailing
	EventTypeValidatorUnjailed = "ValidatorUnjailed"
)

// Add this struct after the existing constants
type EventValidatorUnjailed struct {
	Validator common.Address
}

// EmitValidatorUnjailedEvent emits the ValidatorUnjailed event
func (p Precompile) EmitValidatorUnjailedEvent(ctx sdk.Context, stateDB vm.StateDB, validator common.Address) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeValidatorUnjailed]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(validator)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115
	})

	return nil
}
