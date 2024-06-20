// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"github.com/evmos/evmos/v18/x/evm/core/vm"
	"github.com/evmos/evmos/v18/x/evm/types"
)

func ExtendEips(eips map[int]func(*vm.JumpTable)) {
	vm.ExtendActivators(eips)
}

func UpdateDefaultExtraEIPs(eips []int64) {
	types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, eips...)
}
