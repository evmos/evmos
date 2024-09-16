// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gov

import (
	"bytes"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

const (
	// EventTypeVote defines the event type for the gov VoteMethod transaction.
	EventTypeVote = "Vote"
)

// EmitVoteEvent creates a new event emitted on a Vote transaction.
func (p Precompile) EmitVoteEvent(ctx sdk.Context, stateDB vm.StateDB, voterAddress common.Address, proposalID uint64, option int32) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeVote]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID
	topics[1] = common.BytesToHash(voterAddress.Bytes())

	// Prepare the event data
	var b bytes.Buffer
	b.Write(cmn.PackNum(reflect.ValueOf(proposalID)))
	b.Write(cmn.PackNum(reflect.ValueOf(option)))

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        b.Bytes(),
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115
	})

	return nil
}
