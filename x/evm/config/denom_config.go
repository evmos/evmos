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

var decimals Decimals

set and get