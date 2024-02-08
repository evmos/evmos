// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"cosmossdk.io/math"

	evmostypes "github.com/evmos/evmos/v16/types"
)

const (
	// ReductionFactor is the value used as denominator to divide the provision amount computed
	// with the CalculateEpochMintProvision function.
	ReductionFactor = 3
)

// CalculateEpochProvisions returns mint provision per epoch. The function used to compute the
// emission is the half life times a reduction factor:
//
// f(x) = { a * (1 -r ) ^ x * [1 + maxVariance * (1 - bondedRatio / bTarget)] + c} / reductionFactor
//
// where x represents years. This means that the equation is computed fixing the number of the
// current year with respect to the starting year. Then, f(x) is computed and from this, the epoch
// emission. For example, having x=0, the tokens minted for a specific epochs are proportional to
// f(0) / numberOfEpochs.
func CalculateEpochMintProvision(
	params Params,
	period uint64,
	epochsPerPeriod int64,
	bondedRatio math.LegacyDec,
) math.LegacyDec {
	x := period                                              // period
	a := params.ExponentialCalculation.A                     // initial value
	r := params.ExponentialCalculation.R                     // reduction factor
	c := params.ExponentialCalculation.C                     // long term inflation
	bTarget := params.ExponentialCalculation.BondingTarget   // bonding target
	maxVariance := params.ExponentialCalculation.MaxVariance // max percentage that inflation can be increased by

	// exponentialDecay := a * (1 - r) ^ x
	decay := math.LegacyOneDec().Sub(r)
	exponentialDecay := a.Mul(decay.Power(x))

	// bondingIncentive doesn't increase beyond bonding target (0 < b < bonding_target)
	if bondedRatio.GTE(bTarget) {
		bondedRatio = bTarget
	}

	// bondingIncentive = 1 + max_variance - max_variance * (bondingRatio / bonding_target)
	var bondingIncentive math.LegacyDec
	if maxVariance != math.LegacyZeroDec() {
		sub := maxVariance.Mul(bondedRatio.Quo(bTarget))
		bondingIncentive = math.LegacyOneDec().Add(maxVariance).Sub(sub)
	} else {
		bondingIncentive = math.LegacyOneDec()
	}

	// reducedPeriodProvision = (exponentialDecay * bondingIncentive + c) / reductionFactor
	periodProvision := exponentialDecay.Mul(bondingIncentive).Add(c)
	reducedPeriodProvision := periodProvision.Quo(math.LegacyNewDec(ReductionFactor))

	// epochProvision = periodProvision / epochsPerPeriod
	epochProvision := reducedPeriodProvision.Quo(math.LegacyNewDec(epochsPerPeriod))

	// Multiply epochMintProvision with power reduction (10^18 for evmos) as the
	// calculation is based on `evmos` and the issued tokens need to be given in
	// `aevmos`
	epochProvision = epochProvision.Mul(math.LegacyNewDecFromInt(evmostypes.PowerReduction))
	return epochProvision
}
