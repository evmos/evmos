package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v1types "github.com/evoblockchain/evoblock/v8/x/claims/migrations/v1/types"
	"github.com/evoblockchain/evoblock/v8/x/claims/types"
)

// MigrateStore adds the new parameters AuthorizedChannels and EVMChannels
// to the claims paramstore.
func MigrateStore(ctx sdk.Context, paramstore *paramtypes.Subspace) error {
	if !paramstore.HasKeyTable() {
		ps := paramstore.WithKeyTable(types.ParamKeyTable())
		paramstore = &ps
	}

	paramstore.Set(ctx, types.ParamStoreKeyAuthorizedChannels, types.DefaultAuthorizedChannels)
	paramstore.Set(ctx, types.ParamStoreKeyEVMChannels, types.DefaultEVMChannels)
	return nil
}

// MigrateJSON accepts exported 1 x/claims genesis state and migrates it
// to 2 x/claims genesis state. The migration includes:
// - Add AuthorizedChannels and EVMChannels
func MigrateJSON(oldState v1types.GenesisState) types.GenesisState {
	finalClaims := []types.ClaimsRecordAddress{}
	for _, claim := range oldState.ClaimsRecords {
		finalClaims = append(finalClaims, types.ClaimsRecordAddress(claim))
	}

	return types.GenesisState{
		Params: types.Params{
			EnableClaims:       oldState.Params.EnableClaims,
			ClaimsDenom:        oldState.Params.ClaimsDenom,
			AirdropStartTime:   oldState.Params.AirdropStartTime,
			DurationUntilDecay: oldState.Params.DurationUntilDecay,
			DurationOfDecay:    oldState.Params.DurationOfDecay,
			AuthorizedChannels: types.DefaultAuthorizedChannels,
			EVMChannels:        types.DefaultEVMChannels,
		},
		ClaimsRecords: finalClaims,
	}
}
