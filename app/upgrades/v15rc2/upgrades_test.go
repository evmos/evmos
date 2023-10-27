// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v15rc2_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v15/app/upgrades/v15rc2"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
)

func (s *UpgradesTestSuite) TestRemoveDistributionAuthorizations() {
	granter, _ := testutiltx.NewAccAddressAndKey()
	grantee, _ := testutiltx.NewAccAddressAndKey()
	expiration := s.ctx.BlockTime().Add(10 * time.Minute).UTC()
	otherAuthz, err := stakingtypes.NewStakeAuthorization(
		[]sdk.ValAddress{s.validators[0].GetOperator()},
		nil,
		stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
		&sdk.Coin{Denom: s.bondDenom, Amount: sdk.NewInt(1)},
	)
	s.Require().NoError(err, "failed to create staking authorization")

	distAuthz := &distributiontypes.DistributionAuthorization{
		MessageType: distributiontypes.SetWithdrawerAddressMsg,
	}

	testcases := []struct {
		name     string
		malleate func()
	}{
		{
			name: "no grants",
		},
		{
			name: "grant with no distribution authorization",
			malleate: func() {
				s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, otherAuthz, &expiration))
			},
		},
		{
			name: "grant with distribution authorization",
			malleate: func() {
				s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, otherAuthz, &expiration))
				s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, distAuthz, &expiration))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.malleate != nil {
				tc.malleate()
			}

			v15rc2.RemoveDistributionAuthorizations(s.ctx, s.app.AuthzKeeper)

			// Check that there are no more distribution authorizations.
			nDistAuthz := 0
			s.app.AuthzKeeper.IterateGrants(s.ctx, func(_, _ sdk.AccAddress, grant authz.Grant) bool {
				authorization, err := grant.GetAuthorization()
				s.Require().NoError(err)

				if _, ok := authorization.(*distributiontypes.DistributionAuthorization); ok {
					nDistAuthz++
				}

				return false
			})

			s.Require().Equal(0, nDistAuthz, "expected no distribution authorizations")
		})
	}
}
