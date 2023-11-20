// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

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
	// EventTypeDeposit defines the event type for the Deposit transaction.
	EventTypeDeposit = "Deposit"
	// EventTypeWithdrawal defines the event type for the Withdraw transaction.
	EventTypeWithdrawal = "Withdrawal"
)

// EmitDepositEvent creates a new Deposit event emitted on a Deposit transaction.
func (p Precompile) EmitDepositEvent(ctx sdk.Context, stateDB vm.StateDB, dst common.Address, amount *big.Int) error {
	event := p.ABI.Events[EventTypeDeposit]
	return p.createWERC20Event(ctx, stateDB, event, dst, amount)
}

// EmitWithdrawalEvent creates a new Withdrawal event emitted on Withdraw transaction.
func (p Precompile) EmitWithdrawalEvent(ctx sdk.Context, stateDB vm.StateDB, src common.Address, amount *big.Int) error {
	event := p.ABI.Events[EventTypeWithdrawal]
	return p.createWERC20Event(ctx, stateDB, event, src, amount)
}

func (p Precompile) createWERC20Event(
	ctx sdk.Context,
	stateDB vm.StateDB,
	event abi.Event,
	address common.Address,
	amount *big.Int,
) error {
	// Prepare the event topics
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(address)
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
