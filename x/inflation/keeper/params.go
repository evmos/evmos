package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/evmos/evmos/v10/x/inflation/types"
)

// GetParams returns the total set of inflation parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	mintDenom := k.GetMintDenom(ctx)
	enableInflation := k.IsInflationEnabled(ctx)
	inflationDistribution := k.GetInflationDistribution(ctx)
	exponentialCalculation := k.GetExponentialCalculation(ctx)

	return types.NewParams(mintDenom, exponentialCalculation, inflationDistribution, enableInflation)
}

// SetParams sets the inflation parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)

	mintDenomBz := k.cdc.MustMarshal(&gogotypes.StringValue{Value: params.MintDenom})
	enableInflationBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableInflation})
	inflationDistribBz := k.cdc.MustMarshal(&params.InflationDistribution)
	expCalculationBz := k.cdc.MustMarshal(&params.ExponentialCalculation)

	store.Set(types.ParamStoreKeyMintDenom, mintDenomBz)
	store.Set(types.ParamStoreKeyEnableInflation, enableInflationBz)
	store.Set(types.ParamStoreKeyInflationDistribution, inflationDistribBz)
	store.Set(types.ParamStoreKeyExponentialCalculation, expCalculationBz)

	return nil
}

// GetMintDenom returns the mint denomination
func (k Keeper) GetMintDenom(ctx sdk.Context) string {
	var mintDenom gogotypes.StringValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyMintDenom)
	if bz == nil {
		return mintDenom.Value
	}
	k.cdc.MustUnmarshal(bz, &mintDenom)
	return mintDenom.Value
}

// IsInflationEnabled returns if Inflation is enabled
func (k Keeper) IsInflationEnabled(ctx sdk.Context) bool {
	var enableInflation gogotypes.BoolValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEnableInflation)
	if bz == nil {
		return enableInflation.Value
	}
	k.cdc.MustUnmarshal(bz, &enableInflation)
	return enableInflation.Value
}

// GetInflationDistribution returns the e distribution in which inflation is
// allocated through minting on each epoch
func (k Keeper) GetInflationDistribution(ctx sdk.Context) types.InflationDistribution {
	var inflationDistribution types.InflationDistribution
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyInflationDistribution)
	if bz == nil {
		return inflationDistribution
	}
	k.cdc.MustUnmarshal(bz, &inflationDistribution)
	return inflationDistribution
}

// GetExponentialCalculation returns the factors to calculate exponential inflation
func (k Keeper) GetExponentialCalculation(ctx sdk.Context) types.ExponentialCalculation {
	var exponentialCalculation types.ExponentialCalculation
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyExponentialCalculation)
	if bz == nil {
		return exponentialCalculation
	}
	k.cdc.MustUnmarshal(bz, &exponentialCalculation)
	return exponentialCalculation
}
