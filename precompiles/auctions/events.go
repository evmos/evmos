// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cmn "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
	"math/big"
)

const (
	// EventTypeBid defines the event type for the auctions Bid transaction.
	EventTypeBid = "Bid"
	// EventTypeDepositCoin defines the event type for the auctions DepositCoin transaction.
	EventTypeDepositCoin = "DepositCoin"
	// EventTypeRoundFinished defines the event type for the auctions RoundFinished event.
	EventTypeRoundFinished = "RoundFinished"
)

// EmitBidEvent creates a new event emitted on a Bid transaction.
func (p Precompile) EmitBidEvent(ctx sdk.Context, stateDB vm.StateDB, sender common.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.Events[EventTypeBid]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(sender)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
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

// EmitDepositCoinEvent creates a new event emitted on a DepositCoin transaction.
func (p Precompile) EmitDepositCoinEvent(ctx sdk.Context, stateDB vm.StateDB, sender common.Address, denom string, amount *big.Int) error {
	// Prepare the event topics
	event := p.Events[EventTypeDepositCoin]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(sender)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[1], event.Inputs[2]}
	packed, err := arguments.Pack(denom, amount)
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
