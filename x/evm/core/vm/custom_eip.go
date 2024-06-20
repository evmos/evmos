// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"sort"
)

// ExtendActivators allows to merge the go ethereum activators map
// with additional activators.
func ExtendActivators(eips map[int]func(*JumpTable)) {
	// Sorting key to ensure deterministic execution.
	keys := make([]int, 0, len(eips))
	for k := range eips {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// If an EIP is already present in the activators type, skip it.
	for _, k := range keys {
		if _, exist := activators[k]; !exist {
			activators[k] = eips[k]
		}
	}
}

// SetExecute sets the execution function of the operation.
func (o *operation) SetExecute(ef executionFunc) {
	o.execute = ef
}

// SetConstantGas changes the constant gas of the operation.
func (o *operation) SetConstantGas(gas uint64) {
	o.constantGas = gas
}

// SetDynamicGas sets the dynamic gas function of the operation.
func (o *operation) SetDynamicGas(gf gasFunc) {
	o.dynamicGas = gf
}

// SetMinStack sets the minimum stack size required for the operation.
func (o *operation) SetMinStack(minStack int) {
	o.minStack = minStack
}

// SetMaxStack sets the maximum stack size for the operation.
func (o *operation) SetMaxStack(maxStack int) {
	o.maxStack = maxStack
}

// SetMemorySize sets the memory size function for the operation.
func (o *operation) SetMemorySize(msf memorySizeFunc) {
	o.memorySize = msf
}
