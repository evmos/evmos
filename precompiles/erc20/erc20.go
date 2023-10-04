// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package erc20

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cmn "github.com/evmos/evmos/v14/precompiles/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	auth "github.com/evmos/evmos/v14/precompiles/authorization"
	erc20keeper "github.com/evmos/evmos/v14/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v14/x/erc20/types"
)

// abiPath defines the path to the staking precompile ABI JSON file.
const abiPath = "./abi.json"

var _ vm.PrecompiledContract = &Precompile{}

// Precompile defines the precompiled contract for staking.
type Precompile struct {
	cmn.Precompile
	abi.ABI
	tokenPair          erc20types.TokenPair
	bankKeeper         bankkeeper.Keeper
	erc20Keeper        erc20keeper.Keeper
	authzKeeper        authzkeeper.Keeper
	approvalExpiration time.Duration
}

// NewPrecompile creates a new staking Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	tokenPair erc20types.TokenPair,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	abiJSON, err := os.ReadFile(filepath.Clean(abiPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open newAbi.json file: %w", err)
	}

	newAbi, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return nil, fmt.Errorf("invalid newAbi.json file: %w", err)
	}

	return &Precompile{
		ABI:                newAbi,
		tokenPair:          tokenPair,
		bankKeeper:         bankKeeper,
		erc20Keeper:        erc20Keeper,
		authzKeeper:        authzKeeper,
		approvalExpiration: time.Hour * 24 * 365,
	}, nil
}

// Address defines the address of the ERC20 precompile contract.
func (p Precompile) Address() common.Address {
	return p.tokenPair.GetERC20Contract()
}

// IsStateful returns true since the precompile contract has access to the
// staking state.
func (Precompile) IsStateful() bool {
	return true
}

// RequiredGas calculates the contract gas use
func (Precompile) RequiredGas(_ []byte) uint64 {
	// TODO: gas should be the same ERC20
	return 0
}

// Run executes the precompiled contract staking methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	bz, err = p.HandleMethod(ctx, contract, stateDB, method, args)
	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	return bz, nil
}

// IsTransaction checks if the given methodID corresponds to a transaction or query.
func (Precompile) IsTransaction(methodID string) bool {
	switch methodID {
	case TransferMethod,
		TransferFromMethod,
		auth.ApproveMethod,
		auth.IncreaseAllowanceMethod,
		auth.DecreaseAllowanceMethod:
		return true
	default:
		return false
	}
}

func (p Precompile) HandleMethod(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) (bz []byte, err error) {
	switch method.Name {
	// ERC20 transactions
	case TransferMethod:
		bz, err = p.Transfer(ctx, contract, stateDB, method, args)
	case TransferFromMethod:
		bz, err = p.TransferFrom(ctx, contract, stateDB, method, args)
	case auth.ApproveMethod:
		bz, err = p.Approve(ctx, contract, stateDB, method, args)
	case auth.IncreaseAllowanceMethod:
		bz, err = p.IncreaseAllowance(ctx, contract, stateDB, method, args)
	case auth.DecreaseAllowanceMethod:
		bz, err = p.DecreaseAllowance(ctx, contract, stateDB, method, args)
	// ERC20 queries
	case NameMethod:
		bz, err = p.Name(ctx, contract, stateDB, method, args)
	case SymbolMethod:
		bz, err = p.Symbol(ctx, contract, stateDB, method, args)
	case DecimalsMethod:
		bz, err = p.Decimals(ctx, contract, stateDB, method, args)
	case TotalSupplyMethod:
		bz, err = p.TotalSupply(ctx, contract, stateDB, method, args)
	case BalanceOfMethod:
		bz, err = p.BalanceOf(ctx, contract, stateDB, method, args)
	case auth.AllowanceMethod:
		bz, err = p.Allowance(ctx, contract, stateDB, method, args)
	}

	return bz, err
}
