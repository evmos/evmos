package types_test

import (
	"testing"

	"github.com/evmos/evmos/v19/x/erc20/types"
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
		name        string
		malleate    func() types.Params
		expError    bool
		errContains string
	}{
		{
			"default",
			types.DefaultParams,
			false,
			"",
		},
		{
			"valid",
			func() types.Params { return types.NewParams(true, []string{}, []string{}) },
			false,
			"",
		},
		{
			"valid address - dynamic precompile",
			func() types.Params { return types.NewParams(true, []string{}, []string{types.WEVMOSContractMainnet}) },
			false,
			"",
		},
		{
			"valid address - native precompile",
			func() types.Params { return types.NewParams(true, []string{types.WEVMOSContractMainnet}, []string{}) },
			false,
			"",
		},
		{
			"sorted address",
			// order of creation shouldn't matter since it should be sorted when defining new param
			func() types.Params {
				return types.NewParams(true, []string{types.WEVMOSContractTestnet, types.WEVMOSContractMainnet}, []string{})
			},
			false,
			"",
		},
		{
			"unsorted address",
			// order of creation shouldn't matter since it should be sorted when defining new param
			func() types.Params {
				return types.NewParams(true, []string{types.WEVMOSContractMainnet, types.WEVMOSContractTestnet}, []string{})
			},
			false,
			"",
		},
		{
			"empty",
			func() types.Params { return types.Params{} },
			false,
			"",
		},
		{
			"invalid address - native precompile",
			func() types.Params {
				return types.NewParams(true, []string{"qq"}, []string{})
			},
			true,
			"invalid precompile",
		},
		{
			"invalid address - dynamic precompile",
			func() types.Params {
				return types.NewParams(true, []string{}, []string{"0xqq"})
			},
			true,
			"invalid precompile",
		},
		{
			"repeated address in different params",
			func() types.Params {
				return types.NewParams(true, []string{types.WEVMOSContractMainnet}, []string{types.WEVMOSContractMainnet})
			},
			true,
			"duplicate precompile",
		},
		{
			"repeated address - native precompiles",
			func() types.Params {
				return types.NewParams(true, []string{types.WEVMOSContractMainnet, types.WEVMOSContractMainnet}, []string{})
			},
			true,
			"duplicate precompile",
		},
		{
			"repeated address - dynamic precompiles",
			func() types.Params {
				return types.NewParams(true, []string{}, []string{types.WEVMOSContractMainnet, types.WEVMOSContractMainnet})
			},
			true,
			"duplicate precompile",
		},
		{
			"unsorted addresses",
			func() types.Params {
				params := types.DefaultParams()
				params.NativePrecompiles = []string{types.WEVMOSContractTestnet, types.WEVMOSContractMainnet}
				return params
			},
			true,
			"precompiles need to be sorted",
		},
	}

	for _, tc := range testCases {
		p := tc.malleate()
		err := p.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
			suite.Require().ErrorContains(err, tc.errContains)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

func (suite *ParamsTestSuite) TestParamsValidatePriv() {
	suite.Require().Error(types.ValidateBool(1))
	suite.Require().NoError(types.ValidateBool(true))
}
