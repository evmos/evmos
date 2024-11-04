package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	validExponentialCalculation := ExponentialCalculation{
		A:             math.LegacyNewDec(int64(300_000_000)),
		R:             math.LegacyNewDecWithPrec(5, 1),
		C:             math.LegacyNewDec(int64(9_375_000)),
		BondingTarget: math.LegacyNewDecWithPrec(50, 2),
		MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
	}

	validInflationDistribution := InflationDistribution{
		StakingRewards:  math.LegacyNewDecWithPrec(533334, 6),
		UsageIncentives: math.LegacyZeroDec(),
		CommunityPool:   math.LegacyNewDecWithPrec(466666, 6),
	}

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"default",
			DefaultParams(),
			false,
		},
		{
			"valid",
			NewParams(
				DefaultInflationDenom,
				validExponentialCalculation,
				validInflationDistribution,
				true,
			),
			false,
		},
		{
			"valid param literal",
			Params{
				MintDenom:              DefaultInflationDenom,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				EnableInflation:        true,
			},
			false,
		},
		{
			"invalid - denom",
			NewParams(
				"/aevmos",
				validExponentialCalculation,
				validInflationDistribution,
				true,
			),
			true,
		},
		{
			"invalid - denom",
			Params{
				MintDenom:              "",
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				EnableInflation:        true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative A",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(-1)),
					R:             math.LegacyNewDecWithPrec(5, 1),
					C:             math.LegacyNewDec(int64(9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - R greater than 1",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(5, 0),
					C:             math.LegacyNewDec(int64(9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative R",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(-5, 1),
					C:             math.LegacyNewDec(int64(9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative C",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(5, 1),
					C:             math.LegacyNewDec(int64(-9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - BondingTarget greater than 1",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(5, 1),
					C:             math.LegacyNewDec(int64(9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 1),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative BondingTarget",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(5, 1),
					C:             math.LegacyNewDec(int64(9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 2).Neg(),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - exponential calculation - negative max Variance",
			Params{
				MintDenom: DefaultInflationDenom,
				ExponentialCalculation: ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(5, 1),
					C:             math.LegacyNewDec(int64(9_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(50, 2),
					MaxVariance:   math.LegacyNewDecWithPrec(20, 2).Neg(),
				},
				InflationDistribution: validInflationDistribution,
				EnableInflation:       true,
			},
			true,
		},
		{
			"invalid - inflation distribution - negative staking rewards",
			Params{
				MintDenom:              DefaultInflationDenom,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  math.LegacyOneDec().Neg(),
					UsageIncentives: math.LegacyNewDecWithPrec(333333, 6),
					CommunityPool:   math.LegacyNewDecWithPrec(133333, 6),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation distribution - negative usage incentives",
			Params{
				MintDenom:              DefaultInflationDenom,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  math.LegacyNewDecWithPrec(533334, 6),
					UsageIncentives: math.LegacyOneDec().Neg(),
					CommunityPool:   math.LegacyNewDecWithPrec(133333, 6),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation distribution - negative community pool rewards",
			Params{
				MintDenom:              DefaultInflationDenom,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  math.LegacyNewDecWithPrec(533334, 6),
					UsageIncentives: math.LegacyNewDecWithPrec(333333, 6),
					CommunityPool:   math.LegacyOneDec().Neg(),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation distribution - total distribution ratio unequal 1",
			Params{
				MintDenom:              DefaultInflationDenom,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  math.LegacyNewDecWithPrec(533333, 6),
					UsageIncentives: math.LegacyNewDecWithPrec(333333, 6),
					CommunityPool:   math.LegacyNewDecWithPrec(133333, 6),
				},
				EnableInflation: true,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}
