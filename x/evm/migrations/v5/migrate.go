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
package v5

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v11/x/evm/types"

	v5types "github.com/evmos/evmos/v11/x/evm/migrations/v5/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 4 to
// version 5. Specifically, it takes the parameters that are currently stored
// in separate keys and stores them directly into the x/evm module state using
// a single params key.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		extraEIPs   v5types.V5ExtraEIPs
		chainConfig types.ChainConfig
		params      types.Params
	)

	store := ctx.KVStore(storeKey)

	denom := string(store.Get(types.ParamStoreKeyEVMDenom))

	extraEIPsBz := store.Get(types.ParamStoreKeyExtraEIPs)
	cdc.MustUnmarshal(extraEIPsBz, &extraEIPs)

	// revert ExtraEIP change for Evmos testnet
	if ctx.ChainID() == "evmos_9000-4" {
		extraEIPs.EIPs = []int64{}
	}

	chainCfgBz := store.Get(types.ParamStoreKeyChainConfig)
	cdc.MustUnmarshal(chainCfgBz, &chainConfig)

	params.EvmDenom = denom
	params.ExtraEIPs = extraEIPs.EIPs
	params.ChainConfig = chainConfig
	params.EnableCreate = store.Has(types.ParamStoreKeyEnableCreate)
	params.EnableCall = store.Has(types.ParamStoreKeyEnableCall)
	params.AllowUnprotectedTxs = store.Has(types.ParamStoreKeyAllowUnprotectedTxs)

	store.Delete(types.ParamStoreKeyChainConfig)
	store.Delete(types.ParamStoreKeyExtraEIPs)
	store.Delete(types.ParamStoreKeyEVMDenom)
	store.Delete(types.ParamStoreKeyEnableCreate)
	store.Delete(types.ParamStoreKeyEnableCall)
	store.Delete(types.ParamStoreKeyAllowUnprotectedTxs)

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)

	store.Set(types.KeyPrefixParams, bz)
	return nil
}
