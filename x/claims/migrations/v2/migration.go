package v2

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/x/claims/types"
)

type ClaimsKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)
}

func UpdateParams(ctx sdk.Context, k ClaimsKeeper) error {
	claimsParams := k.GetParams(ctx)
	claimsParams.DurationUntilDecay += time.Hour * 24 * 14 // add 2 weeks
	// TODO: add for v2
	// claimsParams.AuthorizedChannels = claimstypes.DefaultAuthorizedChannels
	// claimsParams.EVMChannels = claimstypes.DefaultEVMChannels
	k.SetParams(ctx, claimsParams)
	return nil
}
