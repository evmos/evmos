package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateEpochProvisions returns mint provision per epoch
func CalculateEpochMintProvisions(params Params, period uint64) sdk.Dec {
	x := period                                           // period
	a := params.ExponentialCalculation.A                  // initial value
	r := params.ExponentialCalculation.R                  // reduction factor
	c := params.ExponentialCalculation.C                  // long term inflation
	b := params.ExponentialCalculation.B                  // bonding factor
	epochsPerPeriod := sdk.NewDec(params.EpochsPerPeriod) //

	// exponentialDecay := a * (1 - r) ^ x + c
	decay := sdk.OneDec().Sub(r)
	exponentialDecay := a.Mul(decay.Power(x)).Add(c)

	// bondingRatio := 1 + (1 - b) / 2
	bondingRatio := (sdk.OneDec().Sub(b)).Mul(sdk.NewDecWithPrec(5, 1)).Add(sdk.OneDec())

	// periodProvision = exponentialDecay * bondingRatio
	periodProvision := exponentialDecay.Mul(bondingRatio)

	// epochProvision = periodProvision / epochsPerPeriod
	decEpochProvision := sdk.OneDec().BigInt().Div(periodProvision.BigInt(), epochsPerPeriod.BigInt())
	epochProvision := sdk.NewDecFromBigInt(decEpochProvision).TruncateDec()
	return epochProvision
}
