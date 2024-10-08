// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// EventTypeVote defines the event type for the gov VoteMethod transaction.
	EventTypeVote = "Vote"
	// EventTypeVoteWeighted defines the event type for the gov VoteWeightedMethod transaction.
	EventTypeVoteWeighted = "VoteWeighted"
)

// EmitVoteEvent creates a new event emitted on a Vote transaction.
func (p Precompile) EmitVoteEvent(ctx sdk.Context, stateDB vm.StateDB, voterAddress common.Address, proposalID uint64, option int32) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeVote]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(voterAddress)
	if err != nil {
		return err
	}

	// Prepare the event data
	arguments := abi.Arguments{event.Inputs[1], event.Inputs[2]}
	packed, err := arguments.Pack(proposalID, uint8(option)) //nolint:gosec // G115
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115
	})

	return nil
}

// EmitVoteWeightedEvent creates a new event emitted on a VoteWeighted transaction.
func (p Precompile) EmitVoteWeightedEvent(ctx sdk.Context, stateDB vm.StateDB, voterAddress common.Address, proposalID uint64, options WeightedVoteOptions) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeVoteWeighted]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(voterAddress)
	if err != nil {
		return err
	}

	// Prepare the event data
	arguments := abi.Arguments{event.Inputs[1], event.Inputs[2]}
	packed, err := arguments.Pack(proposalID, options)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115
	})

	return nil
}
