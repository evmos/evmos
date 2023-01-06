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

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v10/x/incentives/types"
)

// GetIncentivesGasMeters - get all registered GasMeters per Incentive
func (k Keeper) GetIncentivesGasMeters(ctx sdk.Context) []types.GasMeter {
	gms := []types.GasMeter{}

	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixGasMeter)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contract, userAddress := types.SplitGasMeterKey(iterator.Key())
		gas := sdk.BigEndianToUint64(iterator.Value())

		gm := types.GasMeter{
			Contract:      contract.String(),
			Participant:   userAddress.String(),
			CumulativeGas: gas,
		}

		gms = append(gms, gm)
	}

	return gms
}

// GetIncentiveGasMeters - get all registered GasMeters per contract
func (k Keeper) GetIncentiveGasMeters(
	ctx sdk.Context,
	contract common.Address,
) []types.GasMeter {
	gms := []types.GasMeter{}

	k.IterateIncentiveGasMeters(
		ctx, contract,
		func(gm types.GasMeter) (stop bool) {
			gms = append(gms, gm)
			return false
		},
	)

	return gms
}

// IterateIncentiveGasMeters iterates over all the given registered incentivized
// contract's `GasMeter` and performs a callback.
func (k Keeper) IterateIncentiveGasMeters(
	ctx sdk.Context,
	contract common.Address,
	handlerFn func(gm types.GasMeter) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)

	iterator := sdk.KVStorePrefixIterator(store, contract.Bytes())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contract, userAddress := types.SplitGasMeterKey(iterator.Key())
		gas := sdk.BigEndianToUint64(iterator.Value())
		gm := types.GasMeter{
			Contract:      contract.String(),
			Participant:   userAddress.String(),
			CumulativeGas: gas,
		}

		if handlerFn(gm) {
			break
		}
	}
}

// GetGasMeter - get cumulativeGas from gas meter
func (k Keeper) GetGasMeter(
	ctx sdk.Context,
	contract, participant common.Address,
) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)
	key := append(contract.Bytes(), participant.Bytes()...)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, false
	}

	return sdk.BigEndianToUint64(bz), true
}

// SetGasMeter stores a gasMeter
func (k Keeper) SetGasMeter(ctx sdk.Context, gm types.GasMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)
	contract := common.HexToAddress(gm.Contract)
	participant := common.HexToAddress(gm.Participant)
	key := append(contract.Bytes(), participant.Bytes()...)
	store.Set(key, sdk.Uint64ToBigEndian(gm.CumulativeGas))
}

// DeleteGasMeter removes a gasMeter.
func (k Keeper) DeleteGasMeter(ctx sdk.Context, gm types.GasMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)
	contract := common.HexToAddress(gm.Contract)
	participant := common.HexToAddress(gm.Participant)
	key := append(contract.Bytes(), participant.Bytes()...)
	store.Delete(key)
}
