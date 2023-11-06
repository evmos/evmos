package erc20_test

import (
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"math/big"
)

func (s *PrecompileTestSuite) TestApprove() {
	method := s.precompile.Methods[authorization.ApproveMethod]
	amount := int64(100)

	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		expPass     bool
		errContains string
	}{
		{
			name:        "fail - empty args",
			malleate:    func() []interface{} { return nil },
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid number of arguments",
			malleate: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "invalid number of arguments",
		},
		{
			name: "fail - invalid address",
			malleate: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "invalid address",
		},
		{
			name: "fail - invalid amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), "invalid amount",
				}
			},
			errContains: "invalid amount",
		},
		{
			name: "fail - negative amount",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(-1),
				}
			},
			errContains: "cannot approve non-positive values",
		},
		{
			name: "pass - approve",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				// Get approvals from AuthzKeeper
				approvals, err := s.network.App.AuthzKeeper.GetAuthorizations(
					s.network.GetContext(),
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
				)
				s.Require().NoError(err, "expected no error")
				s.Require().Len(approvals, 1, "expected one approval")
				_, ok := approvals[0].(*types.SendAuthorization)
				s.Require().True(ok, "expected send authorization")
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.network.GetContext()

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(
				s.T(),
				ctx,
				s.keyring.GetAddr(0),
				s.precompile,
				200_000,
			)

			args := tc.malleate()

			bz, err := s.precompile.Approve(
				ctx,
				contract,
				s.network.GetStateDB(),
				&method,
				args,
			)

			if tc.expPass {
				s.Require().NoError(err, "expected no error")
				s.Require().NotNil(bz, "expected non-nil bytes")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				s.Require().Empty(bz, "expected empty bytes")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
