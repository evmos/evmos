package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/ethermint/tests"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

// TODO fix iota use
func (suite *ParamsTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	validExponentialCalculation := ExponentialCalculation{
		A: sdk.NewDec(int64(300_000_000)),
		R: sdk.NewDecWithPrec(5, 1),
		C: sdk.NewDec(int64(9_375_000)),
		B: sdk.OneDec(),
	}

	validInflationDistribution := InflationDistribution{
		StakingRewards:  sdk.NewDecWithPrec(533334, 6),
		UsageIncentives: sdk.NewDecWithPrec(333333, 6),
		CommunityPool:   sdk.NewDecWithPrec(133333, 6),
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
				"aevmos",
				"week",
				365,
				validExponentialCalculation,
				validInflationDistribution,
				tests.GenerateAddress().Hex(),
				sdk.NewDec(1_000_000),
			),
			false,
		},
		{
			"valid param literal",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				TeamAddress:            tests.GenerateAddress().Hex(),
				TeamVestingProvision:   sdk.NewDec(1_000_000),
			},
			false,
		},
		{
			"invalid - denom",
			NewParams(
				"/aevmos",
				"week",
				365,
				validExponentialCalculation,
				validInflationDistribution,
				tests.GenerateAddress().Hex(),
				sdk.NewDec(1_000_000),
			),
			true,
		},
		{
			"invalid - denom",
			Params{
				MintDenom:              "",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				TeamAddress:            tests.GenerateAddress().Hex(),
				TeamVestingProvision:   sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - wrong epochIdentifier",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				TeamAddress:            tests.GenerateAddress().Hex(),
				TeamVestingProvision:   sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - zero epochs per period ",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        0,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				TeamAddress:            tests.GenerateAddress().Hex(),
				TeamVestingProvision:   sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - exponential calculation - negative A",
			Params{
				MintDenom:       "aevmos",
				EpochIdentifier: "week",
				EpochsPerPeriod: 365,
				ExponentialCalculation: ExponentialCalculation{
					A: sdk.NewDec(int64(-1)),
					R: sdk.NewDecWithPrec(5, 1),
					C: sdk.NewDec(int64(9_375_000)),
					B: sdk.OneDec(),
				},
				InflationDistribution: validInflationDistribution,
				TeamAddress:           tests.GenerateAddress().Hex(),
				TeamVestingProvision:  sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - exponential calculation - R greater than 1",
			Params{
				MintDenom:       "aevmos",
				EpochIdentifier: "week",
				EpochsPerPeriod: 365,
				ExponentialCalculation: ExponentialCalculation{
					A: sdk.NewDec(int64(300_000_000)),
					R: sdk.NewDecWithPrec(5, 0),
					C: sdk.NewDec(int64(9_375_000)),
					B: sdk.OneDec(),
				},
				InflationDistribution: validInflationDistribution,
				TeamAddress:           tests.GenerateAddress().Hex(),
				TeamVestingProvision:  sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - exponential calculation - negative R",
			Params{
				MintDenom:       "aevmos",
				EpochIdentifier: "week",
				EpochsPerPeriod: 365,
				ExponentialCalculation: ExponentialCalculation{
					A: sdk.NewDec(int64(300_000_000)),
					R: sdk.NewDecWithPrec(-5, 1),
					C: sdk.NewDec(int64(9_375_000)),
					B: sdk.OneDec(),
				},
				InflationDistribution: validInflationDistribution,
				TeamAddress:           tests.GenerateAddress().Hex(),
				TeamVestingProvision:  sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - exponential calculation - negative C",
			Params{
				MintDenom:       "aevmos",
				EpochIdentifier: "week",
				EpochsPerPeriod: 365,
				ExponentialCalculation: ExponentialCalculation{
					A: sdk.NewDec(int64(300_000_000)),
					R: sdk.NewDecWithPrec(5, 1),
					C: sdk.NewDec(int64(-9_375_000)),
					B: sdk.OneDec(),
				},
				InflationDistribution: validInflationDistribution,
				TeamAddress:           tests.GenerateAddress().Hex(),
				TeamVestingProvision:  sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - exponential calculation - R greater than 1",
			Params{
				MintDenom:       "aevmos",
				EpochIdentifier: "week",
				EpochsPerPeriod: 365,
				ExponentialCalculation: ExponentialCalculation{
					A: sdk.NewDec(int64(300_000_000)),
					R: sdk.NewDecWithPrec(5, 0),
					C: sdk.NewDec(int64(9_375_000)),
					B: sdk.NewDec(int64(2)),
				},
				InflationDistribution: validInflationDistribution,
				TeamAddress:           tests.GenerateAddress().Hex(),
				TeamVestingProvision:  sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - exponential calculation - negative B",
			Params{
				MintDenom:       "aevmos",
				EpochIdentifier: "week",
				EpochsPerPeriod: 365,
				ExponentialCalculation: ExponentialCalculation{
					A: sdk.NewDec(int64(300_000_000)),
					R: sdk.NewDecWithPrec(5, 1),
					C: sdk.NewDec(int64(9_375_000)),
					B: sdk.OneDec().Neg(),
				},
				InflationDistribution: validInflationDistribution,
				TeamAddress:           tests.GenerateAddress().Hex(),
				TeamVestingProvision:  sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - inflation distribution - negative staking rewards",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  sdk.OneDec().Neg(),
					UsageIncentives: sdk.NewDecWithPrec(333333, 6),
					CommunityPool:   sdk.NewDecWithPrec(133333, 6),
				},
				TeamAddress:          tests.GenerateAddress().Hex(),
				TeamVestingProvision: sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - inflation distribution - negative usage incentives",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  sdk.NewDecWithPrec(533334, 6),
					UsageIncentives: sdk.OneDec().Neg(),
					CommunityPool:   sdk.NewDecWithPrec(133333, 6),
				},
				TeamAddress:          tests.GenerateAddress().Hex(),
				TeamVestingProvision: sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - inflation distribution - negative community pool rewards",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  sdk.NewDecWithPrec(533334, 6),
					UsageIncentives: sdk.NewDecWithPrec(333333, 6),
					CommunityPool:   sdk.OneDec().Neg(),
				},
				TeamAddress:          tests.GenerateAddress().Hex(),
				TeamVestingProvision: sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - inflation distribution - total distribution ratio unequal 1",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution: InflationDistribution{
					StakingRewards:  sdk.NewDecWithPrec(533333, 6),
					UsageIncentives: sdk.NewDecWithPrec(333333, 6),
					CommunityPool:   sdk.NewDecWithPrec(133333, 6),
				},
				TeamAddress:          tests.GenerateAddress().Hex(),
				TeamVestingProvision: sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - team address",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				TeamAddress:            "",
				TeamVestingProvision:   sdk.NewDec(1_000_000),
			},
			true,
		},
		{
			"invalid - negative team vesting provision",
			Params{
				MintDenom:              "aevmos",
				EpochIdentifier:        "week",
				EpochsPerPeriod:        365,
				ExponentialCalculation: validExponentialCalculation,
				InflationDistribution:  validInflationDistribution,
				TeamAddress:            tests.GenerateAddress().Hex(),
				TeamVestingProvision:   sdk.OneDec().Neg(),
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
