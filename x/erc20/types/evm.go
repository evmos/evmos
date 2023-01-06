// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

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
