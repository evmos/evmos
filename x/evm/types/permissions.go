// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/x/evm/core/vm"
)

// PermissionPolicy is the interface that defines the permission policy for contract creation and calls.
// It is used to enforce access control policies on EVM operations.
// The policy is ran BEFORE the respective opcode execution every time they are called.
type PermissionPolicy interface {
	// CanCreate checks if the contract creation is allowed.
	CanCreate(signer, caller common.Address) bool
	// CanCall checks if the any type of CALL opcode execution is allowed. This includes
	// contract calls and transfers.
	CanCall(signer, caller, recipient common.Address) bool

	// GetCallHook returns a CallHook that checks if the caller is allowed to perform a call.
	// This is used by the EVM opcode hooks to enforce access control policies.
	GetCallHook(signer common.Address) CallHook
	// GetCreateHook returns a CreateHook that checks if the caller is allowed to deploy contracts.
	// This is used by the EVM opcode hooks to enforce access control policies.
	GetCreateHook(signer common.Address) CreateHook
}

// RestrictedPermissionPolicy is a permission policy that restricts contract creation and calls based on a set of accessControl.
// Note that all the properties are private, this enforces the permissions not to be modified
// anywhere else within the code.
// For users that require a custom permission policy, they can implement the PermissionPolicy interface.
type RestrictedPermissionPolicy struct {
	accessControl *AccessControl
	canCreate     callerFn
	canCall       callerFn
}

func NewRestrictedPermissionPolicy(accessControl *AccessControl, signer common.Address) PermissionPolicy {
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

// GetCallHook returns a CallHook that checks if the caller is allowed to perform a call.
func (p RestrictedPermissionPolicy) GetCallHook(signer common.Address) CallHook {
	return func(_ *vm.EVM, caller, recipient common.Address) error {
		if p.CanCall(signer, caller, recipient) {
			return nil
		}
		return fmt.Errorf("caller address %s does not have permission to perform a call", caller)
	}
}

// GetCreateHook returns a CreateHook that checks if the caller is allowed to deploy contracts.
func (p RestrictedPermissionPolicy) GetCreateHook(signer common.Address) CreateHook {
	return func(_ *vm.EVM, caller common.Address) error {
		if p.CanCreate(signer, caller) {
			return nil
		}
		return fmt.Errorf("caller address %s does not have permission to deploy contracts", caller)
	}
}

// CanCreate implements the PermissionPolicy interface.
// It allows contract creation if access type is set to everybody.
// Otherwise, it checks if:
// - The signer is allowed to do so.
// - If the signer is not allowed, then we check if the caller is allowed to do so.
func (p RestrictedPermissionPolicy) CanCreate(_, caller common.Address) bool {
	return p.canCreate(caller)
}

type callerFn = func(caller common.Address) bool

func getCanCreateFn(accessControl *AccessControl, signer common.Address) callerFn {
	addresses := accessControl.Create.AccessControlList

	switch accessControl.Create.AccessType {
	case AccessTypePermissionless:
		return permissionlessCheckFn(addresses, signer)
	case AccessTypeRestricted:
		return func(_ common.Address) bool { return false }
	case AccessTypePermissioned:
		return permissionedCheckFn(addresses, signer)
	}
	return func(_ common.Address) bool { return false }
}

// CanCreate implements the PermissionPolicy interface.
// It allows calls if access type is set to everybody.
// Otherwise, it checks if:
// - The signer is allowed to do so.
// - If the signer is not allowed, then we check if the caller is allowed to do so.
func (p RestrictedPermissionPolicy) CanCall(_, caller, _ common.Address) bool {
	return p.canCall(caller)
}

func getCanCallFn(accessControl *AccessControl, signer common.Address) callerFn {
	addresses := accessControl.Call.AccessControlList

	switch accessControl.Call.AccessType {
	case AccessTypePermissionless:
		return permissionlessCheckFn(addresses, signer)
	case AccessTypeRestricted:
		return func(_ common.Address) bool { return false }
	case AccessTypePermissioned:
		return permissionedCheckFn(addresses, signer)
	}
	return func(_ common.Address) bool { return false }
}

// permissionlessCheckFn returns a callerFn that returns true unless the signer or the caller is
// within the addresses slice.
func permissionlessCheckFn(addresses []string, signer common.Address) callerFn {
	strSigner := signer.String()
	isSignerBlocked := !slices.Contains(addresses, strSigner)
	return func(caller common.Address) bool {
		strCaller := caller.String()
		return isSignerBlocked && !slices.Contains(addresses, strCaller)
	}
}

// permissionedCheckFn returns a callerFn that returns true if the signer or caller
// is within the addresses slice.
func permissionedCheckFn(addresses []string, signer common.Address) callerFn {
	strSigner := signer.String()
	isSignerAllowed := slices.Contains(addresses, strSigner)
	return func(caller common.Address) bool {
		strCaller := caller.String()
		return isSignerAllowed || slices.Contains(addresses, strCaller)
	}
}
