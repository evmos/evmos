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
package v5_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/encoding"
	v5 "github.com/evmos/evmos/v11/x/evm/migrations/v5"
	v5types "github.com/evmos/evmos/v11/x/evm/migrations/v5/types"
	"github.com/evmos/evmos/v11/x/evm/types"
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	extraEIPs := v5types.V5ExtraEIPs{EIPs: types.AvailableExtraEIPs}
	extraEIPsBz := cdc.MustMarshal(&extraEIPs)
	chainConfig := types.DefaultChainConfig()
	chainConfigBz := cdc.MustMarshal(&chainConfig)

	// Set the params in the store
	kvStore.Set(types.ParamStoreKeyEVMDenom, []byte(types.DefaultEVMDenom))
	kvStore.Set(types.ParamStoreKeyEnableCreate, []byte{0x01})
	kvStore.Set(types.ParamStoreKeyEnableCall, []byte{0x01})
	kvStore.Set(types.ParamStoreKeyAllowUnprotectedTxs, []byte{0x01})
	kvStore.Set(types.ParamStoreKeyExtraEIPs, extraEIPsBz)
	kvStore.Set(types.ParamStoreKeyChainConfig, chainConfigBz)

	err := v5.MigrateStore(ctx, storeKey, cdc)
	require.NoError(t, err)

	paramsBz := kvStore.Get(types.KeyPrefixParams)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	// test that the params have been migrated correctly
	require.Equal(t, types.DefaultEVMDenom, params.EvmDenom)
	require.True(t, params.EnableCreate)
	require.True(t, params.EnableCall)
	require.True(t, params.AllowUnprotectedTxs)
	require.Equal(t, chainConfig, params.ChainConfig)
	require.Equal(t, extraEIPs.EIPs, params.ExtraEIPs)

	// check that the keys are deleted
	require.False(t, kvStore.Has(types.ParamStoreKeyEVMDenom))
	require.False(t, kvStore.Has(types.ParamStoreKeyEnableCreate))
	require.False(t, kvStore.Has(types.ParamStoreKeyEnableCall))
	require.False(t, kvStore.Has(types.ParamStoreKeyAllowUnprotectedTxs))
	require.False(t, kvStore.Has(types.ParamStoreKeyExtraEIPs))
	require.False(t, kvStore.Has(types.ParamStoreKeyChainConfig))
}
