// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v17/precompiles/authorization"
)

const (
	// CreateClawbackVestingAccountMethod defines the ABI method name for the vesting CreateClawbackVestingAccount
	// transaction.
	CreateClawbackVestingAccountMethod = "createClawbackVestingAccount"
	// FundVestingAccountMethod defines the ABI method name for the vesting FundVestingAccount transaction.
	FundVestingAccountMethod = "fundVestingAccount"
	// ClawbackMethod defines the ABI method name for the vesting  Clawback transaction.
	ClawbackMethod = "clawback"
	// UpdateVestingFunderMethod defines the ABI method name for the vesting UpdateVestingFunder transaction.
	UpdateVestingFunderMethod = "updateVestingFunder"
	// ConvertVestingAccountMethod defines the ABI method name for the vesting ConvertVestingAccount transaction.
	ConvertVestingAccountMethod = "convertVestingAccount"
)

// CreateClawbackVestingAccount creates a new clawback vesting account
func (p Precompile) CreateClawbackVestingAccount(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, funderAddr, vestingAddr, err := NewMsgCreateClawbackVestingAccount(args)
	if err != nil {
		return nil, err
	}

	// Check if the origin matches the vesting address
	if origin != vestingAddr {
		return nil, fmt.Errorf(ErrDifferentFromOrigin, origin, vestingAddr)
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf("{ from_address: %s, to_address: %s }", msg.FunderAddress, msg.VestingAddress),
	)

	_, err = p.vestingKeeper.CreateClawbackVestingAccount(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitCreateClawbackVestingAccountEvent(ctx, stateDB, funderAddr, vestingAddr); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// FundVestingAccount funds a vesting account by creating vesting schedules
func (p Precompile) FundVestingAccount(
	ctx sdk.Context,
	contract *vm.Contract,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, funderAddr, vestingAddr, lockupPeriods, vestingPeriods, err := NewMsgFundVestingAccount(args, method)
	if err != nil {
		return nil, err
	}

	// if caller address is origin, the funder MUST match the origin
	if contract.CallerAddress == origin && origin != funderAddr {
		return nil, fmt.Errorf(ErrDifferentFromOrigin, origin, funderAddr)
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ from_address: %s, to_address: %s, start_time: %s, lockup_periods: %s, vesting_periods: %s }",
			msg.FunderAddress, msg.VestingAddress, msg.StartTime, msg.LockupPeriods, msg.VestingPeriods,
		),
	)

	if contract.CallerAddress != origin {
		// check if authorization exists
		_, _, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, contract.CallerAddress, origin, FundVestingAccountMsgURL)
		if err != nil {
			return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
		}
	}

	_, err = p.vestingKeeper.FundVestingAccount(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitFundVestingAccountEvent(ctx, stateDB, msg, funderAddr, vestingAddr, lockupPeriods, vestingPeriods); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// Clawback clawbacks tokens from a clawback vesting account
func (p Precompile) Clawback(
	ctx sdk.Context,
	contract *vm.Contract,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, funderAddr, accountAddr, destAddr, err := NewMsgClawback(args)
	if err != nil {
		return nil, err
	}

	// if caller address is origin, the funder MUST match the origin
	if contract.CallerAddress == origin && origin != funderAddr {
		return nil, fmt.Errorf(ErrDifferentFunderOrigin, origin, funderAddr)
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ funder_address: %s, account_address: %s, dest_address: %s }",
			msg.FunderAddress, msg.AccountAddress, msg.DestAddress,
		),
	)

	if contract.CallerAddress != origin {
		// check if authorization exists
		_, _, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, contract.CallerAddress, origin, ClawbackMsgURL)
		if err != nil {
			return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
		}
	}

	response, err := p.vestingKeeper.Clawback(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitClawbackEvent(ctx, stateDB, funderAddr, accountAddr, destAddr); err != nil {
		return nil, err
	}

	out := new(ClawbackOutput).FromResponse(response)

	return method.Outputs.Pack(out.Coins)
}

// UpdateVestingFunder updates the vesting funder of a clawback vesting account
func (p Precompile) UpdateVestingFunder(
	ctx sdk.Context,
	contract *vm.Contract,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, funderAddr, newFunderAddr, vestingAddr, err := NewMsgUpdateVestingFunder(args)
	if err != nil {
		return nil, err
	}

	// if caller address is origin, the funder MUST match the origin
	if contract.CallerAddress == origin && origin != funderAddr {
		return nil, fmt.Errorf(ErrDifferentFunderOrigin, origin, funderAddr)
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ funder_address: %s, new_funder_address: %s, vesting_address: %s }",
			msg.FunderAddress, msg.NewFunderAddress, msg.VestingAddress,
		),
	)

	if contract.CallerAddress != origin {
		// check if authorization exists
		_, _, err := authorization.CheckAuthzExists(ctx, p.AuthzKeeper, contract.CallerAddress, origin, UpdateVestingFunderMsgURL)
		if err != nil {
			return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
		}
	}

	_, err = p.vestingKeeper.UpdateVestingFunder(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitUpdateVestingFunderEvent(ctx, stateDB, funderAddr, newFunderAddr, vestingAddr); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// ConvertVestingAccount converts a clawback vesting account to a base account once the vesting period is over.
func (p Precompile) ConvertVestingAccount(
	ctx sdk.Context,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, vestingAddr, err := NewMsgConvertVestingAccount(args)
	if err != nil {
		return nil, err
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf("{ vestingAddress: %s }", msg.VestingAddress),
	)

	_, err = p.vestingKeeper.ConvertVestingAccount(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitConvertVestingAccountEvent(ctx, stateDB, vestingAddr); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
