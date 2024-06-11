package tokenfactory_test

import (
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v18/precompiles/common"
	tokenfactory "github.com/evmos/evmos/v18/precompiles/token_factory"
)

func (s *PrecompileTestSuite) TestParseCreateERC20Args() {
	testCases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			"fail - invalid number of arguments",
			[]interface{}{
				1, 2, 3,
			},
			false,
			fmt.Sprintf(common.ErrInvalidNumberOfArgs, 4, 3),
		},
		{
			"fail - invalid name argument type",
			[]interface{}{
				1, 2, 3, 4,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidNameArgumentType, 1),
		},
		{
			"fail - invalid symbol argument type",
			[]interface{}{
				"TEST", 2, 3, 4,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidSymbolArgumentType, 1),
		},
		{
			"fail - invalid decimal argument type",
			[]interface{}{
				"TEST", "TST", "3", 4,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidDecimalArgumentType, "3"),
		},
		{
			"fail - invalid initial supply argument type",
			[]interface{}{
				"TEST", "TST", uint8(3), "4",
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidInitialSupplyArgumentType, "4"),
		},
		{
			"pass - correct arguments",
			[]interface{}{
				"TEST", "TST", uint8(18), big.NewInt(1),
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, _, _, _, err := tokenfactory.ParseCreateErc20Args(tc.args)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestParseCreate2ERC20Args() {
	testCases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			"fail - invalid number of arguments",
			[]interface{}{
				1, 2, 3,
			},
			false,
			fmt.Sprintf(common.ErrInvalidNumberOfArgs, 5, 3),
		},
		{
			"fail - invalid name argument type",
			[]interface{}{
				1, 2, 3, 4, 5,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidNameArgumentType, 1),
		},
		{
			"fail - invalid symbol argument type",
			[]interface{}{
				"TEST", 2, 3, 4, 5,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidSymbolArgumentType, 1),
		},
		{
			"fail - invalid decimal argument type",
			[]interface{}{
				"TEST", "TST", "3", 4, 5,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidDecimalArgumentType, "3"),
		},
		{
			"fail - invalid initial supply argument type",
			[]interface{}{
				"TEST", "TST", uint8(3), "4", 5,
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidInitialSupplyArgumentType, "4"),
		},
		{
			"fail - invalid salt argument type",
			[]interface{}{
				"TEST", "TST", uint8(18), big.NewInt(1), "5",
			},
			false,
			fmt.Sprintf(tokenfactory.ErrInvalidSaltArgumentType, "4"),
		},
		{
			"pass - correct arguments",
			[]interface{}{
				"TEST", "TST", uint8(18), big.NewInt(1), [32]byte{},
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, _, _, _, _, err := tokenfactory.ParseCreate2Erc20Args(tc.args)
			if tc.expPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}
