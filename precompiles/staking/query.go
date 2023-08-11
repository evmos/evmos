// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staking

import (
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

const (
	// DelegationMethod defines the ABI method name for the staking Delegation
	// query.
	DelegationMethod = "delegation"
	// UnbondingDelegationMethod defines the ABI method name for the staking
	// UnbondingDelegationMethod query.
	UnbondingDelegationMethod = "unbondingDelegation"
	// ValidatorMethod defines the ABI method name for the staking
	// Validator query.
	ValidatorMethod = "validator"
	// ValidatorsMethod defines the ABI method name for the staking
	// Validators query.
	ValidatorsMethod = "validators"
	// RedelegationMethod defines the ABI method name for the staking
	// Redelegation query.
	RedelegationMethod = "redelegation"
	// RedelegationsMethod defines the ABI method name for the staking
	// Redelegations query.
	RedelegationsMethod = "redelegations"
)

// Delegation returns the delegation that a delegator has with a specific validator.
func (p Precompile) Delegation(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDelegationRequest(args)
	if err != nil {
		return nil, err
	}

	queryServer := stakingkeeper.Querier{Keeper: p.stakingKeeper}

	res, err := queryServer.Delegation(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		// If there is no delegation found, return the response with zero values.
		if strings.Contains(err.Error(), fmt.Sprintf(ErrNoDelegationFound, req.DelegatorAddr, req.ValidatorAddr)) {
			return method.Outputs.Pack(big.NewInt(0), cmn.Coin{Denom: p.stakingKeeper.BondDenom(ctx), Amount: big.NewInt(0)})
		}

		return nil, err
	}

	out := new(DelegationOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}

// UnbondingDelegation returns the delegation currently being unbonded for a delegator from
// a specific validator.
func (p Precompile) UnbondingDelegation(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewUnbondingDelegationRequest(args)
	if err != nil {
		return nil, err
	}

	queryServer := stakingkeeper.Querier{Keeper: p.stakingKeeper}

	res, err := queryServer.UnbondingDelegation(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		// return empty unbonding delegation output if the unbonding delegation is not found
		expError := fmt.Sprintf("unbonding delegation with delegator %s not found for validator %s", req.DelegatorAddr, req.ValidatorAddr)
		if strings.Contains(err.Error(), expError) {
			return method.Outputs.Pack([]UnbondingDelegationEntry{})
		}
		return nil, err
	}

	out := new(UnbondingDelegationOutput).FromResponse(res)

	return method.Outputs.Pack(out.Entries)
}

// Validator returns the validator information for a given validator address.
func (p Precompile) Validator(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorRequest(args)
	if err != nil {
		return nil, err
	}

	queryServer := stakingkeeper.Querier{Keeper: p.stakingKeeper}

	res, err := queryServer.Validator(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		// return empty validator info if the validator is not found
		expError := fmt.Sprintf("validator %s not found", req.ValidatorAddr)
		if strings.Contains(err.Error(), expError) {
			return method.Outputs.Pack(DefaultValidatorOutput().Validator)
		}
		return nil, err
	}

	out := new(ValidatorOutput).FromResponse(res)

	return method.Outputs.Pack(out.Validator)
}

// Validators returns the validators information with a provided status & pagination (optional).
func (p Precompile) Validators(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorsRequest(method, args)
	if err != nil {
		return nil, err
	}

	queryServer := stakingkeeper.Querier{Keeper: p.stakingKeeper}

	res, err := queryServer.Validators(sdk.WrapSDKContext(ctx), req)
	if err != nil {
		return nil, err
	}

	out := new(ValidatorsOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}

// Redelegation returns the redelegation between two validators for a delegator.
func (p Precompile) Redelegation(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	req, err := NewRedelegationRequest(args)
	if err != nil {
		return nil, err
	}

	res, _ := p.stakingKeeper.GetRedelegation(ctx, req.DelegatorAddress, req.ValidatorSrcAddress, req.ValidatorDstAddress)
	out := new(RedelegationOutput).FromResponse(res)

	return method.Outputs.Pack(out.Entries)
}

// Redelegations returns the redelegations according to
// the specified criteria (delegator address and/or validator source address
// and/or validator destination address or all existing redelegations) with pagination.
// Pagination is only supported for querying redelegations from a source validator or to query all redelegations.
func (p Precompile) Redelegations(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	req, err := NewRedelegationsRequest(method, args)
	if err != nil {
		return nil, err
	}

	queryServer := stakingkeeper.Querier{Keeper: p.stakingKeeper}

	res, err := queryServer.Redelegations(ctx, req)
	if err != nil {
		return nil, err
	}

	out := new(RedelegationsOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}

// Allowance returns the remaining allowance of a spender to the contract.
func (p Precompile) Allowance(
	ctx sdk.Context,
	method *abi.Method,
	_ *vm.Contract,
	args []interface{},
) ([]byte, error) {
	granter, grantee, msg, err := authorization.CheckAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	msgAuthz, _ := p.AuthzKeeper.GetAuthorization(ctx, grantee.Bytes(), granter.Bytes(), msg)

	if msgAuthz == nil {
		return method.Outputs.Pack(big.NewInt(0))
	}

	stakeAuthz, ok := msgAuthz.(*stakingtypes.StakeAuthorization)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "staking authorization", &stakingtypes.StakeAuthorization{}, stakeAuthz)
	}

	if stakeAuthz.MaxTokens == nil {
		return method.Outputs.Pack(abi.MaxUint256)
	}

	return method.Outputs.Pack(stakeAuthz.MaxTokens.Amount.BigInt())
}
