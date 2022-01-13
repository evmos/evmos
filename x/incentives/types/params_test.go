package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
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
				"week",
			),
			false,
		},
		{
			"valid - allocation limit 100%",
			NewParams(
				true,
				govtypes.DefaultPeriod,
				sdk.NewDecWithPrec(100, 2),
				"week",
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
				IncentivesEpochIdentifier: "week",
			},
			true,
		},
		{
			"invalid - non-positive allocation limit",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.MustNewDecFromStr("-0.02"),
				IncentivesEpochIdentifier: "week",
			},
			true,
		},
		{
			"invalid - allocation limit > 100%",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.NewDecWithPrec(101, 2),
				IncentivesEpochIdentifier: "week",
			},
			true,
		},
		{
			"invalid - empty epoch identifier",
			Params{
				EnableIncentives:          true,
				AllocationLimit:           sdk.NewDecWithPrec(101, 2),
				IncentivesEpochIdentifier: "",
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
