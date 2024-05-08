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

// RestrictedPermissionPolicy is a permission policy that restricts contract creation and calls based on a set of permissions.
type RestrictedPermissionPolicy struct {
	permissions *types.Permissions
	canCreate   callerFn
	canCall     callerFn
}

type callerFn = func(caller string) bool

func NewRestrictedPermissionPolicy(permissions *types.Permissions, signer string) RestrictedPermissionPolicy {
	// generate create function at instantiation for signer address to be check only once
	// since it remains constant
	canCreate := getCreateCallerFn(permissions, signer)
	canCall := generateCallerFn(permissions, signer)
	return RestrictedPermissionPolicy{
		permissions: permissions,
		canCreate:   canCreate,
		canCall:     canCall,
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

func getCreateCallerFn(permissions *types.Permissions, signer string) callerFn {
	switch permissions.Create.AccessType {
	case types.AccessTypeEverybody:
		return func(_ string) bool { return true }
	case types.AccessTypeNobody:
		return func(_ string) bool { return false }
	case types.AccessTypeWhitelistAddress:
		addresses := permissions.Create.WhitelistAddresses
		isSignerAllowed := slices.Contains(addresses, signer)
		return func(caller string) bool {
			return isSignerAllowed || slices.Contains(addresses, caller)
		}
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

func generateCallerFn(permissions *types.Permissions, signer string) callerFn {
	switch permissions.Call.AccessType {
	case types.AccessTypeEverybody:
		return func(_ string) bool { return true }
	case types.AccessTypeNobody:
		return func(_ string) bool { return false }
	case types.AccessTypeWhitelistAddress:
		addresses := permissions.Call.WhitelistAddresses
		isSignerAllowed := slices.Contains(addresses, signer)
		return func(caller string) bool {
			return isSignerAllowed || slices.Contains(addresses, caller)
		}
	}
	return func(_ string) bool { return false }
}
