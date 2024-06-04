// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	"github.com/evmos/evmos/v18/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

type DefaultOpCodesHooks struct {
	accessControl types.PermissionPolicy
	signer        string
}

func NewDefaultOpCodesHooks(accessControl types.PermissionPolicy, signer string) vm.OpCodeHooks {
	return &DefaultOpCodesHooks{
		accessControl: accessControl,
		signer:        signer,
	}
}

func (h *DefaultOpCodesHooks) CreateHook(_ *vm.EVM, caller common.Address) error {
	if h.accessControl.CanCreate(h.signer, caller.String()) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to deploy contracts", h.signer)
}

func (h *DefaultOpCodesHooks) CallHook(_ *vm.EVM, caller common.Address, recipient common.Address) error {
	if h.accessControl.CanCall(h.signer, caller.String(), recipient.String()) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to perform a call", h.signer)
}
