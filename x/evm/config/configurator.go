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

// EvmConfigurator allows to extend x/evm module configurations. The configurator modifies
// the EVM before starting the node, and all the validation are left to the InitGenesis of
// the module.
type EVMConfigurator struct {
	extendedEIPs        map[int]func(*vm.JumpTable)
	extendedDefaultEIPs []int64
}

func NewEVMConfigurator() *EVMConfigurator {
	return &EVMConfigurator{}
}

// WithExtendedEips allows to add to the go-ethereum activators map the provided
// EIP activators.
func (ec *EVMConfigurator) WithExtendedEips(extendedEIPs map[int]func(*vm.JumpTable)) *EVMConfigurator {
	ec.extendedEIPs = extendedEIPs
	return ec
}

// WithExtendedDefaultExtraEIPs update the x/evm DefaultExtraEIPs params
// by adding provided EIP numbers.
func (ec *EVMConfigurator) WithExtendedDefaultExtraEIPs(eips []int64) *EVMConfigurator {
	ec.extendedDefaultEIPs = eips
	return ec
}

// Apply apply the changes to the virtual machine configuration.
func (ec *EVMConfigurator) Apply() error {
	err := vm.ExtendActivators(ec.extendedEIPs)
	if err != nil {
		return err
	}

	for _, eip := range ec.extendedDefaultEIPs {
		if !slices.Contains(types.DefaultExtraEIPs, eip) {
			types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, eip)
		}
	}
	return nil
}
