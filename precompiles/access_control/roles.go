// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package accesscontrol

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v18/x/access_control/types"
)

const (
	EventRoleAdminChanged = "RoleAdminChanged"
	EventRoleGranted      = "RoleGranted"
	EventRoleRevoked      = "RoleRevoked"

	MethodHasRole      = "hasRole"
	MethodGetRoleAdmin = "getRoleAdmin"
	MethodGrantRole    = "grantRole"
	MethodRevokeRole   = "revokeRole"
	MethodRenounceRole = "renounceRole"
)

var (
	RoleDefaultAdmin = types.RoleDefaultAdmin
	RoleMinter       = crypto.Keccak256Hash([]byte("MINTER_ROLE"))
	RoleBurner       = crypto.Keccak256Hash([]byte("BURNER_ROLE"))
)

func (p Precompile) HasRole(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	role, account, err := ParseRoleArgs(args)
	if err != nil {
		return nil, err
	}

	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)

	return method.Outputs.Pack(hasRole)
}

func (p Precompile) GetRoleAdmin(
	ctx sdk.Context,
	_ *vm.Contract,
	_ vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	role, ok := args[0].(common.Hash)
	if !ok {
		return nil, fmt.Errorf("invalid role argument")
	}

	roleAdmin := p.AccessControlKeeper.GetRoleAdmin(ctx, p.Address(), role)

	return method.Outputs.Pack(roleAdmin)
}

func (p Precompile) GrantRole(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	role, account, err := ParseRoleArgs(args)
	if err != nil {
		return nil, err
	}

	roleAdmin := p.AccessControlKeeper.GetRoleAdmin(ctx, p.Address(), role)

	if err := p.onlyRole(ctx, roleAdmin, contract.CallerAddress); err != nil {
		return nil, err
	}

	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)
	if hasRole {
		return nil, nil
	}

	p.AccessControlKeeper.SetRole(ctx, p.Address(), role, account)

	// TODO: emit event

	return nil, nil
}

func (p Precompile) RevokeRole(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	role, account, err := ParseRoleArgs(args)
	if err != nil {
		return nil, err
	}

	roleAdmin := p.AccessControlKeeper.GetRoleAdmin(ctx, p.Address(), role)

	if err := p.onlyRole(ctx, roleAdmin, contract.CallerAddress); err != nil {
		return nil, err
	}

	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)
	if hasRole {
		return nil, nil
	}

	p.AccessControlKeeper.SetRole(ctx, p.Address(), role, account)

	// TODO: emit event

	return nil, nil
}

func (p Precompile) RenounceRole(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	role, account, err := ParseRoleArgs(args)
	if err != nil {
		return nil, err
	}

	if account != contract.CallerAddress {
		return nil, fmt.Errorf("access_control: can only renounce roles for self")
	}

	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)
	if hasRole {
		return nil, nil
	}

	p.AccessControlKeeper.SetRole(ctx, p.Address(), role, account)

	// TODO: emit event

	return nil, nil
}

func (p Precompile) onlyRole(ctx sdk.Context, role common.Hash, sender common.Address) error {
	if !p.AccessControlKeeper.HasRole(ctx, p.Address(), role, sender) {
		return fmt.Errorf("access_control: sender does not have the role")
	}
	return nil
}
