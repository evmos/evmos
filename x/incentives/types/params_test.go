package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	epochstypes "github.com/evmos/evmos/v15/x/epochs/types"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamsValidate() {
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
			"valid - allocation limit 5%",
			NewParams(
				true,
				math.LegacyNewDecWithPrec(5, 2),
				epochstypes.WeekEpochID,
				math.LegacyNewDecWithPrec(15, 1),
			),
			false,
		},
		{
			"valid - allocation limit 100%",
			NewParams(
				true,
				math.LegacyNewDecWithPrec(100, 2),
				epochstypes.WeekEpochID,
				math.LegacyNewDecWithPrec(15, 1),
			),
			false,
		},
		{
			"valid - reward scaler 1000%",
			NewParams(
				true,
				math.LegacyNewDecWithPrec(100, 2),
				epochstypes.WeekEpochID,
				math.LegacyNewDecWithPrec(10, 0),
			),
			false,
		},
		{
			"invalid - empty Params",
			Params{},
			true,
		},
		{
			"invalid - nil allocation limit",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           math.LegacyDec{},
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              math.LegacyNewDecWithPrec(15, 1),
			},
			true,
		},
		{
			"invalid - non-positive allocation limit",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           math.LegacyMustNewDecFromStr("-0.02"),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              math.LegacyNewDecWithPrec(15, 1),
			},
			true,
		},
		{
			"invalid - allocation limit > 100%",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           math.LegacyNewDecWithPrec(101, 2),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              math.LegacyNewDecWithPrec(15, 1),
			},
			true,
		},
		{
			"invalid - nil reward scaler",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           math.LegacyNewDecWithPrec(5, 2),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              math.LegacyDec{},
			},
			true,
		},
		{
			"invalid - non-positive reward scaler",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           math.LegacyNewDecWithPrec(5, 2),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              math.LegacyMustNewDecFromStr("-0.02"),
			},
			true,
		},
		{
			"invalid - empty epoch identifier",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           math.LegacyNewDecWithPrec(101, 2),
				IncentivesEpochIdentifier: "",
				RewardScaler:              math.LegacyNewDecWithPrec(15, 1),
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

func (suite *ParamsTestSuite) TestParamsValidatePriv() {
	suite.Require().Error(validateBool(1))
	suite.Require().NoError(validateBool(true))
}
