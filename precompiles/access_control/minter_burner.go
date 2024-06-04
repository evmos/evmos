// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package accesscontrol

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v18/precompiles/common"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"

	"github.com/evmos/evmos/v18/x/access_control/types"
)

const (
	EventMint = "Mint"
	EventBurn = "Burn"

	MethodMint = "mint"
	MethodBurn = "burn"
)

// EmitEventPause creates a new Transfer event emitted on transfer and transferFrom transactions.
func (p Precompile) EmitEventMint(ctx sdk.Context, stateDB vm.StateDB, minter, to common.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventMint]
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(minter)
	if err != nil {
		return err
	}

	topics[2], err = cmn.MakeTopic(to)
	if err != nil {
		return err
	}

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

func (p Precompile) EmitEventBurn(ctx sdk.Context, stateDB vm.StateDB, burner common.Address, amount *big.Int) error {
	// Prepare the event topics
	event := p.ABI.Events[EventMint]
	topics := make([]common.Hash, 2)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	topics[1], err = cmn.MakeTopic(burner)
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

func (p Precompile) Mint(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	to, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid minter address")
	}

	if to == (common.Address{}) {
		return nil, fmt.Errorf("mint to the zero address")
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	if amount.Sign() != 1 {
		return nil, fmt.Errorf("mint amount not greater than 0")
	}

	if err := p.onlyRole(ctx, RoleMinter, contract.CallerAddress); err != nil {
		return nil, err
	}

	err := p.BankKeeper.MintCoins(ctx, erc20types.ModuleName, sdk.Coins{{Denom: p.TokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}})
	if err != nil {
		return nil, err
	}

	if err := p.EmitEventMint(ctx, stateDB, contract.CallerAddress, to, amount); err != nil {
		return nil, err
	}

	if err := p.Precompile.EmitTransferEvent(ctx, stateDB, common.Address{}, to, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

func (p Precompile) Burn(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}

	if amount.Sign() != 1 {
		return nil, fmt.Errorf("mint amount not greater than 0")
	}

	if err := p.onlyRole(ctx, RoleBurner, contract.CallerAddress); err != nil {
		return nil, err
	}

	balance := p.BankKeeper.GetBalance(ctx, contract.CallerAddress.Bytes(), p.TokenPair.Denom)
	if balance.Amount.BigInt().Cmp(amount) < 0 {
		return nil, fmt.Errorf("burn amount exceeds balance")
	}

	err := p.BankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{{Denom: p.TokenPair.Denom, Amount: math.NewIntFromBigInt(amount)}})
	if err != nil {
		return nil, err
	}

	if err := p.EmitEventBurn(ctx, stateDB, contract.CallerAddress, amount); err != nil {
		return nil, err
	}

	if err := p.Precompile.EmitTransferEvent(ctx, stateDB, contract.CallerAddress, common.Address{}, amount); err != nil {
		return nil, err
	}

	return nil, nil
}
