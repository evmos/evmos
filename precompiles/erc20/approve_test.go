package erc20_test

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/encoding"
	"github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	commonfactory "github.com/evmos/evmos/v15/testutil/integration/common/factory"
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
				// Get approvals from AuthzKeeper
				approvals, err := s.network.App.AuthzKeeper.GetAuthorizations(
					s.network.GetContext(),
					s.keyring.GetAccAddr(1),
					s.keyring.GetAccAddr(0),
				)
				s.Require().NoError(err, "expected no error")
				s.Require().Len(approvals, 1, "expected one approval")
				_, ok := approvals[0].(*banktypes.SendAuthorization)
				s.Require().True(ok, "expected send authorization")
			},
		},
		{
			name: "pass - approve with existing authorization",
			malleate: func() []interface{} {
				// TODO: refactor into integration test suite
				sendAuthz := banktypes.NewSendAuthorization(
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 1)),
					[]sdk.AccAddress{},
				)

				expiration := s.network.GetContext().BlockHeader().Time.Add(time.Hour)

				msgGrant, err := authz.NewMsgGrant(
					s.keyring.GetAccAddr(0),
					s.keyring.GetAccAddr(1),
					sendAuthz,
					&expiration,
				)
				s.Require().NoError(err, "expected no error creating the MsgGrant")

				// Create an authorization
				txArgs := commonfactory.CosmosTxArgs{Msgs: []sdk.Msg{msgGrant}}
				_, err = s.factory.ExecuteCosmosTx(s.keyring.GetPrivKey(0), txArgs)
				s.Require().NoError(err, "expected no error executing the MsgGrant tx")

				return []interface{}{
					s.keyring.GetAddr(1), big.NewInt(2 * amount),
				}
			},
			expPass: true,
			postCheck: func() {
				// Get approvals from Authz client
				authzClient := s.network.GetAuthzClient()
				req := &authz.QueryGranteeGrantsRequest{Grantee: s.keyring.GetAccAddr(1).String()}
				res, err := authzClient.GranteeGrants(s.network.GetContext(), req)
				s.Require().NoError(err, "expected no error querying the grants")
				s.Require().Len(res.Grants, 1, "expected one grant")

				encodingCfg := encoding.MakeConfig(app.ModuleBasics)
				var authz authz.Authorization
				err = encodingCfg.Codec.UnpackAny(res.Grants[0].Authorization, &authz)
				s.Require().NoError(err, "expected no error unpacking the authorization")
				sendAuthz, ok := authz.(*banktypes.SendAuthorization)
				s.Require().True(ok, "expected send authorization")

				// Check that the authorization has the correct amount
				spendLimits := sendAuthz.SpendLimit
				s.Require().Len(spendLimits, 1, "expected spend limit in one denomination")
				s.Require().Equal(2*amount, spendLimits[0].Amount.Int64(), "expected correct amount")
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
