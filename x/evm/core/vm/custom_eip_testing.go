// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build test
// +build test

// This file is used to allow the testing of EVM configuration initialization
// without the need to introduce testing requirements in the final binary. In
// this case, the file provides the possibility to restore the EIP activator
// functions to the initial state without the need to compile ResetActivators
// in the final binary.

package vm

var originalActivators = make(map[string]func(*JumpTable))

func init() {
	keys := GetActivatorsEipNames()

	originalActivators = make(map[string]func(*JumpTable), len(keys))

	for _, k := range keys {
		originalActivators[k] = activators[k]
	}
}

// ResetActivators resets activators to the original go ethereum activators map
func ResetActivators() {
	activators = make(map[string]func(*JumpTable))
	for k, v := range originalActivators {
		activators[k] = v
	}
}
