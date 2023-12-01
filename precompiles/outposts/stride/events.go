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
	cmn "github.com/evmos/evmos/v16/precompiles/common"
)

const (
	// EventTypeLiquidStake is the event type emitted on a liquidStake transaction to Autopilot on Stride.
	EventTypeLiquidStake = "LiquidStake"
	// EventTypeRedeemStake is the event type emitted on a redeem transaction to Autopilot on Stride.
	EventTypeRedeemStake = "RedeemStake"
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

// EmitRedeemStakeEvent creates a new RedeemStake event on the EVM stateDB.
func (p Precompile) EmitRedeemStakeEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	sender,
	token, receiver common.Address,
	strideForwarder string,
	amount *big.Int,
) error {
	// Prepare the event topics
	event := p.ABI.Events[EventTypeRedeemStake]
	topics := make([]common.Hash, 4)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and token are indexed
	topics[1], err = cmn.MakeTopic(sender)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(receiver)
	if err != nil {
		return err
	}

	topics[3], err = cmn.MakeTopic(token)
	if err != nil {
		return err
	}

	// Prepare the event data: receiver, amount
	arguments := abi.Arguments{event.Inputs[3], event.Inputs[4]}
	packed, err := arguments.Pack(strideForwarder, amount)
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
