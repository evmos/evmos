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

package v3

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3types "github.com/evmos/evmos/v10/x/erc20/migrations/v3/types"
	"github.com/evmos/evmos/v10/x/erc20/types"
)

var isTrue = []byte{0x01}

// MigrateStore migrates the x/erc20 module state from the consensus version 2 to
// version 3. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/erc20 module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	legacySubspace types.Subspace,
) error {
	store := ctx.KVStore(storeKey)
	var params v3types.V3Params
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	if params.EnableErc20 {
		store.Set(v3types.ParamStoreKeyEnableErc20, isTrue)
	}

	if params.EnableEVMHook {
		store.Set(v3types.ParamStoreKeyEnableEVMHook, isTrue)
	}

	return nil
}
