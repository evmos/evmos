// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/x/evm/keeper"
	"github.com/evmos/evmos/v18/x/evm/types"
)

func (suite *UnitTestSuite) TestAccessControl() {
	keyring := testkeyring.New(2)

	testCases := []struct {
		name             string
		getAccessControl func() types.AccessControl
		canCall          bool
		canCreate        bool
		signer           string
		caller           string
		recipient        string
	}{
		{
			name: "should allow call and create with default accessControl",
			getAccessControl: func() types.AccessControl {
				return types.DefaultParams().AccessControl
			},
			canCall:   true,
			canCreate: true,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should not allow call and create with nobody accessControl",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypeRestricted
				p.Call.AccessType = types.AccessTypeRestricted
				return p
			},
			canCall:   false,
			canCreate: false,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should not allow call with permissionless policy and signer in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Call.AccessType = types.AccessTypePermissionless
				p.Call.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
			},
			canCall:   false,
			canCreate: true,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(1).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should not allow call with permissionless policy and signer not in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Call.AccessType = types.AccessTypePermissionless
				p.Call.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
			},
			canCall:   false,
			canCreate: true,
			signer:    keyring.GetAddr(1).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should allow call with permissionless policy while caller nor signer are in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Call.AccessType = types.AccessTypePermissionless
				p.Call.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
			},
			canCall:   true,
			canCreate: true,
			signer:    keyring.GetAddr(1).String(),
			caller:    keyring.GetAddr(1).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should allow call with permissionless policy and caller not in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Call.AccessType = types.AccessTypePermissionless
				p.Call.AccessControlList = []string{keyring.GetAddr(1).String()}
				return p
			},
			canCall:   false,
			canCreate: true,
			signer:    keyring.GetAddr(1).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should not allow create with permissionless policy and signer in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypePermissionless
				p.Create.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
			},
			canCall:   true,
			canCreate: false,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(1).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should not allow create with permissionless policy and signer not in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypePermissionless
				p.Create.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
			},
			canCall:   true,
			canCreate: false,
			signer:    keyring.GetAddr(1).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should allow create with permissionless policy while caller nor signer are in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypePermissionless
				p.Create.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
			},
			canCall:   true,
			canCreate: true,
			signer:    keyring.GetAddr(1).String(),
			caller:    keyring.GetAddr(1).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should allow create with permissionless policy and caller not in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypePermissionless
				p.Create.AccessControlList = []string{keyring.GetAddr(1).String()}
				return p
			},
			canCall:   true,
			canCreate: false,
			signer:    keyring.GetAddr(1).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(1).String(),
		},
		{
			name: "should not allow call with permissioned policy and not in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Call.AccessType = types.AccessTypePermissioned
				p.Call.AccessControlList = []string{keyring.GetAddr(1).String()}
				return p
			},
			canCall:   false,
			canCreate: true,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should not allow create with permissioned policy and not in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypePermissioned
				p.Create.AccessControlList = []string{keyring.GetAddr(1).String()}
				return p
			},
			canCall:   true,
			canCreate: false,
			signer:    keyring.GetAddr(0).String(),
			caller:    keyring.GetAddr(0).String(),
			recipient: keyring.GetAddr(0).String(),
		},
		{
			name: "should allow call and create with permissioned policy and address in AccessControlList",
			getAccessControl: func() types.AccessControl {
				p := types.DefaultParams().AccessControl
				p.Create.AccessType = types.AccessTypePermissioned
				p.Create.AccessControlList = []string{keyring.GetAddr(0).String()}
				p.Call.AccessType = types.AccessTypePermissioned
				p.Call.AccessControlList = []string{keyring.GetAddr(0).String()}
				return p
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
			accessControl := tc.getAccessControl()
			permissionPolicy := keeper.NewRestrictedPermissionPolicy(
				&accessControl,
				tc.signer,
			)

			canCreate := permissionPolicy.CanCreate(tc.signer, tc.caller)
			suite.Require().Equal(tc.canCreate, canCreate, "expected %v, got %v", tc.canCreate, canCreate)

			canCall := permissionPolicy.CanCall(tc.signer, tc.caller, tc.recipient)
			suite.Require().Equal(tc.canCall, canCall, "expected %v, got %v", tc.canCall, canCall)
		})
	}
}
