// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v6_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v17/app"
	"github.com/evmos/evmos/v17/encoding"
	v6 "github.com/evmos/evmos/v17/x/evm/migrations/v6"
	v5types "github.com/evmos/evmos/v17/x/evm/migrations/v6/types"
	"github.com/evmos/evmos/v17/x/evm/types"
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
	var chainCfgV5 v5types.V5ChainConfig
	err = json.Unmarshal(bz, &chainCfgV5)
	require.NoError(t, err)
	v5Params := v5types.V5Params{
		EvmDenom:            types.DefaultEVMDenom,
		EnableCreate:        types.DefaultEnableCreate,
		EnableCall:          types.DefaultEnableCall,
		ChainConfig:         chainCfgV5,
		ExtraEIPs:           types.DefaultExtraEIPs,
		AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   types.AvailableEVMExtensions,
	}

	// Set the params in the store
	paramsV5Bz := cdc.MustMarshal(&v5Params)
	kvStore.Set(types.KeyPrefixParams, paramsV5Bz)

	err = v6.MigrateStore(ctx, storeKey, cdc)
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
}
