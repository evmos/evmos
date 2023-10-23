// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

const (
	// EventTypeLiquidStake is the event type emitted on a liquidStake transaction to Autopilot on Stride.
	EventTypeLiquidStake = "LiquidStake"
	// EventTypeRedeem is the event type emitted on a redeem transaction to Autopilot on Stride.
	EventTypeRedeem = "Redeem"
)

// EmitLiquidStakeEvent creates a new LiquidStake event on the EVM stateDB.
func (p Precompile) EmitLiquidStakeEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	sender, token common.Address,
	amount *big.Int,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeLiquidStake]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and token are indexed
	topics[1], err = cmn.MakeTopic(sender)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(token)
	if err != nil {
		return err
	}

	// Prepare the event data: amount
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

// EmitRedeemEvent creates a new Redeem event on the EVM stateDB.
func (p Precompile) EmitRedeemEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	sender,
	token common.Address,
	receiver string,
	amount *big.Int,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeRedeem]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and token are indexed
	topics[1], err = cmn.MakeTopic(sender)
	if err != nil {
		return err
	}
	topics[2], err = cmn.MakeTopic(token)
	if err != nil {
		return err
	}

	// Prepare the event data: receiver, amount
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3]}
	packed, err := arguments.Pack(receiver, amount)
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
