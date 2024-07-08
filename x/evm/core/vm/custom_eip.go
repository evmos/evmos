// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package vm

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
)

// ExtendActivators allows to merge the go ethereum activators map
// with additional activators.
func ExtendActivators(eips map[int]func(*JumpTable)) error {
	// Catch early duplicated eip.
	keys := make([]int, 0, len(eips))
	for k := range eips {
		if ExistsEipActivator(k) {
			return fmt.Errorf("duplicate activation: %d is already present in %s", k, ActivateableEips())
		}
		keys = append(keys, k)
	}

	// Sorting keys to ensure deterministic execution.
	sort.Ints(keys)

	for _, k := range keys {
		activators[k] = eips[k]
	}
	return nil
}

func GetActivatorsEipNumbers() []int {
	keys := maps.Keys(activators)

	sort.Ints(keys)
	return keys
}

// GetConstantGas return the constant gas used by the operation.
func (o *operation) GetConstantGas() uint64 {
	return o.constantGas
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
