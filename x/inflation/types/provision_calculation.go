package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateEpochProvisions returns mint provision per epoch
func CalculateEpochMintProvisions(params Params, period int64) sdk.Dec {
	x := sdk.NewDec(period)                               // period
	a := params.ExponentialCalculation.A                  // initial value
	r := params.ExponentialCalculation.R                  // reduction factor
	c := params.ExponentialCalculation.C                  // long term inflation
	b := params.ExponentialCalculation.B                  // bonding factor
	epochsPerPeriod := sdk.NewDec(params.EpochsPerPeriod) //

	// exponentialDecay := a * (1 - r) ^ x + c
	decay := sdk.OneDec().Sub(r)
	exponent := sdk.OneDec().BigInt().Exp(decay.BigInt(), x.BigInt(), nil)
	exponentialDecay := a.Mul(sdk.NewDecFromBigInt(exponent)).Add(c)

	// bondingRatio := (1 + (1 - b) / 2)
	bondingRatio := (sdk.OneDec().Sub(b)).Mul(sdk.NewDecWithPrec(5, 1)).Add(sdk.OneDec())

	// periodProvision = exponentialDecau * bondingRatio
	periodProvision := exponentialDecay.Mul(bondingRatio)

	// epochProvision = periodProvision / epochsPerPeriod
	RawEpochProvision := sdk.OneDec().BigInt().Div(periodProvision.BigInt(), epochsPerPeriod.BigInt())
	epochProvision := sdk.NewDecFromBigInt(RawEpochProvision).TruncateDec()
	return epochProvision
}
