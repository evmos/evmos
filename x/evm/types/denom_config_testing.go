// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convenient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

//go:build test
// +build test

package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/utils"
)

// Decimals is a wrapper around uint64 to represent the decimal representation
// of a Cosmos coin.
type Decimals uint64

const (
	// SixDecimals is the Decimals used for Cosmos coin with 6 decimals.
	SixDecimals Decimals = 6
	// EighteenDecimals is the Decimals used for Cosmos coin with 18 decimals.
	EighteenDecimals Decimals = 18
)

// EvmCoinInfo struct holds the name and decimals of the EVM denom. The EVM denom
// is the token used to pay fees in the EVM.
type EvmCoinInfo struct {
	Denom        string
	DisplayDenom string
	Decimals     Decimals
}

// ChainsCoinInfo is a map of the chain id and its corresponding EvmCoinInfo
// that allows initializing the app with different coin info based on the
// chain id
var ChainsCoinInfo = map[string]EvmCoinInfo{
	utils.MainnetChainID: {
		Denom:        types.BaseDenom,
		DisplayDenom: types.DisplayDenom,
		Decimals:     EighteenDecimals,
	},
	utils.TestnetChainID: {
		Denom:        types.BaseDenomTestnet,
		DisplayDenom: types.DisplayDenomTestnet,
		Decimals:     EighteenDecimals,
	},
	utils.SixDecChainID: {
		Denom:        types.BaseDenom,
		DisplayDenom: types.DisplayDenom,
		Decimals:     SixDecimals,
	},
}

// testingEvmCoinInfo hold the information of the coin used in the EVM as gas token. It
// can only be set via `EVMConfigurator` before starting the app.
var testingEvmCoinInfo *EvmCoinInfo

// setEVMCoinDecimals allows to define the decimals used in the representation
// of the EVM coin.
func setEVMCoinDecimals(d Decimals) {
	if d != SixDecimals && d != EighteenDecimals {
		panic(fmt.Errorf("invalid decimal value %d; the evm supports only 6 and 18 decimals", d))
	}

	testingEvmCoinInfo.Decimals = d
}

// setEVMCoinDenom allows to define the denom of the coin used in the EVM.
func setEVMCoinDenom(denom string) {
	if err := sdk.ValidateDenom(denom); err != nil {
		panic(err)
	}
	testingEvmCoinInfo.Denom = denom
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
func setTestingEVMCoinInfo(evmdenom EvmCoinInfo) {
	if testingEvmCoinInfo != nil {
		panic("testing EVM coin info already set. Make sure you run the configurator's ResetTestChainConfig before trying to set a new evm coin info")
	}
	testingEvmCoinInfo = new(EvmCoinInfo)
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

// resetEVMCoinInfo resets to nil the testingEVMCoinInfo
func resetEVMCoinInfo() {
	testingEvmCoinInfo = nil
}
