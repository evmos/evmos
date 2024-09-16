// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package app

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/app/eips"
	evmconfig "github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// The init function of the config file allows to setup the global
// configuration for the EVM, modifying the custom ones defined in evmOS.
func init() {

	// fmt.Println(chainID)

	// if utils.IsMainnet(chainID) {
	// 	sdk.SetBaseDenom("aevmos")
	// } else if utils.IsTestnet(chainID) {
	// 	sdk.SetBaseDenom("atevmos")
	// } else {
	// 	panic("undefined chain denom")
	// }
	err := sdk.RegisterDenom("aevmos", math.LegacyNewDec(18))
	if err != nil {
		panic("cant register base denom")
	}
	err = sdk.SetBaseDenom("aevmos")

	if err != nil {
		panic("cant set base denom")
	}

	err = evmconfig.NewEVMConfigurator().
		WithExtendedEips(evmosActivators).
		// WithChainConfig(&ChainConfig).
		WithDecimals(evmconfig.SixDecimals).
		Configure()
	if err != nil {
		panic(err)
	}
}

// EvmosActivators defines a map of opcode modifiers associated
// with a key defining the corresponding EIP.
var evmosActivators = map[string]func(*vm.JumpTable){
	"evmos_0": eips.Enable0000,
	"evmos_1": eips.Enable0001,
	"evmos_2": eips.Enable0002,
}
