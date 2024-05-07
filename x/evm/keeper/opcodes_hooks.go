package keeper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

type DefaultOpCodesHooks struct {
	permissions PermissionPolicy
	signer      string
}

func NewDefaultOpCodesHooks(permissionsPolicy PermissionPolicy, signer string) vm.OpCodeHooks {
	return &DefaultOpCodesHooks{
		permissions: permissionsPolicy,
		signer:      signer,
	}
}

func (h *DefaultOpCodesHooks) CreateHook(_ *vm.EVM, caller common.Address) error {
	if h.permissions.CanCreate(h.signer, caller.String()) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to deploy contracts", h.signer)
}

func (h *DefaultOpCodesHooks) CallHook(_ *vm.EVM, caller common.Address, recipient common.Address) error {
	if h.permissions.CanCall(h.signer, caller.String(), recipient.String()) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to perform a call", h.signer)
}
