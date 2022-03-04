package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/tharsis/ethermint/types"
)

// CalculateEpochProvisions returns mint provision per epoch
func CalculateEpochMintProvision(
	params Params,
	period uint64,
	epochsPerPeriod int64,
	bondedRatio sdk.Dec,
) sdk.Dec {
	x := period                                              // period
	a := params.ExponentialCalculation.A                     // initial value
	r := params.ExponentialCalculation.R                     // reduction factor
	c := params.ExponentialCalculation.C                     // long term inflation
	bTarget := params.ExponentialCalculation.BondingTarget   // bonding target
	maxVariance := params.ExponentialCalculation.MaxVariance // max percentage that inflation can be increased by

	// exponentialDecay := a * (1 - r) ^ x + c
	decay := sdk.OneDec().Sub(r)
	exponentialDecay := a.Mul(decay.Power(x)).Add(c)

	// bondingIncentive doesn't increase beyond bonding target (0 < b < bonding_target)
	if bondedRatio.GTE(bTarget) {
		bondedRatio = bTarget
	}

	// bondingIncentive = 1 + max_variance - bondingRatio * (max_variance / bonding_target)
	sub := bondedRatio.Mul(maxVariance.Quo(bTarget))
	bondingIncentive := sdk.OneDec().Add(maxVariance).Sub(sub)

	// periodProvision = exponentialDecay * bondingIncentive
	periodProvision := exponentialDecay.Mul(bondingIncentive)

	// epochProvision = periodProvision / epochsPerPeriod
	epochProvision := periodProvision.Quo(sdk.NewDec(epochsPerPeriod))

	// Multiply epochMintProvision with power reduction (10^18 for evmos) as the
	// calculation is based on `evmos` and the issued tokens need to be given in
	// `aevmos`
	epochProvision = epochProvision.Mul(ethermint.PowerReduction.ToDec())
	return epochProvision
}
