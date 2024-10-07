// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convenient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

//go:build test
// +build test

package types

import (
	"errors"
	"fmt"
	"slices"

	geth "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// testChainConfig is the chain configuration used in the EVM to defined which
// opcodes are active based on Ethereum upgrades.
var testChainConfig *geth.ChainConfig

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

// WithExtendedEips allows you to add the provided EIP activators to the go-ethereum activators map.
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

// WithChainConfig allows you to define a custom `chainConfig` to be used in the EVM.
func (ec *EVMConfigurator) WithChainConfig(cc *ChainConfig) *EVMConfigurator {
	ec.chainConfig = cc
	return ec
}

// WithEVMCoinInfo allows you to define the denom and decimals of the token used as the EVM token.
func (ec *EVMConfigurator) WithEVMCoinInfo(denom string, d uint8) *EVMConfigurator {
	ec.evmDenom = EvmCoinInfo{Denom: denom, Decimals: Decimals(d)}
	return ec
}

// Configure applies the changes to the virtual machine configuration.
func (ec *EVMConfigurator) Configure() error {
	// If Configure method has been already used in the object, return
	// an error to avoid overriding configuration.
	if ec.sealed {
		return fmt.Errorf("error configuring EVMConfigurator: already sealed and cannot be modified")
	}

	if err := setTestChainConfig(ec.chainConfig); err != nil {
		return err
	}

	if ec.evmDenom.Denom != "" && ec.evmDenom.Decimals != 0 {
		setTestingEVMCoinInfo(ec.evmDenom)
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

	// After applying modifications, the configurator is sealed. This way, it is not possible
	// to call the configure method twice.
	ec.sealed = true

	return nil
}

func (ec *EVMConfigurator) ResetTestChainConfig() {
	vm.ResetActivators()
	resetEVMCoinInfo()
	testChainConfig = nil
}

func setTestChainConfig(cc *ChainConfig) error {
	if testChainConfig != nil {
		return errors.New("chainConfig already set. Cannot set again the chainConfig. Call the configurators ResetTestChainConfig method before configuring a new chain.")
	}
	config := DefaultChainConfig("")
	if cc != nil {
		config = cc
	}
	if err := config.Validate(); err != nil {
		return err
	}
	testChainConfig = config.EthereumConfig(nil)
	return nil
}

// GetChainConfig returns the `testChainConfig` used in the EVM.
func GetChainConfig() *geth.ChainConfig {
	return testChainConfig
}
