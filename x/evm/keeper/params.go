// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/x/evm/types"
	"golang.org/x/exp/slices"
)

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixParams)
	if len(bz) == 0 {
		return k.GetLegacyParams(ctx)
	}
	k.cdc.MustUnmarshal(bz, &params)
	return
}

// SetParams sets the EVM params each in their individual key for better get performance
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	// NOTE: We need to sort the precompiles in order to enable searching with binary search
	// in params.IsActivePrecompile.
	slices.Sort(params.ActivePrecompiles)

	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.KeyPrefixParams, bz)
	return nil
}

// GetLegacyParams returns param set for version before migrate
func (k Keeper) GetLegacyParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.ss.GetParamSetIfExists(ctx, &params)
	return params
}

// EnableStaticPrecompiles appends the addresses of the given Precompiles to the list
// of active precompiles.
func (k Keeper) EnableStaticPrecompiles(ctx sdk.Context, addresses ...common.Address) error {
	params := k.GetParams(ctx)
	activePrecompiles := params.ActivePrecompiles

	// Append and sort the new precompiles
	activePrecompiles, err := appendPrecompiles(activePrecompiles, addresses...)
	if err != nil {
		return err
	}

	params.ActivePrecompiles = activePrecompiles

	return k.SetParams(ctx, params)
}

// EnableDynamicPrecompiles appends the addresses of the given Precompiles to the list
// of active precompiles.
func (k Keeper) EnableDynamicPrecompiles(ctx sdk.Context, addresses ...common.Address) error {
	// Get the current params and append the new precompiles
	params := k.GetParams(ctx)
	activePrecompiles := params.ActiveDynamicPrecompiles

	// Append and sort the new precompiles
	activePrecompiles, err := appendPrecompiles(activePrecompiles, addresses...)
	if err != nil {
		return err
	}

	// Update params
	params.ActiveDynamicPrecompiles = activePrecompiles
	return k.SetParams(ctx, params)
}

func appendPrecompiles(existingPrecompiles []string, addresses ...common.Address) ([]string, error) {
	updatedPrecompiles := []string{}
	for _, address := range addresses {
		// Check for duplicates
		if slices.Contains(existingPrecompiles, address.String()) {
			return nil, fmt.Errorf("precompile already registered: %s", address)
		}
		updatedPrecompiles = append(updatedPrecompiles, address.String())
	}
	updatedPrecompiles = append(existingPrecompiles, updatedPrecompiles...)
	sortPrecompiles(updatedPrecompiles)
	return updatedPrecompiles, nil
}

func sortPrecompiles(precompiles []string) {
	sort.Slice(precompiles, func(i, j int) bool {
		return precompiles[i] < precompiles[j]
	})
}

// EnableEIPs enables the given EIPs in the EVM parameters.
func (k Keeper) EnableEIPs(ctx sdk.Context, eips ...int64) error {
	evmParams := k.GetParams(ctx)
	evmParams.ExtraEIPs = append(evmParams.ExtraEIPs, eips...)

	sort.Slice(evmParams.ExtraEIPs, func(i, j int) bool {
		return evmParams.ExtraEIPs[i] < evmParams.ExtraEIPs[j]
	})

	return k.SetParams(ctx, evmParams)
}
