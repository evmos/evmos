// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v8

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v7types "github.com/evmos/evmos/v20/x/evm/migrations/v8/types"
	"github.com/evmos/evmos/v20/x/evm/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 7 to
// version 8.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		paramsV7 v7types.V7Params
		params   types.Params
	)

	store := ctx.KVStore(storeKey)

	paramsV7Bz := store.Get(types.KeyPrefixParams)
	cdc.MustUnmarshal(paramsV7Bz, &paramsV7)
	params.AllowUnprotectedTxs = paramsV7.AllowUnprotectedTxs
	params.ActiveStaticPrecompiles = paramsV7.ActiveStaticPrecompiles

	params.EVMChannels = paramsV7.EVMChannels
	params.AccessControl.Call.AccessType = types.AccessType(paramsV7.AccessControl.Call.AccessType)
	params.AccessControl.Create.AccessControlList = paramsV7.AccessControl.Create.AccessControlList
	params.AccessControl.Call.AccessControlList = paramsV7.AccessControl.Call.AccessControlList
	params.AccessControl.Create.AccessType = types.AccessType(paramsV7.AccessControl.Create.AccessType)
	params.ExtraEIPs = paramsV7.ExtraEIPs

	if err := params.Validate(); err != nil {
		return err
	}

	bz := cdc.MustMarshal(&params)

	store.Set(types.KeyPrefixParams, bz)
	return nil
}
