package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateEpochProvisions returns mint provision per epoch
func CalculateEpochMintProvision(params Params, period uint64, epochsPerPeriod int64) sdk.Dec {
	x := period                          // period
	a := params.ExponentialCalculation.A // initial value
	r := params.ExponentialCalculation.R // reduction factor
	c := params.ExponentialCalculation.C // long term inflation
	b := params.ExponentialCalculation.B // bonding factor

	// exponentialDecay := a * (1 - r) ^ x + c
	decay := sdk.OneDec().Sub(r)
	exponentialDecay := a.Mul(decay.Power(x)).Add(c)

	// bondingRatio := (2 - b) / 2
	bondingRatio := (sdk.NewDec(2).Sub(b)).Mul(sdk.NewDecWithPrec(5, 1))

	// periodProvision = exponentialDecay * bondingRatio
	periodProvision := exponentialDecay.Mul(bondingRatio)

	// epochProvision = periodProvision / epochsPerPeriod
	decEpochProvision := sdk.OneDec().BigInt().Quo(
		periodProvision.BigInt(),
		sdk.NewDec(epochsPerPeriod).BigInt(),
	)
	epochProvision := sdk.NewDecFromBigInt(decEpochProvision)
	return epochProvision
}
