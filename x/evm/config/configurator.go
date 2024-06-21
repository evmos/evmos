// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/evmos/evmos/v18/x/evm/core/vm"
	"github.com/evmos/evmos/v18/x/evm/types"
)

// EvmConfigurator allows to extend x/evm module configurations. The configurator modifies
// the EVM before starting the node, and all the validation are left to the InitGenesis of
// the module.
type EVMConfigurator struct {
	extended_eips         map[int]func(*vm.JumpTable)
	extended_default_eips []int64
}

func NewEVMConfigurator() *EVMConfigurator {
	return &EVMConfigurator{}
}

// WithExtendedEips allows to add to the go-ethereum activators map the provided
// EIP activators.
func (ec *EVMConfigurator) WithExtendedEips(extended_eips map[int]func(*vm.JumpTable)) *EVMConfigurator {
	ec.extended_eips = extended_eips
	return ec
}

// WithExtendedDefaultExtraEIPs update the x/evm DefaultExtraEIPs params
// by adding provided EIP numbers.
func (ec *EVMConfigurator) WithExtendedDefaultExtraEIPs(eips []int64) *EVMConfigurator {
	ec.extended_default_eips = eips
	return ec
}

// Apply apply the changes to the virtual machine configuration.
func (ec *EVMConfigurator) Apply() error {
	err := vm.ExtendActivators(ec.extended_eips)
	if err != nil {
		return err
	}

	for _, eip := range ec.extended_default_eips {
		if slices.Contains(types.DefaultExtraEIPs, eip) {
			return fmt.Errorf("duplicate default EIP: %d is already present in %v", eip, types.DefaultExtraEIPs)
		}
	}
	types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, ec.extended_default_eips...)
	return nil
}
