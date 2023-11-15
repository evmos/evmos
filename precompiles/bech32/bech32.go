// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bech32

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

var _ vm.PrecompiledContract = &Precompile{}

const (
	// PrecompileAddress defines the address of the bech32 precompile contract.
	PrecompileAddress = "0x0000000000000000000000000000000000000400"
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for Bech32 encoding.
type Precompile struct {
	abi.ABI
	baseGas uint64
}

// NewPrecompile creates a new bech32 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(baseGas uint64) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, err
	}

	if baseGas == 0 {
		return nil, fmt.Errorf("baseGas cannot be zero")
	}

	return &Precompile{
		ABI:     newABI,
		baseGas: baseGas,
	}, nil
}

// Address defines the address of the bech32 compile contract.
// address: 0x0000000000000000000000000000000000000400
func (Precompile) Address() common.Address {
	return common.HexToAddress(PrecompileAddress)
}

// RequiredGas calculates the contract gas use.
func (p Precompile) RequiredGas(_ []byte) uint64 {
	return p.baseGas
}

// Run executes the precompiled contract bech32 methods defined in the ABI.
func (p Precompile) Run(_ *vm.EVM, contract *vm.Contract, _ bool) (bz []byte, err error) {
	methodID := contract.Input[:4]
	// NOTE: this function iterates over the method map and returns
	// the method with the given ID
	method, err := p.MethodById(methodID)
	if err != nil {
		return nil, err
	}

	argsBz := contract.Input[4:]
	args, err := method.Inputs.Unpack(argsBz)
	if err != nil {
		return nil, err
	}

	switch method.Name {
	case HexToBech32Method:
		bz, err = p.HexToBech32(method, args)
	case Bech32ToHexMethod:
		bz, err = p.Bech32ToHex(method, args)
	}

	if err != nil {
		return nil, err
	}

	return bz, nil
}
