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

package werc20

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	erc20 "github.com/evmos/evmos/v14/precompiles/erc20"
	erc20types "github.com/evmos/evmos/v14/x/erc20/types"
)

// abiPath defines the path to the staking precompile ABI JSON file.
const abiPath = "./abi.json"

var _ vm.PrecompiledContract = &Precompile{}

// Precompile defines the precompiled contract for staking.
type Precompile struct {
	*erc20.Precompile
}

// NewPrecompile creates a new staking Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	tokenPair erc20types.TokenPair,
	bankKeeper bankkeeper.Keeper,
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

	erc20Precompile, err := erc20.NewPrecompile(tokenPair, bankKeeper, authzKeeper)
	if err != nil {
		return nil, err
	}

	// use the IWERC20 ABI
	erc20Precompile.ABI = newAbi

	return &Precompile{
		Precompile: erc20Precompile,
	}, nil
}

// Address defines the address of the ERC20 precompile contract.
func (p Precompile) Address() common.Address {
	return p.Precompile.Address()
}

// RequiredGas calculates the contract gas use
func (Precompile) RequiredGas(_ []byte) uint64 {
	// TODO: gas should be the same ERC20
	return 0
}

// Run executes the precompiled contract staking methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.Precompile.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch method.Name {
	// WERC20 transactions
	case cmn.FallbackMethod, DepositMethod:
		bz, err = p.Deposit(ctx, contract, stateDB, method, args)
	case WithdrawMethod:
		bz, err = p.Withdraw(ctx, contract, stateDB, method, args)
		// ERC20 transactions and queries
	default:
		bz, err = p.Precompile.HandleMethod(ctx, contract, stateDB, method, args)
	}

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
func (p Precompile) IsTransaction(methodID string) bool {
	switch methodID {
	case cmn.FallbackMethod,
		DepositMethod,
		WithdrawMethod:
		return true
	default:
		return p.Precompile.IsTransaction(methodID)
	}
}
