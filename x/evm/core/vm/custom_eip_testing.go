// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build test
// +build test

package vm

var originalActivators = map[string]func(*JumpTable){
	"ethereum_3855": enable3855,
	"ethereum_3529": enable3529,
	"ethereum_3198": enable3198,
	"ethereum_2929": enable2929,
	"ethereum_2200": enable2200,
	"ethereum_1884": enable1884,
	"ethereum_1344": enable1344,
}

// ResetActivators resets activators to the original go ethereum activators map
func ResetActivators() {
	activators = make(map[string]func(*JumpTable))
	for k, v := range originalActivators {
		activators[k] = v
	}
}
