// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v8_test

import (
	"encoding/json"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v20/encoding"
	v8 "github.com/evmos/evmos/v20/x/evm/migrations/v8"
	v7types "github.com/evmos/evmos/v20/x/evm/migrations/v8/types"
	"github.com/evmos/evmos/v20/x/evm/types"
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig()
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	chainConfig := types.DefaultChainConfig(ctx.ChainID())
	bz, err := json.Marshal(chainConfig)
	require.NoError(t, err)
	var chainCfgV7 v7types.V7ChainConfig
	err = json.Unmarshal(bz, &chainCfgV7)
	require.NoError(t, err)

	// Create a pre migration environment with default params.
	paramsV7 := v7types.V7Params{
		EvmDenom:                v7types.DefaultEVMDenom,
		ChainConfig:             chainCfgV7,
		ExtraEIPs:               types.DefaultExtraEIPs,
		AllowUnprotectedTxs:     types.DefaultAllowUnprotectedTxs,
		ActiveStaticPrecompiles: types.DefaultStaticPrecompiles,
		EVMChannels:             types.DefaultEVMChannels,
	}
	paramsV6Bz := cdc.MustMarshal(&paramsV7)
	kvStore.Set(types.KeyPrefixParams, paramsV6Bz)

	err = v8.MigrateStore(ctx, storeKey, cdc)
	require.NoError(t, err)

	paramsBz := kvStore.Get(types.KeyPrefixParams)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.False(t, params.AllowUnprotectedTxs)
	require.Equal(t, types.DefaultExtraEIPs, params.ExtraEIPs)
	require.Equal(t, types.DefaultEVMChannels, params.EVMChannels)
	require.Equal(t, types.DefaultAccessControl, params.AccessControl)
}
