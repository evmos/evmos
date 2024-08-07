// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"fmt"
<<<<<<< HEAD
	"slices"

	"github.com/evmos/evmos/v18/x/evm/core/vm"
	"github.com/evmos/evmos/v18/x/evm/types"
=======

	"golang.org/x/exp/slices"

	"github.com/evmos/evmos/v19/x/evm/core/vm"
	"github.com/evmos/evmos/v19/x/evm/types"
>>>>>>> main
)

// EVMConfigurator allows to extend x/evm module configurations. The configurator modifies
// the EVM before starting the node. This means that all init genesis validations will be
// applied to each change.
type EVMConfigurator struct {
<<<<<<< HEAD
	extendedEIPs             map[int]func(*vm.JumpTable)
	extendedDefaultExtraEIPs []int64
=======
	extendedEIPs             map[string]func(*vm.JumpTable)
	extendedDefaultExtraEIPs []string
>>>>>>> main
	sealed                   bool
}

// NewEVMConfigurator returns a pointer to a new EVMConfigurator object.
func NewEVMConfigurator() *EVMConfigurator {
	return &EVMConfigurator{}
}

// WithExtendedEips allows to add to the go-ethereum activators map the provided
// EIP activators.
<<<<<<< HEAD
func (ec *EVMConfigurator) WithExtendedEips(extendedEIPs map[int]func(*vm.JumpTable)) *EVMConfigurator {
=======
func (ec *EVMConfigurator) WithExtendedEips(extendedEIPs map[string]func(*vm.JumpTable)) *EVMConfigurator {
>>>>>>> main
	ec.extendedEIPs = extendedEIPs
	return ec
}

// WithExtendedDefaultExtraEIPs update the x/evm DefaultExtraEIPs params
// by adding provided EIP numbers.
<<<<<<< HEAD
func (ec *EVMConfigurator) WithExtendedDefaultExtraEIPs(eips ...int64) *EVMConfigurator {
=======
func (ec *EVMConfigurator) WithExtendedDefaultExtraEIPs(eips ...string) *EVMConfigurator {
>>>>>>> main
	ec.extendedDefaultExtraEIPs = eips
	return ec
}

// Configure apply the changes to the virtual machine configuration.
func (ec *EVMConfigurator) Configure() error {
	// If Configure method has been already used in the object, return
	// an error to avoid overriding configuration.
	if ec.sealed {
<<<<<<< HEAD
		return fmt.Errorf("EVMConfigurator has been sealed and cannot be modified")
=======
		return fmt.Errorf("error configuring EVMConfigurator: already sealed and cannot be modified")
>>>>>>> main
	}

	if err := vm.ExtendActivators(ec.extendedEIPs); err != nil {
		return err
	}

	for _, eip := range ec.extendedDefaultExtraEIPs {
		if slices.Contains(types.DefaultExtraEIPs, eip) {
<<<<<<< HEAD
			return fmt.Errorf("EIP %d is already present in the default list: %v", eip, types.DefaultExtraEIPs)
=======
			return fmt.Errorf("error configuring EVMConfigurator: EIP %s is already present in the default list: %v", eip, types.DefaultExtraEIPs)
		}

		if err := vm.ValidateEIPName(eip); err != nil {
			return fmt.Errorf("error configuring EVMConfigurator: %s", err)
>>>>>>> main
		}

		types.DefaultExtraEIPs = append(types.DefaultExtraEIPs, eip)
	}

	// After applying modifier the configurator is sealed. This way, it is not possible
	// to call the configure method twice.
	ec.sealed = true

	return nil
}
