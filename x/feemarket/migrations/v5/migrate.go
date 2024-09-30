// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v5

import (
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesV4 "github.com/evmos/evmos/v20/x/feemarket/migrations/v5/types"
	"github.com/evmos/evmos/v20/x/feemarket/types"
)

// MigrateStore migrates the x/feemarket module state from the consensus version 4 to
// version 5. Specifically, it converts the base fee from Int to LegacyDec.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) error {
	var (
		store    = ctx.KVStore(storeKey)
		paramsV4 typesV4.ParamsV4
		params   types.Params
	)

	paramsV4Bz := store.Get(types.ParamsKey)
	cdc.MustUnmarshal(paramsV4Bz, &paramsV4)

	params.NoBaseFee = paramsV4.NoBaseFee
	params.BaseFeeChangeDenominator = paramsV4.BaseFeeChangeDenominator
	params.ElasticityMultiplier = paramsV4.ElasticityMultiplier
	params.EnableHeight = paramsV4.EnableHeight
	params.BaseFee = math.LegacyNewDecFromInt(paramsV4.BaseFee) // convert to dec
	params.MinGasPrice = paramsV4.MinGasPrice
	params.MinGasMultiplier = paramsV4.MinGasMultiplier

	if err := params.Validate(); err != nil {
		return err
	}

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}
