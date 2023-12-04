// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	"github.com/evmos/evmos/v16/x/evm/statedb"
)

const (
	// CreateValidatorMethod defines the ABI method name for the staking create validator transaction
	CreateValidatorMethod = "createValidator"
	// DelegateMethod defines the ABI method name for the staking Delegate
	// transaction.
	DelegateMethod = "delegate"
	// UndelegateMethod defines the ABI method name for the staking Undelegate
	// transaction.
	UndelegateMethod = "undelegate"
	// RedelegateMethod defines the ABI method name for the staking Redelegate
	// transaction.
	RedelegateMethod = "redelegate"
	// CancelUnbondingDelegationMethod defines the ABI method name for the staking
	// CancelUnbondingDelegation transaction.
	CancelUnbondingDelegationMethod = "cancelUnbondingDelegation"
)

const (
	// DelegateAuthz defines the authorization type for the staking Delegate
	DelegateAuthz = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE
	// UndelegateAuthz defines the authorization type for the staking Undelegate
	UndelegateAuthz = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE
	// RedelegateAuthz defines the authorization type for the staking Redelegate
	RedelegateAuthz = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE
	// CancelUnbondingDelegationAuthz defines the authorization type for the staking
	CancelUnbondingDelegationAuthz = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_CANCEL_UNBONDING_DELEGATION
)

// CreateValidator performs create validator.
func (p Precompile) CreateValidator(
	ctx sdk.Context,
	origin common.Address,
	_ *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgCreateValidator(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"commission", msg.Commission.String(),
		"min_self_delegation", msg.MinSelfDelegation.String(),
		"delegator_address", delegatorHexAddr.String(),
		"validator_address", msg.ValidatorAddress,
		"pubkey", msg.Pubkey.String(),
		"value", msg.Value.Amount.String(),
	)

	// we only allow the tx signer "origin" to create their own validator.
	if origin != delegatorHexAddr {
		return nil, fmt.Errorf(ErrDifferentOriginFromDelegator, origin.String(), delegatorHexAddr.String())
	}

	// Execute the transaction using the message server
	msgSrv := stakingkeeper.NewMsgServerImpl(&p.stakingKeeper)
	if _, err = msgSrv.CreateValidator(sdk.WrapSDKContext(ctx), msg); err != nil {
		return nil, err
	}

	// Emit the event for the delegate transaction
	if err = p.EmitCreateValidatorEvent(ctx, stateDB, msg, delegatorHexAddr); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// Delegate performs a delegation of coins from a delegator to a validator.
func (p Precompile) Delegate(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgDelegate(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ delegator_address: %s, validator_address: %s, amount: %s }",
			delegatorHexAddr,
			msg.ValidatorAddress,
			msg.Amount.Amount,
		),
	)

	var (
		// stakeAuthz is the authorization grant for the caller and the delegator address
		stakeAuthz *stakingtypes.StakeAuthorization
		// expiration is the expiration time of the authorization grant
		expiration *time.Time

		// isCallerOrigin is true when the contract caller is the same as the origin
		isCallerOrigin = contract.CallerAddress == origin
		// isCallerDelegator is true when the contract caller is the same as the delegator
		isCallerDelegator = contract.CallerAddress == delegatorHexAddr
	)

	// The provided delegator address should always be equal to the origin address.
	// In case the contract caller address is the same as the delegator address provided,
	// update the delegator address to be equal to the origin address.
	// Otherwise, if the provided delegator address is different from the origin address,
	// return an error because is a forbidden operation
	if isCallerDelegator {
		delegatorHexAddr = origin
	} else if origin != delegatorHexAddr {
		return nil, fmt.Errorf(ErrDifferentOriginFromDelegator, origin.String(), delegatorHexAddr.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	if !isCallerOrigin {
		// Check if the authorization grant exists for the caller and the origin
		stakeAuthz, expiration, err = authorization.CheckAuthzAndAllowanceForGranter(ctx, p.AuthzKeeper, contract.CallerAddress, delegatorHexAddr, &msg.Amount, DelegateMsg)
		if err != nil {
			return nil, err
		}
	}

	// Execute the transaction using the message server
	msgSrv := stakingkeeper.NewMsgServerImpl(&p.stakingKeeper)
	if _, err = msgSrv.Delegate(sdk.WrapSDKContext(ctx), msg); err != nil {
		return nil, err
	}

	// Only update the authorization if the contract caller is different from the origin
	if !isCallerOrigin {
		if err := p.UpdateStakingAuthorization(ctx, contract.CallerAddress, delegatorHexAddr, stakeAuthz, expiration, DelegateMsg, msg); err != nil {
			return nil, err
		}
	}

	// Emit the event for the delegate transaction
	if err = p.EmitDelegateEvent(ctx, stateDB, msg, delegatorHexAddr); err != nil {
		return nil, err
	}

	// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB.
	// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
	if isCallerDelegator {
		stateDB.(*statedb.StateDB).SubBalance(contract.CallerAddress, msg.Amount.Amount.BigInt())
	}

	return method.Outputs.Pack(true)
}

// Undelegate performs the undelegation of coins from a validator for a delegate.
// The provided amount cannot be negative. This is validated in the msg.ValidateBasic() function.
func (p Precompile) Undelegate(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgUndelegate(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ delegator_address: %s, validator_address: %s, amount: %s }",
			delegatorHexAddr,
			msg.ValidatorAddress,
			msg.Amount.Amount,
		),
	)

	var (
		// stakeAuthz is the authorization grant for the caller and the delegator address
		stakeAuthz *stakingtypes.StakeAuthorization
		// expiration is the expiration time of the authorization grant
		expiration *time.Time

		// isCallerOrigin is true when the contract caller is the same as the origin
		isCallerOrigin = contract.CallerAddress == origin
		// isCallerDelegator is true when the contract caller is the same as the delegator
		isCallerDelegator = contract.CallerAddress == delegatorHexAddr
	)

	// The provided delegator address should always be equal to the origin address.
	// In case the contract caller address is the same as the delegator address provided,
	// update the delegator address to be equal to the origin address.
	// Otherwise, if the provided delegator address is different from the origin address,
	// return an error because is a forbidden operation
	if isCallerDelegator {
		delegatorHexAddr = origin
	} else if origin != delegatorHexAddr {
		return nil, fmt.Errorf(ErrDifferentOriginFromDelegator, origin.String(), delegatorHexAddr.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	if !isCallerOrigin {
		// Check if the authorization grant exists for the caller and the origin
		stakeAuthz, expiration, err = authorization.CheckAuthzAndAllowanceForGranter(ctx, p.AuthzKeeper, contract.CallerAddress, delegatorHexAddr, &msg.Amount, UndelegateMsg)
		if err != nil {
			return nil, err
		}
	}

	// Execute the transaction using the message server
	msgSrv := stakingkeeper.NewMsgServerImpl(&p.stakingKeeper)
	res, err := msgSrv.Undelegate(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Only update the authorization if the contract caller is different from the origin
	if !isCallerOrigin {
		if err := p.UpdateStakingAuthorization(ctx, contract.CallerAddress, delegatorHexAddr, stakeAuthz, expiration, UndelegateMsg, msg); err != nil {
			return nil, err
		}
	}

	// Emit the event for the undelegate transaction
	if err = p.EmitUnbondEvent(ctx, stateDB, msg, delegatorHexAddr, res.CompletionTime.UTC().Unix()); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.CompletionTime.UTC().Unix())
}

// Redelegate performs a redelegation of coins for a delegate from a source validator
// to a destination validator.
// The provided amount cannot be negative. This is validated in the msg.ValidateBasic() function.
func (p Precompile) Redelegate(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgRedelegate(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ delegator_address: %s, validator_src_address: %s, validator_dst_address: %s, amount: %s }",
			delegatorHexAddr,
			msg.ValidatorSrcAddress,
			msg.ValidatorDstAddress,
			msg.Amount.Amount,
		),
	)

	var (
		// stakeAuthz is the authorization grant for the caller and the delegator address
		stakeAuthz *stakingtypes.StakeAuthorization
		// expiration is the expiration time of the authorization grant
		expiration *time.Time

		// isCallerOrigin is true when the contract caller is the same as the origin
		isCallerOrigin = contract.CallerAddress == origin
		// isCallerDelegator is true when the contract caller is the same as the delegator
		isCallerDelegator = contract.CallerAddress == delegatorHexAddr
	)

	// The provided delegator address should always be equal to the origin address.
	// In case the contract caller address is the same as the delegator address provided,
	// update the delegator address to be equal to the origin address.
	// Otherwise, if the provided delegator address is different from the origin address,
	// return an error because is a forbidden operation
	if isCallerDelegator {
		delegatorHexAddr = origin
	} else if origin != delegatorHexAddr {
		return nil, fmt.Errorf(ErrDifferentOriginFromDelegator, origin.String(), delegatorHexAddr.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	if !isCallerOrigin {
		// Check if the authorization grant exists for the caller and the origin
		stakeAuthz, expiration, err = authorization.CheckAuthzAndAllowanceForGranter(ctx, p.AuthzKeeper, contract.CallerAddress, delegatorHexAddr, &msg.Amount, RedelegateMsg)
		if err != nil {
			return nil, err
		}
	}

	msgSrv := stakingkeeper.NewMsgServerImpl(&p.stakingKeeper)
	res, err := msgSrv.BeginRedelegate(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Only update the authorization if the contract caller is different from the origin
	if !isCallerOrigin {
		if err := p.UpdateStakingAuthorization(ctx, contract.CallerAddress, delegatorHexAddr, stakeAuthz, expiration, RedelegateMsg, msg); err != nil {
			return nil, err
		}
	}

	if err = p.EmitRedelegateEvent(ctx, stateDB, msg, delegatorHexAddr, res.CompletionTime.UTC().Unix()); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.CompletionTime.UTC().Unix())
}

// CancelUnbondingDelegation will cancel the unbonding of a delegation and delegate
// back to the validator being unbonded from.
// The provided amount cannot be negative. This is validated in the msg.ValidateBasic() function.
func (p Precompile) CancelUnbondingDelegation(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, delegatorHexAddr, err := NewMsgCancelUnbondingDelegation(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ delegator_address: %s, validator_address: %s, amount: %s, creation_height: %d }",
			delegatorHexAddr,
			msg.ValidatorAddress,
			msg.Amount.Amount,
			msg.CreationHeight,
		),
	)

	var (
		// stakeAuthz is the authorization grant for the caller and the delegator address
		stakeAuthz *stakingtypes.StakeAuthorization
		// expiration is the expiration time of the authorization grant
		expiration *time.Time

		// isCallerOrigin is true when the contract caller is the same as the origin
		isCallerOrigin = contract.CallerAddress == origin
		// isCallerDelegator is true when the contract caller is the same as the delegator
		isCallerDelegator = contract.CallerAddress == delegatorHexAddr
	)

	// The provided delegator address should always be equal to the origin address.
	// In case the contract caller address is the same as the delegator address provided,
	// update the delegator address to be equal to the origin address.
	// Otherwise, if the provided delegator address is different from the origin address,
	// return an error because is a forbidden operation
	if isCallerDelegator {
		delegatorHexAddr = origin
	} else if origin != delegatorHexAddr {
		return nil, fmt.Errorf(ErrDifferentOriginFromDelegator, origin.String(), delegatorHexAddr.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	if !isCallerOrigin {
		// Check if the authorization grant exists for the caller and the origin
		stakeAuthz, expiration, err = authorization.CheckAuthzAndAllowanceForGranter(ctx, p.AuthzKeeper, contract.CallerAddress, delegatorHexAddr, &msg.Amount, CancelUnbondingDelegationMsg)
		if err != nil {
			return nil, err
		}
	}

	msgSrv := stakingkeeper.NewMsgServerImpl(&p.stakingKeeper)
	if _, err = msgSrv.CancelUnbondingDelegation(sdk.WrapSDKContext(ctx), msg); err != nil {
		return nil, err
	}

	// Only update the authorization if the contract caller is different from the origin
	if !isCallerOrigin {
		if err := p.UpdateStakingAuthorization(ctx, contract.CallerAddress, delegatorHexAddr, stakeAuthz, expiration, CancelUnbondingDelegationMsg, msg); err != nil {
			return nil, err
		}
	}

	if err = p.EmitCancelUnbondingDelegationEvent(ctx, stateDB, msg, delegatorHexAddr); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
