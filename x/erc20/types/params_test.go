package types_test

import (
	"testing"

	"github.com/evmos/evmos/v18/x/erc20/types"
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
			types.NewParams(true, []string{}, []string{}),
			false,
		},
		{
			"invalid address - native precompile",
			types.NewParams(true, []string{"qq"}, []string{}),
			true,
		},
		{
			"invalid address - dynamic precompile",
			types.NewParams(true, []string{}, []string{"0xqq"}),
			true,
		},
		{
			"valid address - dynamic precompile",
			types.NewParams(true, []string{}, []string{types.WEVMOSContractMainnet}),
			false,
		},
		{
			"valid address - native precompile",
			types.NewParams(true, []string{types.WEVMOSContractMainnet}, []string{}),
			false,
		},
		{
			"repeated address in different params",
			types.NewParams(true, []string{types.WEVMOSContractMainnet}, []string{types.WEVMOSContractMainnet}),
			true,
		},
		{
			"repeated address - native precompiles",
			types.NewParams(true, []string{types.WEVMOSContractMainnet, types.WEVMOSContractMainnet}, []string{}),
			true,
		},
		{
			"repeated address - dynamic precompiles",
			types.NewParams(true, []string{}, []string{types.WEVMOSContractMainnet, types.WEVMOSContractMainnet}),
			true,
		},

		{
			"sorted address",
			// order of creation shouldnt matter since it should be sorted when defining new param
			types.NewParams(true, []string{types.WEVMOSContractTestnet, types.WEVMOSContractMainnet}, []string{}),
			false,
		},
		{
			"unsorted address",
			// order of creation shouldnt matter since it should be sorted when defining new param
			types.NewParams(true, []string{types.WEVMOSContractMainnet, types.WEVMOSContractTestnet}, []string{}),
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
