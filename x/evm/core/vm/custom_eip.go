// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

// ExtendActivators allows to merge the go ethereum activators map
// with additional activators.
func ExtendActivators(eips map[int]func(*JumpTable)) {
	for k, v := range eips {
		activators[k] = v
	}
}

// SetConstantGas change the constant gas of the operation.
func (o *operation) SetConstantGas(gas uint64) {
	o.constantGas = gas
}
