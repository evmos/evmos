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

// EvmDenom struct holds the name and decimals of the EVM denom
type EvmDenom struct {
	denom    string
	decimals Decimals
}

var evmDenom EvmDenom

func SetDecimals(d Decimals) {
	if d != SixDecimals && d != EighteenDecimals {
		panic("evm does not support these decimals")
	}

	evmDenom.decimals = d
}

func SetDenom(denom string) {
	evmDenom.denom = denom
}

func GetDecimals() Decimals {
	return evmDenom.decimals
}

func GetDenom() string {
	return evmDenom.denom
}

func SetEVMDenom(evmdenom EvmDenom) {
	SetDenom(evmdenom.denom)
	SetDecimals(evmdenom.decimals)
}
