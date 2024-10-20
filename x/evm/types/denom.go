// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convenient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package types

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/utils"
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
	utils.ICSChainID: {
		Denom:        "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		DisplayDenom: "atom",
		Decimals:     SixDecimals,
	},
}
