// Copyright Tharsis Labs Ltd.(Eidon-chain)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/Eidon-AI/eidon-chain/blob/main/LICENSE)

package eips

import (
	"github.com/Eidon-AI/eidon-chain/v20/x/evm/core/vm"
)

var (
	Multiplier        = uint64(10)
	SstoreConstantGas = uint64(500)
)

// enable0000 contains the logic to modify the CREATE and CREATE2 opcodes
// constant gas value.
func Enable0000(jt *vm.JumpTable) {
	currentValCreate := jt[vm.CREATE].GetConstantGas()
	jt[vm.CREATE].SetConstantGas(currentValCreate * Multiplier)

	currentValCreate2 := jt[vm.CREATE2].GetConstantGas()
	jt[vm.CREATE2].SetConstantGas(currentValCreate2 * Multiplier)
}

// enable0001 contains the logic to modify the CALL opcode
// constant gas value.
func Enable0001(jt *vm.JumpTable) {
	currentVal := jt[vm.CALL].GetConstantGas()
	jt[vm.CALL].SetConstantGas(currentVal * Multiplier)
}

// enable0002 contains the logic to modify the SSTORE opcode
// constant gas value.
func Enable0002(jt *vm.JumpTable) {
	jt[vm.SSTORE].SetConstantGas(SstoreConstantGas)
}
