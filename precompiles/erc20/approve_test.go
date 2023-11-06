package erc20_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
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
			name: "pass - approve without existing authorization",
			malleate: func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, amount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - approve with existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(amount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, 1)),
					[]string{s.tokenDenom},
				)
			},
		},
		{
			name: "pass - approve with existing authorization in different denomination",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(2 * amount),
				}
			},
			expPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
					// NOTE: The approval in the different denomination is overwritten by the
					// approval for the passed token denomination.
					//
					// TODO: check if this behaviour is the same for ERC20s? Or can there be separate
					// approvals for different denominations?
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, 2*amount)),
					[]string{},
				)
			},
		},
		{
			name: "pass - delete existing authorization",
			malleate: func() []interface{} {
				s.setupSendAuthz(
					s.keyring.GetAccAddr(1),
					s.keyring.GetPrivKey(0),
					sdk.NewCoins(sdk.NewInt64Coin(s.tokenDenom, 1)),
				)

				return []interface{}{
					s.keyring.GetAddr(1), common.Big0,
				}
			},
			expPass: true,
			postCheck: func() {
				grants, err := s.grpcHandler.GetGrantsByGrantee(s.keyring.GetAccAddr(1).String())
				s.Require().NoError(err, "expected no error querying the grants")
				s.Require().Len(grants, 0, "expected grant to be deleted")
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

			var args []interface{}
			if tc.malleate != nil {
				args = tc.malleate()
			}

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
