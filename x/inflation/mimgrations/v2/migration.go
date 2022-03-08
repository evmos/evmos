package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v2/x/inflation/types"
)

type InflationKeeper interface {
	SetParams(ctx sdk.Context, params types.Params)
}

func UpdateParams(ctx sdk.Context, k InflationKeeper) error {
	inflationParams := types.Params{
		MintDenom: "aevmos",
		ExponentialCalculation: types.ExponentialCalculation{
			A:             sdk.NewDec(int64(300_000_000)),
			R:             sdk.NewDecWithPrec(50, 2), // 50%
			C:             sdk.NewDec(int64(9_375_000)),
			BondingTarget: sdk.NewDecWithPrec(66, 2), // 66%
			MaxVariance:   sdk.ZeroDec(),             // 0%
		},
		InflationDistribution: types.InflationDistribution{
			StakingRewards:  sdk.NewDecWithPrec(533333334, 9), // 0.53 = 40% / (1 - 25%)
			UsageIncentives: sdk.NewDecWithPrec(333333333, 9), // 0.33 = 25% / (1 - 25%)
			CommunityPool:   sdk.NewDecWithPrec(133333333, 9), // 0.13 = 10% / (1 - 25%)
		},
		EnableInflation: true,
	}

	k.SetParams(ctx, inflationParams)
	return nil
}
