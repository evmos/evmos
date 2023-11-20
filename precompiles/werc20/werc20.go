// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

import (
	"embed"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	erc20 "github.com/evmos/evmos/v15/precompiles/erc20"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	transferkeeper "github.com/evmos/evmos/v15/x/ibc/transfer/keeper"
)

// abiPath defines the path to the WERC-20 precompile ABI JSON file.
const abiPath = "abi.json"

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

var _ vm.PrecompiledContract = &Precompile{}

// Precompile defines the precompiled contract for WERC20.
type Precompile struct {
	*erc20.Precompile
}

const (
	// DepositRequiredGas defines the gas required for the Deposit transaction.
	DepositRequiredGas uint64 = 28_799
	// WithdrawRequiredGas defines the gas required for the Withdraw transaction.
	WithdrawRequiredGas uint64 = 35_960
)

// NewPrecompile creates a new WERC20 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	tokenPair erc20types.TokenPair,
	bankKeeper bankkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, abiPath)
	if err != nil {
		return nil, err
	}

	erc20Precompile, err := erc20.NewPrecompile(tokenPair, bankKeeper, authzKeeper, transferKeeper)
	if err != nil {
		return nil, err
	}

	// use the IWERC20 ABI
	erc20Precompile.Precompile.ABI = newABI

	return &Precompile{
		Precompile: erc20Precompile,
	}, nil
}

// Address defines the address of the ERC20 precompile contract.
func (p Precompile) Address() common.Address {
	return p.Precompile.Address()
}

// RequiredGas calculates the contract gas use.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// TODO: these values were obtained from Remix using the WEVMOS9.sol.
	// We should execute the transactions from Evmos testnet
	// to ensure parity in the values.

	// If there is no method ID, then it's the fallback or receive case
	if len(input) == 0 {
		return DepositRequiredGas
	}

	methodID := input[:4]
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	switch method.Name {
	case DepositMethod, cmn.FallbackMethod, cmn.ReceiveMethod:
		return DepositRequiredGas
	case WithdrawMethod:
		return WithdrawRequiredGas
	}

	return p.Precompile.RequiredGas(input)
}

// Run executes the precompiled contract WERC20 methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, method, initialGas, args, err := p.Precompile.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch {
	case method == nil,
		method.Name == cmn.FallbackMethod,
		method.Name == cmn.ReceiveMethod,
		method.Name == DepositMethod:
		// WERC20 transactions
		bz, err = p.Deposit(ctx, contract, stateDB, method, args)
	case method.Name == WithdrawMethod:
		// Withdraw Method
		bz, err = p.Withdraw(ctx, contract, stateDB, method, args)
	default:
		// ERC20 transactions and queries
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
		cmn.ReceiveMethod,
		DepositMethod,
		WithdrawMethod:
		return true
	default:
		return p.Precompile.IsTransaction(methodID)
	}
}
