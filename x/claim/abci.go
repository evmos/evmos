package claim

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/claim/keeper"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// End Airdrop
	goneTime := ctx.BlockTime().Sub(params.AirdropStartTime)
	if goneTime > params.DurationUntilDecay+params.DurationOfDecay {
		// airdrop time has passed

		// now we sanity check that the airdrop claim hasn't already happened yet.
		// This logic is hacky, but done so due to v5 deployment timelines.
		minBalanceOsmo := sdk.NewCoin(params.ClaimDenom, sdk.NewInt(1_000_000_000_000))
		if k.GetModuleAccountBalance(ctx).IsGTE(minBalanceOsmo) {
			// airdrop not already ended
			err := k.EndAirdrop(ctx)
			if err != nil {
				panic(err)
			}
		}
	}
}
