// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package common

import (
	"fmt"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v13/x/evm/statedb"
)

// Precompile is a common struct for all precompiles that holds the common data each
// precompile needs to run which includes the ABI, Gas config, approval expiration and the authz keeper.
type Precompile struct {
	abi.ABI
	AuthzKeeper          authzkeeper.Keeper
	ApprovalExpiration   time.Duration
	KvGasConfig          storetypes.GasConfig
	TransientKVGasConfig storetypes.GasConfig
}

// RequiredGas calculates the base minimum required gas for a transaction or a query.
// It uses the method ID to determine if the input is a transaction or a query and
// uses the Cosmos SDK gas config flat cost and the flat per byte cost * len(argBz) to calculate the gas.
func (p Precompile) RequiredGas(input []byte, isTransaction bool) uint64 {
	argsBz := input[4:]

	if isTransaction {
		return p.KvGasConfig.WriteCostFlat + (p.KvGasConfig.WriteCostPerByte * uint64(len(argsBz)))
	}

	return p.KvGasConfig.ReadCostFlat + (p.KvGasConfig.ReadCostPerByte * uint64(len(argsBz)))
}

// RunSetup runs the initial setup required to run a transaction or a query.
// It returns the sdk Context, EVM stateDB, ABI method, initial gas and calling arguments.
func (p Precompile) RunSetup(
	evm *vm.EVM,
	contract *vm.Contract,
	readOnly bool,
	isTransaction func(name string) bool,
) (ctx sdk.Context, stateDB *statedb.StateDB, method *abi.Method, gasConfig sdk.Gas, args []interface{}, err error) {
	stateDB, ok := evm.StateDB.(*statedb.StateDB)
	if !ok {
		return sdk.Context{}, nil, nil, uint64(0), nil, fmt.Errorf(ErrNotRunInEvm)
	}
	ctx = stateDB.GetContext()

	methodID := contract.Input[:4]
	// NOTE: this function iterates over the method map and returns
	// the method with the given ID
	method, err = p.MethodById(methodID)
	if err != nil {
		return sdk.Context{}, nil, nil, uint64(0), nil, err
	}

	// return error if trying to write to state during a read-only call
	if readOnly && isTransaction(method.Name) {
		return sdk.Context{}, nil, nil, uint64(0), nil, vm.ErrWriteProtection
	}

	argsBz := contract.Input[4:]
	args, err = method.Inputs.Unpack(argsBz)
	if err != nil {
		return sdk.Context{}, nil, nil, uint64(0), nil, err
	}

	initialGas := ctx.GasMeter().GasConsumed()

	defer HandleGasError(ctx, contract, initialGas, &err)()

	// set the default SDK gas configuration to track gas usage
	// we are changing the gas meter type, so it panics gracefully when out of gas
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(contract.Gas)).
		WithKVGasConfig(p.KvGasConfig).
		WithTransientKVGasConfig(p.TransientKVGasConfig)
	// we need to consume the gas that was already used by the EVM
	ctx.GasMeter().ConsumeGas(initialGas, "creating a new gas meter")

	return ctx, stateDB, method, initialGas, args, nil
}

// HandleGasError handles the out of gas panic by resetting the gas meter and returning an error.
// This is used in order to avoid panics and to allow for the EVM to continue cleanup if the tx or query run out of gas.
func HandleGasError(ctx sdk.Context, contract *vm.Contract, initialGas sdk.Gas, err *error) func() {
	return func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case sdk.ErrorOutOfGas:
				// update contract gas
				usedGas := ctx.GasMeter().GasConsumed() - initialGas
				_ = contract.UseGas(usedGas)

				*err = vm.ErrOutOfGas
				// FIXME: add InfiniteGasMeter with previous Gas limit.
				ctx = ctx.WithKVGasConfig(storetypes.GasConfig{}).
					WithTransientKVGasConfig(storetypes.GasConfig{})
			default:
				panic(r)
			}
		}
	}
}
