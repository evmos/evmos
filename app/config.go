// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package app

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/app/eips"
	"github.com/evmos/evmos/v20/types"
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

	// on node init baseApp doesnt have the chainID set
	// we dont need to configure evm if the node is not running yet
	if chainID == "" {
		return
	}

	// set the base denom considering if its mainnet or testnet
	SetBaseDenomWithChainID(chainID)

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		panic("no base denom")
	}

	err = evmconfig.NewEVMConfigurator().
		WithExtendedEips(evmosActivators).
		// WithChainConfig(&ChainConfig).
		WithDenom(baseDenom, evmconfig.EighteenDecimals).
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

func SetBaseDenomWithChainID(chainID string) {
	if utils.IsTestnet(chainID) {
		if err := sdk.RegisterDenom(types.DisplayDenomTestnet, math.LegacyOneDec()); err != nil {
			panic(err)
		}
		if err := sdk.RegisterDenom(types.BaseDenomTestnet, math.LegacyNewDecWithPrec(1, types.BaseDenomUnit)); err != nil {
			panic(err)
		}
		if err := sdk.SetBaseDenom(types.BaseDenomTestnet); err != nil {
			panic("cant set base denom")
		}
		return
	}

	sdk.RegisterDenom(types.DisplayDenom, math.LegacyOneDec()) //nolint:errcheck
	if err := sdk.RegisterDenom(types.BaseDenom, math.LegacyNewDecWithPrec(1, types.BaseDenomUnit)); err != nil {
		panic("cant register base denom")
	}
	if err := sdk.SetBaseDenom(types.BaseDenom); err != nil {
		panic("cant set base denom")
	}
}
