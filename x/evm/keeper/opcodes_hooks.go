package keeper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
)

type DefaultOpCodesHooks struct {
	isTransferCall bool
	accessControl  PermissionPolicy
	signer         string
}

func NewDefaultOpCodesHooks(msg core.Message, accessControl PermissionPolicy, signer string) vm.OpCodeHooks {
	isTransferCall := IsTransferCall(msg)
	return &DefaultOpCodesHooks{
		isTransferCall: isTransferCall,
		accessControl:  accessControl,
		signer:         signer,
	}
}

func (h *DefaultOpCodesHooks) CreateHook(_ *vm.EVM, caller common.Address) error {
	if h.accessControl.CanCreate(h.signer, caller.String()) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to deploy contracts", h.signer)
}

func (h *DefaultOpCodesHooks) CallHook(_ *vm.EVM, caller common.Address, recipient common.Address) error {
	if h.isTransferCall || h.accessControl.CanCall(h.signer, caller.String(), recipient.String()) {
		return nil
	}
	return fmt.Errorf("caller address %s does not have permission to perform a call", h.signer)
}
