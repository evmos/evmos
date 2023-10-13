// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package p256

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/crypto/secp256r1"
)

var _ vm.PrecompiledContract = &Precompile{}

const (
	// P256VerifyGas is the secp256r1 elliptic curve signature verifier gas price
	P256VerifyGas uint64 = 3450
	// Required input length is 160 bytes
	p256VerifyInputLength = 160
)

// Precompile secp256r1 (P256) signature verification
// implemented as a native contract as per EIP-7212
type Precompile struct{}

// Address defines the address of the staking compile contract.
// address: 0x0000000000000000000000000000000000000800
func (Precompile) Address() common.Address {
	return common.BytesToAddress([]byte{19})
}

// RequiredGas returns the gas required to execute the precompiled contract
func (p Precompile) RequiredGas(input []byte) uint64 {
	return P256VerifyGas
}

// Run executes the precompiled contract with given 160 bytes of param, returning the output and the used gas
func (p *Precompile) Run(_ *vm.EVM, contract *vm.Contract, _ bool) (bz []byte, err error) {
	input := contract.Input
	// Check the input length
	if len(input) != p256VerifyInputLength {
		// Input length is invalid
		return nil, nil
	}

	// Extract the hash, r, s, x, y from the input
	hash := input[0:32]
	r, s := new(big.Int).SetBytes(input[32:64]), new(big.Int).SetBytes(input[64:96])
	x, y := new(big.Int).SetBytes(input[96:128]), new(big.Int).SetBytes(input[128:160])

	// Verify the secp256r1 signature

	if secp256r1.Verify(hash, r, s, x, y) {
		// Signature is valid
		return common.LeftPadBytes(common.Big1.Bytes(), 32), nil
	}

	// Signature is invalid
	return nil, nil
}
