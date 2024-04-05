// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v17/precompiles/common"
)

const (
	// EventTypeSwap is the event type emitted on a swap transaction to the
	// Osmosis chain.
	EventTypeSwap = "Swap"
)

// EmitSwapEvent creates a new SwapEvent event on the EVM stateDB.
func (p Precompile) EmitSwapEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	sender, input, output common.Address,
	amount *big.Int,
	receiver string,
) error {
	// Prepare the event topics.
	event := p.ABI.Events[EventTypeSwap]
	topics := make([]common.Hash, 4)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender, input, and output are indexed.
	topics[1], err = cmn.MakeTopic(sender)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(input)
	if err != nil {
		return err
	}

	topics[3], err = cmn.MakeTopic(output)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{event.Inputs[3], event.Inputs[4]}
	packed, err := arguments.Pack(amount, receiver)
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
