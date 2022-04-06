package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	devShares := sdk.NewDecWithPrec(60, 2)
	validatorShares := sdk.NewDecWithPrec(40, 2)

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid: enabled",
			NewParams(true, devShares, validatorShares),
			false,
		},
		{
			"valid: disabled",
			NewParams(false, devShares, validatorShares),
			false,
		},
		{
			"valid: 100% devs",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(1)), sdk.NewDecFromInt(sdk.NewInt(0))},
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"invalid: > 1",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(2)), sdk.NewDecFromInt(sdk.NewInt(0))},
			true,
		},
		{
			"invalid: < 0",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(-1)), sdk.NewDecFromInt(sdk.NewInt(0))},
			true,
		},
		{
			"invalid: sum > 1 ",
			Params{true, sdk.NewDecFromInt(sdk.NewInt(1)), sdk.NewDecFromInt(sdk.NewInt(1))},
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
