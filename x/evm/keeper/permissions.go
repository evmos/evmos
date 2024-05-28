// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"slices"

	"github.com/evmos/evmos/v18/x/evm/types"
)

type PermissionPolicy interface {
	// CanCreate checks if the contract creation is allowed.
	CanCreate(signer string, caller string) bool
	// CanCall checks if the any type of CALL opcode execution is allowed. This includes
	// contract calls and transfers.
	CanCall(signer string, caller string, recipient string) bool
}

// RestrictedPermissionPolicy is a permission policy that restricts contract creation and calls based on a set of accessControl.
type RestrictedPermissionPolicy struct {
	accessControl *types.AccessControl
	canCreate     callerFn
	canCall       callerFn
}

type callerFn = func(caller string) bool

func NewRestrictedPermissionPolicy(accessControl *types.AccessControl, signer string) RestrictedPermissionPolicy {
	// generate create function at instantiation for signer address to be check only once
	// since it remains constant
	canCreate := getCanCreateFn(accessControl, signer)
	canCall := getCanCallFn(accessControl, signer)
	return RestrictedPermissionPolicy{
		accessControl: accessControl,
		canCreate:     canCreate,
		canCall:       canCall,
	}
}

var _ PermissionPolicy = RestrictedPermissionPolicy{}

// CanCreate implements the PermissionPolicy interface.
// It allows contract creation if access type is set to everybody.
// Otherwise, it checks if:
// - The signer is allowed to do so.
// - If the signer is not allowed, then we check if the caller is allowed to do so.
func (p RestrictedPermissionPolicy) CanCreate(_, caller string) bool {
	return p.canCreate(caller)
}

func getCanCreateFn(accessControl *types.AccessControl, signer string) callerFn {
	switch accessControl.Create.AccessType {
	case types.AccessTypePermissionless:
		return func(_ string) bool { return true }
	case types.AccessTypeRestricted:
		return func(_ string) bool { return false }
	case types.AccessTypePermissioned:
		addresses := accessControl.Create.AllowlistAddresses
		return permissionedCheckFn(addresses, signer)
	}
	return func(_ string) bool { return false }
}

// CanCreate implements the PermissionPolicy interface.
// It allows calls if access type is set to everybody.
// Otherwise, it checks if:
// - The signer is allowed to do so.
// - If the signer is not allowed, then we check if the caller is allowed to do so.
func (p RestrictedPermissionPolicy) CanCall(_, caller, _ string) bool {
	return p.canCall(caller)
}

func getCanCallFn(accessControl *types.AccessControl, signer string) callerFn {
	switch accessControl.Call.AccessType {
	case types.AccessTypePermissionless:
		return func(_ string) bool { return true }
	case types.AccessTypeRestricted:
		return func(_ string) bool { return false }
	case types.AccessTypePermissioned:
		addresses := accessControl.Call.AllowlistAddresses
		return permissionedCheckFn(addresses, signer)
	}
	return func(_ string) bool { return false }
}

func permissionedCheckFn(addresses []string, signer string) callerFn {
	isSignerAllowed := slices.Contains(addresses, signer)
	return func(caller string) bool {
		return isSignerAllowed || slices.Contains(addresses, caller)
	}
}
