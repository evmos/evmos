package erc20_test

import (
	"math/big"

	"github.com/evmos/evmos/v15/precompiles/erc20"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
)

func (s *PrecompileTestSuite) TestParseTransferArgs() {
	to := utiltx.GenerateAddress()
	amount := big.NewInt(100)

	testcases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			name: "pass - correct arguments",
			args: []interface{}{
				to,
				amount,
			},
			expPass: true,
		},
		{
			name: "fail - invalid to address",
			args: []interface{}{
				"invalid address",
				amount,
			},
			errContains: "invalid to address",
		},
		{
			name: "fail - invalid amount",
			args: []interface{}{
				to,
				"invalid amount",
			},
			errContains: "invalid amount",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			to, amount, err := erc20.ParseTransferArgs(tc.args)
			if tc.expPass {
				s.Require().NoError(err, "unexpected error parsing the transfer arguments")
				s.Require().Equal(to, tc.args[0], "expected different to address")
				s.Require().Equal(amount, tc.args[1], "expected different amount")
			} else {
				s.Require().Error(err, "expected an error parsing the transfer arguments")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
			}
		})
	}
}
