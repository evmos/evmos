// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// NOTE: The AuctionEnd event is emitted when the auction ends which happens in the epoch hooks
// within the auctions module. The event is added manually to the logs and not emitted by the
// precompile directly as the auction end is not triggered by a transaction.
const (
	// EventTypeBid defines the event type for the auctions Bid transaction.
	EventTypeBid = "Bid"
	// EventTypeDepositCoin defines the event type for the auctions DepositCoin transaction.
	EventTypeDepositCoin = "CoinDeposit"
)

// EmitBidEvent creates a new event emitted on a Bid transaction.
func (p Precompile) EmitBidEvent(ctx sdk.Context, stateDB vm.StateDB, sender gethcommon.Address, round uint64, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeBid]
	topics := make([]gethcommon.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = common.MakeTopic(sender)
	if err != nil {
		return err
	}

	topics[2], err = common.MakeTopic(round)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2]}
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
func (p Precompile) EmitDepositCoinEvent(ctx sdk.Context, stateDB vm.StateDB, sender gethcommon.Address, round uint64, denom gethcommon.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeDepositCoin]
	topics := make([]gethcommon.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = common.MakeTopic(sender)
	if err != nil {
		return err
	}

	topics[2], err = common.MakeTopic(round)
	if err != nil {
		return err
	}

	// Pack the arguments to be used as the Data field
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3]}
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
