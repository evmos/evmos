// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convenient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

//go:build !test
// +build !test

package types

import (
	"fmt"
	"slices"

	geth "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// EVMConfigurator allows to extend x/evm module configurations. The configurator modifies
// the EVM before starting the node. This means that all init genesis validations will be
// applied to each change.
type EVMConfigurator struct {
	sealed                   bool
	extendedEIPs             map[string]func(*vm.JumpTable)
	extendedDefaultExtraEIPs []string
	chainConfig              *ChainConfig
	evmDenom                 EvmCoinInfo
}

// NewEVMConfigurator returns a pointer to a new EVMConfigurator object.
func NewEVMConfigurator() *EVMConfigurator {
	return &EVMConfigurator{}
}

// WithExtendedEips allows to add to the go-ethereum activators map the provided
// EIP activators.
func (ec *EVMConfigurator) WithExtendedEips(extendedEIPs map[string]func(*vm.JumpTable)) *EVMConfigurator {
	ec.extendedEIPs = extendedEIPs
	return ec
}

// WithExtendedDefaultExtraEIPs update the x/evm DefaultExtraEIPs params
// by adding provided EIP numbers.
func (ec *EVMConfigurator) WithExtendedDefaultExtraEIPs(eips ...string) *EVMConfigurator {
	ec.extendedDefaultExtraEIPs = eips
	return ec
}

// WithChainConfig allows to define a custom `chainConfig` to be used in the
// EVM.
func (ec *EVMConfigurator) WithChainConfig(cc *ChainConfig) *EVMConfigurator {
	ec.chainConfig = cc
	return ec
}

// WithEVMCoinInfo allows to define the denom and decimals of the token used as the
// EVM token.
func (ec *EVMConfigurator) WithEVMCoinInfo(denom string, d Decimals) *EVMConfigurator {
	ec.evmDenom = EvmCoinInfo{Denom: denom, Decimals: d}
	return ec
}

// Configure applies the changes to the virtual machine configuration.
func (ec *EVMConfigurator) Configure() error {
	// If Configure method has been already used in the object, return
	// an error to avoid overriding configuration.
	if ec.sealed {
		return fmt.Errorf("error configuring EVMConfigurator: already sealed and cannot be modified")
	}

	if err := setChainConfig(ec.chainConfig); err != nil {
		return err
	}

	if ec.evmDenom.Denom != "" && ec.evmDenom.Decimals != 0 {
		setEVMCoinInfo(ec.evmDenom)
	}

	if err := vm.ExtendActivators(ec.extendedEIPs); err != nil {
		return err
	}

	for _, eip := range ec.extendedDefaultExtraEIPs {
		if slices.Contains(DefaultExtraEIPs, eip) {
			return fmt.Errorf("error configuring EVMConfigurator: EIP %s is already present in the default list: %v", eip, DefaultExtraEIPs)
		}

		if err := vm.ValidateEIPName(eip); err != nil {
			return fmt.Errorf("error configuring EVMConfigurator: %s", err)
		}

		DefaultExtraEIPs = append(DefaultExtraEIPs, eip)
	}

	// After applying modifier the configurator is sealed. This way, it is not possible
	// to call the configure method twice.
	ec.sealed = true

	return nil
}

func (ec *EVMConfigurator) ResetTestChainConfig() {
	panic("this is only implemented with the 'test' build flag. Make sure you're running your tests using the '-tags=test' flag.")
}

// GetChainConfig returns the `chainConfig` used in the EVM.
func GetChainConfig() *geth.ChainConfig {
	return chainConfig
}
