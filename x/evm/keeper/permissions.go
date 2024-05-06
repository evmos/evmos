// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"slices"

	"github.com/evmos/evmos/v18/x/evm/types"
)

type PermissionPolicy interface {
	CanCreate(signer string, caller string) bool
	CanCall(signer string, caller string, recipient string) bool
}

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
// It allows contract creation if:
// - The signer is allowed to do so.
// - If the signer is not allowed, then we check if the caller is allowed to do so.
func (p RestrictedPermissionPolicy) CanCreate(_, caller string) bool {
	return p.canCreate(caller)
}

func getCreateCallerFn(permissions *types.Permissions, signer string) callerFn {
	switch permissions.Create.AccessType {
	case types.AccessTypeEverybody:
		return func(caller string) bool { return true }
	case types.AccessTypeNobody:
		return func(caller string) bool { return false }
	case types.AccessTypeWhitelistAddress:
		addresses := permissions.Create.WhitelistAddresses
		isSignerAllowed := slices.Contains(addresses, signer)
		return func(caller string) bool {
			return isSignerAllowed || slices.Contains(addresses, caller)
		}
	}
	return func(caller string) bool { return false }
}

func (p RestrictedPermissionPolicy) CanCall(_, caller, _ string) bool {
	return p.canCall(caller)
}

func generateCallerFn(permissions *types.Permissions, signer string) callerFn {
	switch permissions.Call.AccessType {
	case types.AccessTypeEverybody:
		return func(caller string) bool { return true }
	case types.AccessTypeNobody:
		return func(caller string) bool { return false }
	case types.AccessTypeWhitelistAddress:
		addresses := permissions.Call.WhitelistAddresses
		isSignerAllowed := slices.Contains(addresses, signer)
		return func(caller string) bool {
			return isSignerAllowed || slices.Contains(addresses, caller)
		}
	}
	return func(caller string) bool { return false }
}
