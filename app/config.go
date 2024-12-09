// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build !test
// +build !test

package app

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

// EvmosOptionsFn defines a function type for setting app options specifically for
// the Evmos app. The function should receive the chainID and return an error if
// any.
type EvmosOptionsFn func(string) error

// NoOpEvmosOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEvmosOptions(_ string) error {
	return nil
}

var sealed = false

// EvmosAppOptions allows to setup the global configuration
// for the Evmos chain.
func EvmosAppOptions(chainID string) error {
	if sealed {
		return nil
	}

	id := strings.Split(chainID, "-")[0]
	coinInfo, found := evmtypes.ChainsCoinInfo[id]
	if !found {
		return fmt.Errorf("unknown chain id: %s", chainID)
	}

	if err := setBaseDenom(coinInfo); err != nil {
		return err
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	err = evmtypes.NewEVMConfigurator().
		WithExtendedEips(evmosActivators).
		WithChainConfig(ethCfg).
		WithEVMCoinInfo(baseDenom, uint8(coinInfo.Decimals)).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain. The function registers different values based on
// the EvmCoinInfo to allow different configurations in mainnet and testnet.
func setBaseDenom(ci evmtypes.EvmCoinInfo) error {
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}
	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() are lower than the current base denom's units.
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
