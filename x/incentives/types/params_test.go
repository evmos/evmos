package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	epochstypes "github.com/tharsis/evmos/v4/x/epochs/types"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
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
				govtypes.DefaultPeriod,
				sdk.NewDecWithPrec(5, 2),
				epochstypes.WeekEpochID,
				sdk.NewDecWithPrec(15, 1),
			),
			false,
		},
		{
			"valid - allocation limit 100%",
			NewParams(
				true,
				govtypes.DefaultPeriod,
				sdk.NewDecWithPrec(100, 2),
				epochstypes.WeekEpochID,
				sdk.NewDecWithPrec(15, 1),
			),
			false,
		},
		{
			"valid - reward scaler 1000%",
			NewParams(
				true,
				govtypes.DefaultPeriod,
				sdk.NewDecWithPrec(100, 2),
				epochstypes.WeekEpochID,
				sdk.NewDecWithPrec(10, 0),
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
				AllocationLimit:           sdk.Dec{},
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              sdk.NewDecWithPrec(15, 1),
			},
			true,
		},
		{
			"invalid - non-positive allocation limit",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.MustNewDecFromStr("-0.02"),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              sdk.NewDecWithPrec(15, 1),
			},
			true,
		},
		{
			"invalid - allocation limit > 100%",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.NewDecWithPrec(101, 2),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              sdk.NewDecWithPrec(15, 1),
			},
			true,
		},
		{
			"invalid - nil reward scaler",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.NewDecWithPrec(5, 2),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              sdk.Dec{},
			},
			true,
		},
		{
			"invalid - non-positive reward scaler",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.NewDecWithPrec(5, 2),
				IncentivesEpochIdentifier: epochstypes.WeekEpochID,
				RewardScaler:              sdk.MustNewDecFromStr("-0.02"),
			},
			true,
		},
		{
			"invalid - empty epoch identifier",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.NewDecWithPrec(101, 2),
				IncentivesEpochIdentifier: "",
				RewardScaler:              sdk.NewDecWithPrec(15, 1),
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
