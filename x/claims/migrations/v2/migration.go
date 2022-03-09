package v2

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

type ClaimsKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)
}

func UpdateParams(ctx sdk.Context, k ClaimsKeeper) error {
	claimsParams := types.Params{
		EnableClaims:       true,
		AirdropStartTime:   time.Date(2022, time.March, 3, 18, 0, 0, 0, time.UTC),
		DurationUntilDecay: (2592000 * time.Second) + (time.Hour * 24 * 14), // add 2 weeks
		DurationOfDecay:    5184000 * time.Second,
		ClaimsDenom:        "aevmos",
		AuthorizedChannels: types.DefaultAuthorizedChannels,
		EVMChannels:        types.DefaultEVMChannels,
	}
	k.SetParams(ctx, claimsParams)
	return nil
}
