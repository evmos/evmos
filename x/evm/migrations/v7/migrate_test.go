// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v7_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/encoding"
	v7 "github.com/evmos/evmos/v16/x/evm/migrations/v7"
	v6types "github.com/evmos/evmos/v16/x/evm/migrations/v7/types"
	"github.com/evmos/evmos/v16/x/evm/types"
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	chainConfig := types.DefaultChainConfig()
	bz, err := json.Marshal(chainConfig)
	require.NoError(t, err)
	var chainCfgv16 v6types.V6ChainConfig
	err = json.Unmarshal(bz, &chainCfgv16)
	require.NoError(t, err)
	v16Params := v6types.V6Params{
		EvmDenom:            types.DefaultEVMDenom,
		EnableCreate:        types.DefaultEnableCreate,
		EnableCall:          types.DefaultEnableCall,
		ChainConfig:         chainCfgv16,
		ExtraEIPs:           types.DefaultExtraEIPs,
		AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   types.AvailableEVMExtensions,
		EVMChannels:         types.DefaultEVMChannels,
	}

	// Set the params in the store
	paramsV16Bz := cdc.MustMarshal(&v16Params)
	kvStore.Set(types.KeyPrefixParams, paramsV16Bz)

	err = v7.MigrateStore(ctx, storeKey, cdc)
	require.NoError(t, err)

	paramsBz := kvStore.Get(types.KeyPrefixParams)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	// test that the params have been migrated correctly
	require.Equal(t, types.DefaultEVMDenom, params.EvmDenom)
	require.True(t, params.EnableCreate)
	require.True(t, params.EnableCall)
	require.False(t, params.AllowUnprotectedTxs)
	require.Equal(t, chainConfig, params.ChainConfig)
	require.Equal(t, types.DefaultExtraEIPs, params.ExtraEIPs)
	require.Equal(t, types.DefaultEVMChannels, params.EVMChannels)
	require.Equal(t, types.AvailableEVMExtensions, params.ActiveStaticPrecompiles)
}
