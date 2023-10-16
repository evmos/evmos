// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"bytes"
	"embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/precompiles/authorization"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for staking.
type Precompile struct {
	cmn.Precompile
	stakingKeeper stakingkeeper.Keeper
}

// RequiredGas returns the required bare minimum gas to execute the precompile.
func (p Precompile) RequiredGas(input []byte) uint64 {
	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method.Name))
}

// NewPrecompile creates a new staking Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	stakingKeeper stakingkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	abiBz, err := f.ReadFile("abi.json")
	if err != nil {
		return nil, fmt.Errorf("error loading the staking ABI %s", err)
	}

	newAbi, err := abi.JSON(bytes.NewReader(abiBz))
	if err != nil {
		return nil, fmt.Errorf(cmn.ErrInvalidABI, err)
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		stakingKeeper: stakingKeeper,
	}, nil
}

// Address defines the address of the staking compile contract.
// address: 0x0000000000000000000000000000000000000800
func (Precompile) Address() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000800")
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

	if err := stateDB.Commit(); err != nil {
		return nil, err
	}

	switch method.Name {
	// Authorization transactions
	case authorization.ApproveMethod:
		bz, err = p.Approve(ctx, evm.Origin, stateDB, method, args)
	case authorization.RevokeMethod:
		bz, err = p.Revoke(ctx, evm.Origin, stateDB, method, args)
	case authorization.IncreaseAllowanceMethod:
		bz, err = p.IncreaseAllowance(ctx, evm.Origin, stateDB, method, args)
	case authorization.DecreaseAllowanceMethod:
		bz, err = p.DecreaseAllowance(ctx, evm.Origin, stateDB, method, args)
	// Staking transactions
	case DelegateMethod:
		bz, err = p.Delegate(ctx, evm.Origin, contract, stateDB, method, args)
	case UndelegateMethod:
		bz, err = p.Undelegate(ctx, evm.Origin, contract, stateDB, method, args)
	case RedelegateMethod:
		bz, err = p.Redelegate(ctx, evm.Origin, contract, stateDB, method, args)
	case CancelUnbondingDelegationMethod:
		bz, err = p.CancelUnbondingDelegation(ctx, evm.Origin, contract, stateDB, method, args)
	// Staking queries
	case DelegationMethod:
		bz, err = p.Delegation(ctx, contract, method, args)
	case UnbondingDelegationMethod:
		bz, err = p.UnbondingDelegation(ctx, contract, method, args)
	case ValidatorMethod:
		bz, err = p.Validator(ctx, method, contract, args)
	case ValidatorsMethod:
		bz, err = p.Validators(ctx, method, contract, args)
	case RedelegationMethod:
		bz, err = p.Redelegation(ctx, method, contract, args)
	case RedelegationsMethod:
		bz, err = p.Redelegations(ctx, method, contract, args)
	// Authorization queries
	case authorization.AllowanceMethod:
		bz, err = p.Allowance(ctx, method, contract, args)
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

// IsTransaction checks if the given method name corresponds to a transaction or query.
//
// Available staking transactions are:
//   - Delegate
//   - Undelegate
//   - Redelegate
//   - CancelUnbondingDelegation
//
// Available authorization transactions are:
//   - Approve
//   - Revoke
//   - IncreaseAllowance
//   - DecreaseAllowance
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case DelegateMethod,
		UndelegateMethod,
		RedelegateMethod,
		CancelUnbondingDelegationMethod,
		authorization.ApproveMethod,
		authorization.RevokeMethod,
		authorization.IncreaseAllowanceMethod,
		authorization.DecreaseAllowanceMethod:
		return true
	default:
		return false
	}
}

// Logger returns a precompile-specific logger.
func (p Precompile) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("evm extension", "staking")
}
