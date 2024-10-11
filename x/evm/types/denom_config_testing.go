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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/types"
)

var (
	// testingEvmCoinInfo hold the information of the coin used in the EVM as gas token. It
	// can only be set via `EVMConfigurator` before starting the app.
	testingEvmCoinInfo *EvmCoinInfo
	// defaultCoinInfo is the default coin info used
	// when the coin info is not specified
	defaultCoinInfo = EvmCoinInfo{
		Denom:        types.BaseDenom,
		DisplayDenom: types.DisplayDenom,
		Decimals:     EighteenDecimals,
	}
)

// setEVMCoinDecimals allows to define the decimals used in the representation
// of the EVM coin.
func setEVMCoinDecimals(d Decimals) error {
	if err := d.Validate(); err != nil {
		return fmt.Errorf("setting EVM coin decimals: %w", err)
	}

	testingEvmCoinInfo.Decimals = d
	return nil
}

// setEVMCoinDenom allows to define the denom of the coin used in the EVM.
func setEVMCoinDenom(denom string) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}
	testingEvmCoinInfo.Denom = denom
	return nil
}

// GetEVMCoinDecimals returns the decimals used in the representation of the EVM
// coin.
func GetEVMCoinDecimals() Decimals {
	return testingEvmCoinInfo.Decimals
}

// GetEVMCoinDenom returns the denom used for the EVM coin.
func GetEVMCoinDenom() string {
	return testingEvmCoinInfo.Denom
}

// SetEVMCoinInfo allows to define denom and decimals of the coin used in the EVM.
func setTestingEVMCoinInfo(eci EvmCoinInfo) error {
	if testingEvmCoinInfo != nil {
		return errors.New("testing EVM coin info already set. Make sure you run the configurator's ResetTestConfig before trying to set a new evm coin info")
	}
	testingEvmCoinInfo = new(EvmCoinInfo)
	// fill up the denom with default values
	// if EvmCoinInfo is not defined
	if eci.Denom == "" {
		eci = defaultCoinInfo
	}
	if err := setEVMCoinDenom(eci.Denom); err != nil {
		return err
	}
	return setEVMCoinDecimals(eci.Decimals)
}

// resetEVMCoinInfo resets to nil the testingEVMCoinInfo
func resetEVMCoinInfo() {
	testingEvmCoinInfo = nil
}
