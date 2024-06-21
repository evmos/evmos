// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"golang.org/x/exp/slices"

	"github.com/evmos/evmos/v18/x/evm/core/vm"
	"github.com/evmos/evmos/v18/x/evm/types"
)

// ExtendEips allows to add to the go-ethereum activators map the provided
// EIP activators.
func ExtendEips(eips map[int]func(*vm.JumpTable)) {
	vm.ExtendActivators(eips)
}

// UpdateDefaultExtraEIPs update the x/evm DefaultExtraEIPs params
// by adding provided EIP numbers.
func UpdateDefaultExtraEIPs(eips []int64) {
	for _, eip := range eips {
		if !slices.Contains(types.DefaultExtraEIPs, eip) {
			types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, eip)
		}
	}
}
