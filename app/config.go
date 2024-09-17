// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package app

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/app/eips"
	"github.com/evmos/evmos/v20/utils"
	evmconfig "github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

var sealed = false

// The init function of the config file allows to setup the global
// configuration for the EVM, modifying the custom ones defined in evmOS.
// func init() {
func InitializeEVMConfiguration(chainID string) {

	if sealed {
		return
	}

	if chainID == "" {
		return
	}

	if utils.IsMainnet(chainID) {
		sdk.RegisterDenom("evmos", math.LegacyOneDec())
		if err := sdk.RegisterDenom("aevmos", math.LegacyNewDecWithPrec(1, 18)); err != nil {
			panic("cant register base denom")
		}
		if err := sdk.SetBaseDenom("aevmos"); err != nil {
			panic("cant set base denom")
		}

	} else if utils.IsTestnet(chainID) {
		if err := sdk.RegisterDenom("tevmos", math.LegacyOneDec()); err != nil {
			panic(err)
		}
		if err := sdk.RegisterDenom("atevmos", math.LegacyNewDecWithPrec(1, 18)); err != nil {
			panic(err)
		}
		if err := sdk.SetBaseDenom("atevmos"); err != nil {
			panic("cant set base denom")
		}
	} else {
		panic("undefined chain denom")
	}

	err := evmconfig.NewEVMConfigurator().
		WithExtendedEips(evmosActivators).
		// WithChainConfig(&ChainConfig).
		WithDenom("aevmos", evmconfig.EighteenDecimals).
		Configure()
	if err != nil {
		panic(err)
	}

	sealed = true
}

// EvmosActivators defines a map of opcode modifiers associated
// with a key defining the corresponding EIP.
var evmosActivators = map[string]func(*vm.JumpTable){
	"evmos_0": eips.Enable0000,
	"evmos_1": eips.Enable0001,
	"evmos_2": eips.Enable0002,
}
