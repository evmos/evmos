package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO move to parmas

// EpochProvision returns the provisions for a block based on the epoch
// provisions rate.
// func EpochProvision(params Params) sdk.Coin {
// 	provisionAmt := m.EpochProvisions
// 	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
// }

// CalculateEpochProvisions returns mint provision per epoch
func CalculateEpochProvisions(params Params, period int64) sdk.Dec {
	x := sdk.NewDec(period)                               // period
	a := params.ExponentialCalculation.A                  // sdk.NewDec(int64(300000000)) // initial value
	r := params.ExponentialCalculation.R                  // sdk.NewDecWithPrec(5, 1)     // 0.5 // decay factor
	c := params.ExponentialCalculation.C                  // sdk.NewDec(int64(9375000))   // long term inflation
	b := params.ExponentialCalculation.B                  // sdk.ZeroDec()                // bonding factor
	epochsPerPeriod := sdk.NewDec(params.EpochsPerPeriod) // sdk.NewDec(int64(365))

	// exponentialDecay := a * (1 - r) ^ x + c
	decay := sdk.OneDec().Sub(r)
	exponent := sdk.OneDec().BigInt().Exp(decay.BigInt(), x.BigInt(), nil)
	exponentialDecay := a.Mul(sdk.NewDecFromBigInt(exponent)).Add(c)

	// bondingRatio := (1.0 + (1.0 - b) / 2)
	bondingRatio := (sdk.OneDec().Sub(b)).Mul(sdk.NewDecWithPrec(5, 1)).Add(sdk.OneDec())

	// periodProvision = exponentialDecau * bondingRatio
	periodProvision := exponentialDecay.Mul(bondingRatio)

	// epochProvision = periodProvision / epochsPerPeriod
	RawEpochProvision := sdk.OneDec().BigInt().Div(periodProvision.BigInt(), epochsPerPeriod.BigInt())
	epochProvision := sdk.NewDecFromBigInt(RawEpochProvision).TruncateDec()
	return epochProvision
}
