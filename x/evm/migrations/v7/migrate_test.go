// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
<<<<<<< HEAD
=======

>>>>>>> main
package v7_test

import (
	"encoding/json"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
<<<<<<< HEAD
	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/encoding"
	v7 "github.com/evmos/evmos/v18/x/evm/migrations/v7"
	v6types "github.com/evmos/evmos/v18/x/evm/migrations/v7/types"
	"github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
=======
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/encoding"
	v7 "github.com/evmos/evmos/v19/x/evm/migrations/v7"
	v6types "github.com/evmos/evmos/v19/x/evm/migrations/v7/types"
	"github.com/evmos/evmos/v19/x/evm/types"
>>>>>>> main
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

<<<<<<< HEAD
	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
=======
<<<<<<<< HEAD:x/evm/migrations/v6/migrate_test.go
	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
========
	// Initialize the store
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_storekey")
>>>>>>>> main:x/evm/migrations/v7/migrate_test.go
>>>>>>> main
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	chainConfig := types.DefaultChainConfig()
	bz, err := json.Marshal(chainConfig)
	require.NoError(t, err)
	var chainCfgV6 v6types.V6ChainConfig
	err = json.Unmarshal(bz, &chainCfgV6)
	require.NoError(t, err)
<<<<<<< HEAD
	v6Params := v6types.V6Params{
		EvmDenom:            types.DefaultEVMDenom,
		ChainConfig:         chainCfgV6,
		ExtraEIPs:           types.DefaultExtraEIPs,
		AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   types.DefaultStaticPrecompiles,
		EVMChannels:         types.DefaultEVMChannels,
	}

	// Set the params in the store
	paramsV6Bz := cdc.MustMarshal(&v6Params)
=======

	// Create a pre migration environment with default params.
	paramsV6 := v6types.V6Params{
		EvmDenom:            types.DefaultEVMDenom,
<<<<<<<< HEAD:x/evm/migrations/v6/migrate_test.go
		ChainConfig:         chainCfgV5,
		ExtraEIPs:           types.DefaultExtraEIPs,
		AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   types.DefaultStaticPrecompiles,
========
		ChainConfig:         chainCfgV6,
		ExtraEIPs:           v6types.DefaultExtraEIPs,
		AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   types.DefaultStaticPrecompiles,
		EVMChannels:         types.DefaultEVMChannels,
>>>>>>>> main:x/evm/migrations/v7/migrate_test.go
	}
	paramsV6Bz := cdc.MustMarshal(&paramsV6)
>>>>>>> main
	kvStore.Set(types.KeyPrefixParams, paramsV6Bz)

	err = v7.MigrateStore(ctx, storeKey, cdc)
	require.NoError(t, err)

	paramsBz := kvStore.Get(types.KeyPrefixParams)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

<<<<<<< HEAD
	// test that the params have been migrated correctly
=======
>>>>>>> main
	require.Equal(t, types.DefaultEVMDenom, params.EvmDenom)
	require.False(t, params.AllowUnprotectedTxs)
	require.Equal(t, chainConfig, params.ChainConfig)
	require.Equal(t, types.DefaultExtraEIPs, params.ExtraEIPs)
	require.Equal(t, types.DefaultEVMChannels, params.EVMChannels)
	require.Equal(t, types.DefaultAccessControl, params.AccessControl)
}
