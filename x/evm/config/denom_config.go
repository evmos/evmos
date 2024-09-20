// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

type Decimals uint32

const (
	SixDecimals      Decimals = 6
	EighteenDecimals Decimals = 18
)

// EvmCoinInfo struct holds the name and decimals of the EVM denom. The EVM denom
// is the token used to pay fees in the EVM.
type EvmCoinInfo struct {
	denom    string
	decimals Decimals
}

var evmCoinInfo EvmCoinInfo

// setEVMCoinDecimals allows to define the decimals used in the representation
// of the EVM coin.
func setEVMCoinDecimals(d Decimals) {
	if d != SixDecimals && d != EighteenDecimals {
		panic("evm does not support these decimals")
	}

	evmCoinInfo.decimals = d
}

// setEVMCoinDenom allows to define the denom of the coin used in the EVM.
func setEVMCoinDenom(denom string) {
	evmCoinInfo.denom = denom
}

// GetEVMCoinDecimals returns the decimals used in the representation of the EVM
// coin.
func GetEVMCoinDecimals() Decimals {
	return evmCoinInfo.decimals
}

func GetDenom() string {
	return evmCoinInfo.denom
}

// setEVMCoinInfo allows to define denom and decimals of the coin used in the EVM.
func setEVMCoinInfo(evmdenom EvmCoinInfo) {
	setEVMCoinDenom(evmdenom.denom)
	setEVMCoinDecimals(evmdenom.decimals)
}
