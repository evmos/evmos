package erc20_test

import (
	"math/big"

	"github.com/evmos/evmos/v16/precompiles/erc20"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
)

//nolint:dupl // these tests are not duplicates
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
		{
			name: "fail - invalid number of arguments",
			args: []interface{}{
				1, 2, 3,
			},
			errContains: "invalid number of arguments",
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

func (s *PrecompileTestSuite) TestParseTransferFromArgs() {
	from := utiltx.GenerateAddress()
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
				from,
				to,
				amount,
			},
			expPass: true,
		},
		{
			name: "fail - invalid from address",
			args: []interface{}{
				"invalid address",
				to,
				amount,
			},
			errContains: "invalid from address",
		},
		{
			name: "fail - invalid to address",
			args: []interface{}{
				from,
				"invalid address",
				amount,
			},
			errContains: "invalid to address",
		},
		{
			name: "fail - invalid amount",
			args: []interface{}{
				from,
				to,
				"invalid amount",
			},
			errContains: "invalid amount",
		},
		{
			name: "fail - invalid number of arguments",
			args: []interface{}{
				1, 2, 3, 4,
			},
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			from, to, amount, err := erc20.ParseTransferFromArgs(tc.args)
			if tc.expPass {
				s.Require().NoError(err, "unexpected error parsing the transferFrom arguments")
				s.Require().Equal(from, tc.args[0], "expected different from address")
				s.Require().Equal(to, tc.args[1], "expected different to address")
				s.Require().Equal(amount, tc.args[2], "expected different amount")
			} else {
				s.Require().Error(err, "expected an error parsing the transferFrom arguments")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
			}
		})
	}
}

//nolint:dupl // these tests are not duplicates
func (s *PrecompileTestSuite) TestParseApproveArgs() {
	spender := utiltx.GenerateAddress()
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
				spender,
				amount,
			},
			expPass: true,
		},
		{
			name: "fail - invalid spender address",
			args: []interface{}{
				"invalid address",
				amount,
			},
			errContains: "invalid spender address",
		},
		{
			name: "fail - invalid amount",
			args: []interface{}{
				spender,
				"invalid amount",
			},
			errContains: "invalid amount",
		},
		{
			name: "fail - invalid number of arguments",
			args: []interface{}{
				1, 2, 3,
			},
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			spender, amount, err := erc20.ParseApproveArgs(tc.args)
			if tc.expPass {
				s.Require().NoError(err, "unexpected error parsing the approve arguments")
				s.Require().Equal(spender, tc.args[0], "expected different spender address")
				s.Require().Equal(amount, tc.args[1], "expected different amount")
			} else {
				s.Require().Error(err, "expected an error parsing the approve arguments")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
			}
		})
	}
}

func (s *PrecompileTestSuite) TestParseAllowanceArgs() {
	owner := utiltx.GenerateAddress()
	spender := utiltx.GenerateAddress()

	testcases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			name: "pass - correct arguments",
			args: []interface{}{
				owner,
				spender,
			},
			expPass: true,
		},
		{
			name: "fail - invalid owner address",
			args: []interface{}{
				"invalid address",
				spender,
			},
			errContains: "invalid owner address",
		},
		{
			name: "fail - invalid spender address",
			args: []interface{}{
				owner,
				"invalid address",
			},
			errContains: "invalid spender address",
		},
		{
			name: "fail - invalid number of arguments",
			args: []interface{}{
				1, 2, 3,
			},
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			owner, spender, err := erc20.ParseAllowanceArgs(tc.args)
			if tc.expPass {
				s.Require().NoError(err, "unexpected error parsing the allowance arguments")
				s.Require().Equal(owner, tc.args[0], "expected different owner address")
				s.Require().Equal(spender, tc.args[1], "expected different spender address")
			} else {
				s.Require().Error(err, "expected an error parsing the allowance arguments")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
			}
		})
	}
}

func (s *PrecompileTestSuite) TestParseBalanceOfArgs() {
	account := utiltx.GenerateAddress()

	testcases := []struct {
		name        string
		args        []interface{}
		expPass     bool
		errContains string
	}{
		{
			name: "pass - correct arguments",
			args: []interface{}{
				account,
			},
			expPass: true,
		},
		{
			name: "fail - invalid account address",
			args: []interface{}{
				"invalid address",
			},
			errContains: "invalid account address",
		},
		{
			name: "fail - invalid number of arguments",
			args: []interface{}{
				1, 2, 3,
			},
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			account, err := erc20.ParseBalanceOfArgs(tc.args)
			if tc.expPass {
				s.Require().NoError(err, "unexpected error parsing the balanceOf arguments")
				s.Require().Equal(account, tc.args[0], "expected different account address")
			} else {
				s.Require().Error(err, "expected an error parsing the balanceOf arguments")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
			}
		})
	}
}
