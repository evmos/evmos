// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package common

import (
	"fmt"
	"math/big"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/x/evm/statedb"
)

// Precompile is a common struct for all precompiles that holds the common data each
// precompile needs to run which includes the ABI, Gas config, approval expiration and the authz keeper.
type Precompile struct {
	abi.ABI
	AuthzKeeper          authzkeeper.Keeper
	ApprovalExpiration   time.Duration
	KvGasConfig          storetypes.GasConfig
	TransientKVGasConfig storetypes.GasConfig
	address              common.Address
	journalEntries       []balanceChangeEntry
}

// Operation is a type that defines if the precompile call
// produced an addition or substraction of an account's balance
type Operation int8

const (
	Sub Operation = iota
	Add
)

type balanceChangeEntry struct {
	Account common.Address
	Amount  *big.Int
	Op      Operation
}

func NewBalanceChangeEntry(acc common.Address, amt *big.Int, op Operation) balanceChangeEntry {
	return balanceChangeEntry{acc, amt, op}
}

// snapshot contains all state and events previous to the precompile call
// This is needed to allow us to revert the changes
// during the EVM execution
type snapshot struct {
	MultiStore sdk.CacheMultiStore
	Events     sdk.Events
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
) (ctx sdk.Context, stateDB *statedb.StateDB, s snapshot, method *abi.Method, gasConfig storetypes.Gas, args []interface{}, err error) {
	stateDB, ok := evm.StateDB.(*statedb.StateDB)
	if !ok {
		return sdk.Context{}, nil, s, nil, uint64(0), nil, fmt.Errorf(ErrNotRunInEvm)
	}

	// get the stateDB cache ctx
	ctx, err = stateDB.GetCacheContext()
	if err != nil {
		return sdk.Context{}, nil, s, nil, uint64(0), nil, err
	}

	// take a snapshot of the current state before any changes
	// to be able to revert the changes
	s.MultiStore = stateDB.MultiStoreSnapshot()
	s.Events = ctx.EventManager().Events()

	// commit the current changes in the cache ctx
	// to get the updated state for the precompile call
	if err := stateDB.CommitWithCacheCtx(); err != nil {
		return sdk.Context{}, nil, s, nil, uint64(0), nil, err
	}

	// NOTE: This is a special case where the calling transaction does not specify a function name.
	// In this case we default to a `fallback` or `receive` function on the contract.

	// Simplify the calldata checks
	isEmptyCallData := len(contract.Input) == 0
	isShortCallData := len(contract.Input) > 0 && len(contract.Input) < 4
	isStandardCallData := len(contract.Input) >= 4

	switch {
	// Case 1: Calldata is empty
	case isEmptyCallData:
		method, err = p.emptyCallData(contract)

	// Case 2: calldata is non-empty but less than 4 bytes needed for a method
	case isShortCallData:
		method, err = p.methodIDCallData()

	// Case 3: calldata is non-empty and contains the minimum 4 bytes needed for a method
	case isStandardCallData:
		method, err = p.standardCallData(contract)
	}

	if err != nil {
		return sdk.Context{}, nil, s, nil, uint64(0), nil, err
	}

	// return error if trying to write to state during a read-only call
	if readOnly && isTransaction(method.Name) {
		return sdk.Context{}, nil, s, nil, uint64(0), nil, vm.ErrWriteProtection
	}

	// if the method type is `function` continue looking for arguments
	if method.Type == abi.Function {
		argsBz := contract.Input[4:]
		args, err = method.Inputs.Unpack(argsBz)
		if err != nil {
			return sdk.Context{}, nil, s, nil, uint64(0), nil, err
		}
	}

	initialGas := ctx.GasMeter().GasConsumed()

	defer HandleGasError(ctx, contract, initialGas, &err)()

	// set the default SDK gas configuration to track gas usage
	// we are changing the gas meter type, so it panics gracefully when out of gas
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(contract.Gas)).
		WithKVGasConfig(p.KvGasConfig).
		WithTransientKVGasConfig(p.TransientKVGasConfig)
	// we need to consume the gas that was already used by the EVM
	ctx.GasMeter().ConsumeGas(initialGas, "creating a new gas meter")

	return ctx, stateDB, s, method, initialGas, args, nil
}

// HandleGasError handles the out of gas panic by resetting the gas meter and returning an error.
// This is used in order to avoid panics and to allow for the EVM to continue cleanup if the tx or query run out of gas.
func HandleGasError(ctx sdk.Context, contract *vm.Contract, initialGas storetypes.Gas, err *error) func() {
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

// AddJournalEntries adds the balanceChange (if corresponds)
// and precompileCall entries on the stateDB journal
// This allows to revert the call changes within an evm tx
func (p Precompile) AddJournalEntries(stateDB *statedb.StateDB, s snapshot) error {
	for _, entry := range p.journalEntries {
		switch entry.Op {
		case Sub:
			// add the corresponding balance change to the journal
			stateDB.SubBalance(entry.Account, entry.Amount)
		case Add:
			// add the corresponding balance change to the journal
			stateDB.AddBalance(entry.Account, entry.Amount)
		}
	}

	if err := stateDB.AddPrecompileFn(p.Address(), s.MultiStore, s.Events); err != nil {
		return err
	}
	return nil
}

// SetBalanceChangeEntries sets the balanceChange entries
// as the journalEntries field of the precompile.
// These entries will be added to the stateDB's journal
// when calling the AddJournalEntries function
func (p *Precompile) SetBalanceChangeEntries(entries ...balanceChangeEntry) {
	p.journalEntries = entries
}

func (p Precompile) Address() common.Address {
	return p.address
}

func (p *Precompile) SetAddress(addr common.Address) {
	p.address = addr
}

// emptyCallData is a helper function that returns the method to be called when the calldata is empty.
func (p Precompile) emptyCallData(contract *vm.Contract) (method *abi.Method, err error) {
	switch {
	// Case 1.1: Send call or transfer tx - 'receive' is called if present and value is transferred
	case contract.Value().Sign() > 0 && p.HasReceive():
		return &p.Receive, nil
	// Case 1.2: Either 'receive' is not present, or no value is transferred - call 'fallback' if present
	case p.HasFallback():
		return &p.Fallback, nil
	// Case 1.3: Neither 'receive' nor 'fallback' are present - return error
	default:
		return nil, vm.ErrExecutionReverted
	}
}

// methodIDCallData is a helper function that returns the method to be called when the calldata is less than 4 bytes.
func (p Precompile) methodIDCallData() (method *abi.Method, err error) {
	// Case 2.2: calldata contains less than 4 bytes needed for a method and 'fallback' is not present - return error
	if !p.HasFallback() {
		return nil, vm.ErrExecutionReverted
	}
	// Case 2.1: calldata contains less than 4 bytes needed for a method - 'fallback' is called if present
	return &p.Fallback, nil
}

// standardCallData is a helper function that returns the method to be called when the calldata is 4 bytes or more.
func (p Precompile) standardCallData(contract *vm.Contract) (method *abi.Method, err error) {
	methodID := contract.Input[:4]
	// NOTE: this function iterates over the method map and returns
	// the method with the given ID
	method, err = p.MethodById(methodID)

	// Case 3.1 calldata contains a non-existing method ID, and `fallback` is not present - return error
	if err != nil && !p.HasFallback() {
		return nil, err
	}

	// Case 3.2: calldata contains a non-existing method ID - 'fallback' is called if present
	if err != nil && p.HasFallback() {
		return &p.Fallback, nil
	}

	return method, nil
}
