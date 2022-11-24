package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/inflation/exported"
	v2types "github.com/evmos/evmos/v10/x/inflation/migrations/v2/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// MigrateStore migrates the x/inflation module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/inflation module state.
func MigrateStore(
	ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	var params v2types.Params
	legacySubspace.GetParamSet(ctx, &params)

	//expected: types.Params(types.Params{MintDenom:"aphoton", ExponentialCalculation:types.ExponentialCalculation{A:300000000.000000000000000000, R:0.500000000000000000, C:9375000.000000000000000000, BondingTarget:0.660000000000000000, MaxVariance:0.000000000000000000}, InflationDistribution:types.InflationDistribution{StakingRewards:0.533333334000000000, UsageIncentives:0.333333333000000000, CommunityPool:0.133333333000000000}, EnableInflation:true})
	//actual  : types.Params(types.Params{MintDenom:"aphoton", ExponentialCalculation:types.ExponentialCalculation{A:300000000.000000000000000000, R:0.500000000000000000, C:9375000.000000000000000000, BondingTarget:0.660000000000000000, MaxVariance:0.000000000000000000}, InflationDistribution:types.InflationDistribution{StakingRewards:0.533333334000000000, UsageIncentives:0.333333333000000000, CommunityPool:0.133333333000000000}, EnableInflation:true})
	//Test:           TestMigrate
	if err := params.Validate(); err != nil {
		return err
	}

	mintDenomBz := cdc.MustMarshal(&gogotypes.StringValue{Value: params.MintDenom})
	enableInflationBz := cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableInflation})
	inflationDistribBz := cdc.MustMarshal(&params.InflationDistribution)
	exponentialCalcBz := cdc.MustMarshal(&params.ExponentialCalculation)

	store.Set(v2types.ParamStoreKeyMintDenom, mintDenomBz)
	store.Set(v2types.ParamStoreKeyInflationDistribution, inflationDistribBz)
	store.Set(v2types.ParamStoreKeyEnableInflation, enableInflationBz)
	store.Set(v2types.ParamStoreKeyExponentialCalculation, exponentialCalcBz)

	return nil
}
