package erc20_test

import (
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	commonfactory "github.com/evmos/evmos/v15/testutil/integration/common/factory"
)

// setupSendAuthz is a helper function to set up a SendAuthorization for
// a given grantee and granter combination for a given amount.
//
// NOTE: A default expiration of 1 hour after the current block time is used.
func (s *PrecompileTestSuite) setupSendAuthz(
	grantee sdk.AccAddress, granterPriv cryptotypes.PrivKey, amount sdk.Coins,
) {
	granter := sdk.AccAddress(granterPriv.PubKey().Address())
	expiration := s.network.GetContext().BlockHeader().Time.Add(time.Hour)
	sendAuthz := banktypes.NewSendAuthorization(
		amount,
		[]sdk.AccAddress{},
	)

	msgGrant, err := authz.NewMsgGrant(
		granter,
		grantee,
		sendAuthz,
		&expiration,
	)
	s.Require().NoError(err, "failed to create MsgGrant")

	// Create an authorization
	txArgs := commonfactory.CosmosTxArgs{Msgs: []sdk.Msg{msgGrant}}
	_, err = s.factory.ExecuteCosmosTx(granterPriv, txArgs)
	s.Require().NoError(err, "failed to execute MsgGrant")
}

// requireSendAuthz is a helper function to check that a SendAuthorization
// exists for a given grantee and granter combination for a given amount.
//
// NOTE: This helper expects only one authorization to exist.
func (s *PrecompileTestSuite) requireSendAuthz(grantee, granter sdk.AccAddress, amount sdk.Coins, allowList []string) {
	grants, err := s.grpcHandler.GetGrantsByGrantee(grantee.String())
	s.Require().NoError(err, "expected no error querying the grants")
	s.Require().Len(grants, 1, "expected one grant")
	s.Require().Equal(grantee.String(), grants[0].Grantee, "expected different grantee")
	s.Require().Equal(granter.String(), grants[0].Granter, "expected different granter")

	authzs, err := s.grpcHandler.GetAuthorizationsByGrantee(grantee.String())
	s.Require().NoError(err, "expected no error unpacking the authorization")
	s.Require().Len(authzs, 1, "expected one authorization")

	sendAuthz, ok := authzs[0].(*banktypes.SendAuthorization)
	s.Require().True(ok, "expected send authorization")

	spendLimits := sendAuthz.SpendLimit
	s.Require().Equal(amount, spendLimits, "expected different spend limit amount")
	if len(allowList) == 0 {
		s.Require().Empty(sendAuthz.AllowList, "expected empty allow list")
	} else {
		s.Require().Equal(allowList, sendAuthz.AllowList, "expected different allow list")
	}
}
