// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convenient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

//go:build !test
// +build !test

package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NOTE: Remember to add the ConversionFactor associated with constants.
const (
	// SixDecimals is the Decimals used for Cosmos coin with 6 decimals.
	SixDecimals Decimals = 6
	// EighteenDecimals is the Decimals used for Cosmos coin with 18 decimals.
	EighteenDecimals Decimals = 18
)

// Decimals represents the decimal representation of a Cosmos coin.
type Decimals uint8

// Validate checks if the Decimals instance represent a supported decimals value
// or not.
func (d Decimals) Validate() error {
	switch d {
	case SixDecimals:
		return nil
	case EighteenDecimals:
		return nil
	default:
		return fmt.Errorf("received unsupported decimals: %d", d)
	}
}

// ConversionFactor returns the conversion factor between the Decimals value and
// the 18 decimals representation, i.e. `EighteenDecimals`.
//
// NOTE: This function does not check if the Decimal instance is valid or
// not and by default returns the conversion factor of 1, i.e. from 18 decimals
// to 18 decimals. We cannot have a non supported Decimal since it is checked
// and validated.
func (d Decimals) ConversionFactor() math.Int {
	if d == SixDecimals {
		return math.NewInt(1e12)
	}

	return math.NewInt(1)
}

// EvmCoinInfo struct holds the name and decimals of the EVM denom. The EVM denom
// is the token used to pay fees in the EVM.
type EvmCoinInfo struct {
	Denom    string
	Decimals Decimals
}

// evmCoinInfo hold the information of the coin used in the EVM as gas token. It
// can only be set via `EVMConfigurator` before starting the app.
var evmCoinInfo *EvmCoinInfo

// setEVMCoinDecimals allows to define the decimals used in the representation
// of the EVM coin.
func setEVMCoinDecimals(d Decimals) error {
	if err := d.Validate(); err != nil {
		return fmt.Errorf("setting EVM coin decimals: %w", err)
	}

	evmCoinInfo.Decimals = d
	return nil
}

// setEVMCoinDenom allows to define the denom of the coin used in the EVM.
func setEVMCoinDenom(denom string) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return fmt.Errorf("setting EVM coin denom: %w", err)
	}
	evmCoinInfo.Denom = denom
	return nil
}

// GetEVMCoinDecimals returns the decimals used in the representation of the EVM
// coin.
func GetEVMCoinDecimals() Decimals {
	return evmCoinInfo.Decimals
}

// GetEVMCoinDenom returns the denom used for the EVM coin.
func GetEVMCoinDenom() string {
	return evmCoinInfo.Denom
}

// setEVMCoinInfo allows to define denom and decimals of the coin used in the EVM.
func setEVMCoinInfo(eci EvmCoinInfo) error {
	if evmCoinInfo != nil {
		return errors.New("EVM coin info already set")
	}

	evmCoinInfo = new(EvmCoinInfo)

	if err := setEVMCoinDenom(eci.Denom); err != nil {
		return err
	}
	return setEVMCoinDecimals(eci.Decimals)
}
