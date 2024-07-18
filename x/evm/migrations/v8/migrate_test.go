// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v8_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/encoding"
	v8 "github.com/evmos/evmos/v18/x/evm/migrations/v8"
	v7types "github.com/evmos/evmos/v18/x/evm/migrations/v8/types"
	"github.com/evmos/evmos/v18/x/evm/types"
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	// Initialize the store
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_key")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	chainConfig := types.DefaultChainConfig()
	bz, err := json.Marshal(chainConfig)
	require.NoError(t, err)
	var chainCfgV7 v7types.V7ChainConfig
	err = json.Unmarshal(bz, &chainCfgV7)
	require.NoError(t, err)

	// Create a pre migration environment with default params.
	paramsV7 := v7types.V7Params{
		EvmDenom:            types.DefaultEVMDenom,
		ChainConfig:         chainCfgV7,
		ExtraEIPs:           v7types.DefaultExtraEIPs,
		AllowUnprotectedTxs: types.DefaultAllowUnprotectedTxs,
		ActivePrecompiles:   types.AvailableEVMExtensions,
		EVMChannels:         types.DefaultEVMChannels,
	}
	paramsV7Bz := cdc.MustMarshal(&paramsV7)
	kvStore.Set(types.KeyPrefixParams, paramsV7Bz)

	err = v8.MigrateStore(ctx, storeKey, cdc)
	require.NoError(t, err)

	paramsBz := kvStore.Get(types.KeyPrefixParams)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, types.DefaultEVMDenom, params.EvmDenom)
	require.False(t, params.AllowUnprotectedTxs)
	require.Equal(t, types.DefaultChainConfig(), params.ChainConfig)
	require.Equal(t, types.DefaultExtraEIPs, params.ExtraEIPs)
	require.Equal(t, types.DefaultEVMChannels, params.EVMChannels)
	require.Equal(t, types.DefaultAccessControl, params.AccessControl)
}
