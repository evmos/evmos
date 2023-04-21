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

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// RegisterInterfaces registers the tendermint concrete client-related
// implementations and interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*authtypes.AccountI)(nil),
		&EthAccount{},
	)
	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&EthAccount{},
	)
	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&ExtensionOptionsWeb3Tx{},
		&ExtensionOptionDynamicFeeTx{},
	)
}
