// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vesting

import (
	"embed"
	"fmt"

	"github.com/evmos/evmos/v20/precompiles/authorization"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	vestingkeeper "github.com/evmos/evmos/v20/x/vesting/keeper"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for staking.
type Precompile struct {
	cmn.Precompile
	vestingKeeper vestingkeeper.Keeper
}

// RequiredGas returns the required bare minimum gas to execute the precompile.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method.Name))
}

// NewPrecompile creates a new vesting Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	vestingKeeper vestingkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	newAbi, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("error loading the staking ABI %s", err)
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		vestingKeeper: vestingKeeper,
	}

	// SetAddress defines the address of the vesting precompiled contract.
	p.SetAddress(common.HexToAddress(evmtypes.VestingPrecompileAddress))

	return p, nil
}

// Run executes the precompiled contract staking methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch method.Name {
	// Approval transaction
	case authorization.ApproveMethod:
		bz, err = p.Approve(ctx, evm.Origin, stateDB, method, args)
	// Vesting transactions
	case CreateClawbackVestingAccountMethod:
		bz, err = p.CreateClawbackVestingAccount(ctx, evm.Origin, stateDB, method, args)
	case FundVestingAccountMethod:
		bz, err = p.FundVestingAccount(ctx, contract, evm.Origin, stateDB, method, args)
	case ClawbackMethod:
		bz, err = p.Clawback(ctx, contract, evm.Origin, stateDB, method, args)
	case UpdateVestingFunderMethod:
		bz, err = p.UpdateVestingFunder(ctx, contract, evm.Origin, stateDB, method, args)
	case ConvertVestingAccountMethod:
		bz, err = p.ConvertVestingAccount(ctx, stateDB, method, args)
	// Vesting queries
	case BalancesMethod:
		bz, err = p.Balances(ctx, method, args)
	}

	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}

	return bz, nil
}

// IsTransaction checks if the given method name corresponds to a transaction or query.
//
// Available vesting transactions are:
//   - CreateClawbackVestingAccount
//   - FundVestingAccount
//   - Clawback
//   - UpdateVestingFunder
//   - ConvertVestingAccount
//   - Approve
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case CreateClawbackVestingAccountMethod,
		FundVestingAccountMethod,
		ClawbackMethod,
		UpdateVestingFunderMethod,
		ConvertVestingAccountMethod,
		authorization.ApproveMethod:
		return true
	default:
		return false
	}
}

// Logger returns a precompile-specific logger.
func (p Precompile) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("evm extension", "vesting")
}
