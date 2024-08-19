// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

// OpCodeHooks extends the geth OpCodeHooks interface to add custom hooks for EVM operations.
// The hooks run before the respective opcode execution every time they are called.
type OpCodeHooks interface {
	vm.OpCodeHooks
	AddCallHooks(hooks ...CallHook)
	AddCreateHooks(hooks ...CreateHook)
}

// DefaultOpCodesHooks is the default implementation of OpCodeHooks for EVMOS chain
// The hooks are used to enforce access control policies on EVM operations.
// They are ran BEFORE the respective opcode execution every time they are called.
type DefaultOpCodesHooks struct {
	callHooks   []CallHook
	createHooks []CreateHook
}

// Make sure we comply with geth's OpCodeHooks interface
var _ OpCodeHooks = (*DefaultOpCodesHooks)(nil)

type (
	CreateHook func(ev *vm.EVM, caller common.Address) error
	CallHook   func(ev *vm.EVM, caller common.Address, recipient common.Address) error
)

// NewDefaultOpCodesHooks creates a new DefaultOpCodesHooks instance
func NewDefaultOpCodesHooks() OpCodeHooks {
	return &DefaultOpCodesHooks{}
}

// AddCallHooks adds one or more hooks to the queue to be executed before the CALL opcode.
// Hooks will be executed in the order they are added.
func (h *DefaultOpCodesHooks) AddCallHooks(hooks ...CallHook) {
	h.callHooks = append(h.callHooks, hooks...)
}

// AddCreateHooks adds one or more hooks to the queue to be executed before the CREATE opcode.
// Hooks will be executed in the order they are added.
func (h *DefaultOpCodesHooks) AddCreateHooks(hooks ...CreateHook) {
	h.createHooks = append(h.createHooks, hooks...)
}

// CreateHook checks if the caller has permission to deploy contracts
func (h *DefaultOpCodesHooks) CreateHook(evm *vm.EVM, caller common.Address) error {
	for _, hook := range h.createHooks {
		if err := hook(evm, caller); err != nil {
			return err
		}
	}
	return nil
}

// CallHook checks if the caller has permission to perform a call
func (h *DefaultOpCodesHooks) CallHook(evm *vm.EVM, caller common.Address, recipient common.Address) error {
	for _, hook := range h.callHooks {
		if err := hook(evm, caller, recipient); err != nil {
			return err
		}
	}
	return nil
}
