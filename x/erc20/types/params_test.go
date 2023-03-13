package types_test

import (
	"testing"

	"github.com/evmos/evmos/v12/x/erc20/types"
	"github.com/stretchr/testify/suite"
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
		params   types.Params
		expError bool
	}{
		{"default", types.DefaultParams(), false},
		{
			"valid",
			types.NewParams(true, true),
			false,
		},
		{
			"empty",
			types.Params{},
			false,
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
	suite.Require().Error(types.ValidateBool(1))
	suite.Require().NoError(types.ValidateBool(true))
}
