// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
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
	// EventTypeUpdateParams defines the event type for the gov UpdateParams transaction.
	EventTypeUpdateParams = "UpdateParams"
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

// EmitUpdateParamsEvent emits an event after updating governance parameters
func (p Precompile) EmitUpdateParamsEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	params v1.Params,
) error {
	// Create the event signature
	event := p.ABI.Events[EventTypeUpdateParams]
	topics := make([]common.Hash, 1)

	// Create the event topics
	topics[0] = event.ID

	outputParams := ParamsOutput{
		VotingPeriod:               params.VotingPeriod.Nanoseconds(),
		MinDeposit:                 cmn.NewCoinsResponse(params.MinDeposit),
		MaxDepositPeriod:           params.MaxDepositPeriod.Nanoseconds(),
		Quorum:                     params.Quorum,
		Threshold:                  params.Threshold,
		VetoThreshold:              params.VetoThreshold,
		MinInitialDepositRatio:     params.MinInitialDepositRatio,
		ProposalCancelRatio:        params.ProposalCancelRatio,
		ProposalCancelDest:         params.ProposalCancelDest,
		ExpeditedVotingPeriod:      params.ExpeditedVotingPeriod.Nanoseconds(),
		ExpeditedThreshold:         params.ExpeditedThreshold,
		ExpeditedMinDeposit:        cmn.NewCoinsResponse(params.ExpeditedMinDeposit),
		BurnVoteQuorum:             params.BurnVoteQuorum,
		BurnProposalDepositPrevote: params.BurnProposalDepositPrevote,
		BurnVoteVeto:               params.BurnVoteVeto,
		MinDepositRatio:            params.MinDepositRatio,
	}
	arguments := abi.Arguments{event.Inputs[0]}
	packed, err := arguments.Pack(outputParams)
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
