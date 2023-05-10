// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

// ERC20Data represents the ERC20 token details used to map
// the token to a Cosmos Coin
type ERC20Data struct {
	Name     string
	Symbol   string
	Decimals uint8
}

// ERC20StringResponse defines the string value from the call response
type ERC20StringResponse struct {
	Value string
}

// ERC20Uint8Response defines the uint8 value from the call response
type ERC20Uint8Response struct {
	Value uint8
}

// ERC20BoolResponse defines the bool value from the call response
type ERC20BoolResponse struct {
	Value bool
}

// NewERC20Data creates a new ERC20Data instance
func NewERC20Data(name, symbol string, decimals uint8) ERC20Data {
	return ERC20Data{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}
}
