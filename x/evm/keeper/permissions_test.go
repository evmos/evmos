package keeper_test

import (
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/x/evm/keeper"
	"github.com/evmos/evmos/v18/x/evm/types"
)

func (suite *UnitTestSuite) TestPermissions() {
	keyring := testkeyring.New(2)

	testCases := []struct {
		name           string
		getPermissions func() types.Permissions
		canCall        bool
		canCreate      bool
		signer         string
		caller         string
		recipient      string
	}{
		{
			name: "should allow call and create with default permissions",
			getPermissions: func() types.Permissions {
				return *types.DefaultParams().PermissionsPolicy
			},
			canCall:   true,
			canCreate: true,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should not allow call and create with nobody permissions",
			getPermissions: func() types.Permissions {
				p := types.DefaultParams().PermissionsPolicy
				p.Create.AccessType = types.AccessTypeNobody
				p.Call.AccessType = types.AccessTypeNobody
				return *p
			},
			canCall:   false,
			canCreate: false,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should not allow call with whitelisted policy and not in whitelist",
			getPermissions: func() types.Permissions {
				p := types.DefaultParams().PermissionsPolicy
				p.Call.AccessType = types.AccessTypeWhitelistAddress
				p.Call.WhitelistAddresses = []string{keyring.GetAddr(1).String()}
				return *p
			},
			canCall:   false,
			canCreate: true,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should not allow create with whitelisted policy and not in whitelist",
			getPermissions: func() types.Permissions {
				p := types.DefaultParams().PermissionsPolicy
				p.Create.AccessType = types.AccessTypeWhitelistAddress
				p.Create.WhitelistAddresses = []string{keyring.GetAddr(1).String()}
				return *p
			},
			canCall:   true,
			canCreate: false,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should allow call and create with whitelisted policy and address in whitelist",
			getPermissions: func() types.Permissions {
				p := types.DefaultParams().PermissionsPolicy
				p.Create.AccessType = types.AccessTypeWhitelistAddress
				p.Create.WhitelistAddresses = []string{keyring.GetAddr(0).String()}
				p.Call.AccessType = types.AccessTypeWhitelistAddress
				p.Call.WhitelistAddresses = []string{keyring.GetAddr(0).String()}
				return *p
			},
			canCall:   true,
			canCreate: true,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			permissions := tc.getPermissions()
			permissionPolicy := keeper.NewRestrictedPermissionPolicy(
				&permissions,
				tc.signer,
			)

			canCreate := permissionPolicy.CanCreate(tc.signer, tc.caller)
			suite.Require().Equal(tc.canCreate, canCreate, "expected %v, got %v", tc.canCreate, canCreate)

			canCall := permissionPolicy.CanCall(tc.signer, tc.caller, tc.recipient)
			suite.Require().Equal(tc.canCall, canCall, "expected %v, got %v", tc.canCall, canCall)
		})
	}
}
