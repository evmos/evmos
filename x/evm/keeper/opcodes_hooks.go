// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/evmos/evmos/v18/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// DefaultOpCodesHooks is the default implementation of OpCodeHooks for EVMOS chain
// The hooks are used to enforce access control policies on EVM operations.
// They are ran BEFORE the respective opcode execution every time they are called.
type DefaultOpCodesHooks struct {
	accessControl types.PermissionPolicy
	signer        common.Address
}

// NewDefaultOpCodesHooks creates a new DefaultOpCodesHooks instance
func NewDefaultOpCodesHooks(accessControl types.PermissionPolicy, signer common.Address) vm.OpCodeHooks {
	return &DefaultOpCodesHooks{
		accessControl: accessControl,
		signer:        signer,
	}
}

// CreateHook checks if the caller has permission to deploy contracts
func (h *DefaultOpCodesHooks) CreateHook(_ *vm.EVM, caller common.Address) error {
	if h.accessControl.CanCreate(h.signer, caller) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to deploy contracts", h.signer)
}

// CallHook checks if the caller has permission to perform a call
func (h *DefaultOpCodesHooks) CallHook(_ *vm.EVM, caller common.Address, recipient common.Address) error {
	if h.accessControl.CanCall(h.signer, caller, recipient) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to perform a call", h.signer)
}
