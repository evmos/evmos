// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bank

import (
	"bytes"
	"embed"
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	erc20keeper "github.com/evmos/evmos/v15/x/erc20/keeper"
)

// PrecompileAddress defines the bank precompile address in Hex format
const PrecompileAddress = "0x0000000000000000000000000000000000000804"

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type Precompile struct {
	cmn.Precompile
	bankKeeper  bankkeeper.Keeper
	erc20Keeper erc20keeper.Keeper
}

// NewPrecompile creates a new bank Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) (*Precompile, error) {
	abiBz, err := f.ReadFile("abi.json")
	if err != nil {
		return nil, err
	}

	newAbi, err := abi.JSON(bytes.NewReader(abiBz))
	if err != nil {
		return nil, err
	}

	return &Precompile{
		Precompile:  cmn.Precompile{ABI: newAbi},
		bankKeeper:  bankKeeper,
		erc20Keeper: erc20Keeper,
	}, nil
}

// Address defines the address of the bank compile contract.
// address: 0x0000000000000000000000000000000000000804
func (Precompile) Address() common.Address {
	return common.HexToAddress(PrecompileAddress)
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	// NOTE: This charges the amount of gas required for a single ERC20 balanceOf or totalSupplt=y
	// If more than one item is returned in the query, the total gas will be: n*len(response)
	switch method.Name {
	case BalancesMethod:
		return 1 // TODO: define gas cost for ERC-20 balanceOf
	case TotalSupplyMethod:
		return 1 // TODO: define gas cost for ERC-20 totalSupply
	}

	return 0
}

// Run executes the precompiled contract bank query methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, _, method, initialGas, _, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch method.Name { //nolint: gocritic
	// Bank queries
	// case BalancesMethod:
	// 	bz, err = p.Balances(ctx, contract, method, args)
	// case TotalSupplyMethod:
	// 	bz, err = p.TotalSupply(ctx, contract, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	// if err != nil {
	// 	return nil, err
	// }

	// cost := ctx.GasMeter().GasConsumed() - initialGas

	// if !contract.UseGas(cost) {
	// 	return nil, vm.ErrOutOfGas
	// }

	// return bz, nil
}

// IsTransaction checks if the given method name corresponds to a transaction or query.
// It returns false since all bank methods are queries.
func (Precompile) IsTransaction(_ string) bool {
	return false
}
