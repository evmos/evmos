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
package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v19/x/evm/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/encoding"
	v4 "github.com/evmos/evmos/v19/x/evm/migrations/v4"
	v4types "github.com/evmos/evmos/v19/x/evm/migrations/v4/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey(types.TransientKey)
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	// Get all the new parameters from the kvStore
	var evmDenom string
	bz := kvStore.Get(types.ParamStoreKeyEVMDenom)
	evmDenom = string(bz)

	allowUnprotectedTx := kvStore.Has(types.ParamStoreKeyAllowUnprotectedTxs)
	enableCreate := kvStore.Has(types.ParamStoreKeyEnableCreate)
	enableCall := kvStore.Has(types.ParamStoreKeyEnableCall)

	var chainCfg v4types.V4ChainConfig
	bz = kvStore.Get(types.ParamStoreKeyChainConfig)
	cdc.MustUnmarshal(bz, &chainCfg)

	var extraEIPs v4types.ExtraEIPs
	bz = kvStore.Get(types.ParamStoreKeyExtraEIPs)
	cdc.MustUnmarshal(bz, &extraEIPs)
	require.Equal(t, types.DefaultExtraEIPs, extraEIPs.EIPs)

	params := v4types.V4Params{
		EvmDenom:            evmDenom,
		AllowUnprotectedTxs: allowUnprotectedTx,
		EnableCreate:        enableCreate,
		EnableCall:          enableCall,
		V4ChainConfig:       chainCfg,
		ExtraEIPs:           extraEIPs,
	}

	require.Equal(t, legacySubspace.ps.AllowUnprotectedTxs, params.AllowUnprotectedTxs)
	require.Equal(t, legacySubspace.ps.ExtraEIPs, params.ExtraEIPs.EIPs)
	require.EqualValues(t, legacySubspace.ps.ChainConfig, params.V4ChainConfig)
}
