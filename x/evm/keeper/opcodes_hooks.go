// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/evmos/evmos/v18/x/evm/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// DefaultOpCodesHooks is the default implementation of OpCodeHooks for EVMOS chain
// The hooks are used to enforce access control policies on EVM operations.
// They are ran BEFORE the respective opcode execution every time they are called.
type DefaultOpCodesHooks struct {
	accessControl types.PermissionPolicy
	signer        common.Address
	keeper        Keeper
	ctx           sdktypes.Context
	callHooks     []callHook
	createHooks   []createHook
}

type createHook func(ev *vm.EVM, caller common.Address) error
type callHook func(ev *vm.EVM, caller common.Address, recipient common.Address) error

// NewDefaultOpCodesHooks creates a new DefaultOpCodesHooks instance
func NewDefaultOpCodesHooks() *DefaultOpCodesHooks {
	return &DefaultOpCodesHooks{}
}

func (h *DefaultOpCodesHooks) AddCallHook(hook callHook) {
	h.callHooks = append(h.callHooks, hook)
}

func (h *DefaultOpCodesHooks) AddCreateHook(hook createHook) {
	h.createHooks = append(h.createHooks, hook)
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
