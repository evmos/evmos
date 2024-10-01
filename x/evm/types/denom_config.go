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

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Decimals is a wrapper around uint32 to represent the decimal representation
// of a Cosmos coin.
type Decimals uint32

const (
	// SixDecimals is the Decimals used for Cosmos coin with 6 decimals.
	SixDecimals Decimals = 6
	// EighteenDecimals is the Decimals used for Cosmos coin with 18 decimals.
	EighteenDecimals Decimals = 18
)

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
func setEVMCoinDecimals(d Decimals) {
	if d != SixDecimals && d != EighteenDecimals {
		panic(fmt.Errorf("invalid decimal value %d; the evm supports only 6 and 18 decimals", d))
	}

	evmCoinInfo.Decimals = d
}

// setEVMCoinDenom allows to define the denom of the coin used in the EVM.
func setEVMCoinDenom(denom string) {
	if err := sdk.ValidateDenom(denom); err != nil {
		panic(err)
	}
	evmCoinInfo.Denom = denom
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
func setEVMCoinInfo(evmdenom EvmCoinInfo) {
	if evmCoinInfo != nil {
		panic("EVM coin info already set")
	}
	evmCoinInfo = new(EvmCoinInfo)
	setEVMCoinDenom(evmdenom.Denom)
	setEVMCoinDecimals(evmdenom.Decimals)
}

// ConversionFactor returns the conversion factor between the Decimals value and
// the 18 decimals representation, i.e. `EighteenDecimals`.
//
// NOTE: This function does not check if the Decimal instance is valid or
// not and by default returns the conversion factor of 1, i.e. from 18 decimals
// to 18 decimals.
func (d Decimals) ConversionFactor() math.Int {
	if d == SixDecimals {
		return math.NewInt(1e12)
	}

	return math.NewInt(1)
}
