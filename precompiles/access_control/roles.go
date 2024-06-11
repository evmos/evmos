// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE

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

// HasRole checks if an account has a role.
func (p Precompile) HasRole(
	ctx sdk.Context,
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

// GetRoleAdmin returns the admin role of a role.
func (p Precompile) GetRoleAdmin(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	roleArray, ok := args[0].([32]uint8)
	if !ok {
		return nil, fmt.Errorf(ErrInvalidRoleArgument)
	}

	var role common.Hash
	copy(role[:], roleArray[:])

	roleAdmin := p.AccessControlKeeper.GetRoleAdmin(ctx, p.Address(), role)

	fmt.Println(roleAdmin)

	return method.Outputs.Pack(roleAdmin)
}

// GrantRole grants a role to an account.
func (p Precompile) GrantRole(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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

	// If the user already has the role, return
	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)
	if hasRole {
		return nil, nil
	}

	p.AccessControlKeeper.SetRole(ctx, p.Address(), role, account)

	if err := p.EmitEventRoleGranted(ctx, stateDB, role, account, contract.CallerAddress); err != nil {
		return nil, err
	}

	return nil, nil
}

// RevokeRole revokes a role from an account.
func (p Precompile) RevokeRole(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
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

	// If the user does not have the role, return
	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)
	if !hasRole {
		return nil, nil
	}

	p.AccessControlKeeper.DeleteRole(ctx, p.Address(), role, account)

	if err := p.EmitEventRoleRevoked(ctx, stateDB, role, account, contract.CallerAddress); err != nil {
		return nil, err
	}

	return nil, nil
}

// RenounceRole renounces a role from an account.
func (p Precompile) RenounceRole(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	args []interface{},
) ([]byte, error) {
	role, account, err := ParseRoleArgs(args)
	if err != nil {
		return nil, err
	}

	if account != contract.CallerAddress {
		return nil, fmt.Errorf(ErrRenounceRoleDifferentThanCaller)
	}

	// If the user does not have the role, return
	hasRole := p.AccessControlKeeper.HasRole(ctx, p.Address(), role, account)
	if !hasRole {
		return nil, nil
	}

	p.AccessControlKeeper.DeleteRole(ctx, p.Address(), role, account)

	if err := p.EmitEventRoleRevoked(ctx, stateDB, role, account, contract.CallerAddress); err != nil {
		return nil, err
	}

	return nil, nil
}

// onlyRole checks if the sender has the role.
func (p Precompile) onlyRole(ctx sdk.Context, role common.Hash, sender common.Address) error {
	if !p.AccessControlKeeper.HasRole(ctx, p.Address(), role, sender) {
		return fmt.Errorf(ErrSenderNoRole)
	}
	return nil
}
