// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v17/precompiles/common"
)

const (
	// ValidatorDistributionInfoMethod defines the ABI method name for the
	// ValidatorDistributionInfo query.
	ValidatorDistributionInfoMethod = "validatorDistributionInfo"
	// ValidatorOutstandingRewardsMethod defines the ABI method name for the
	// ValidatorOutstandingRewards query.
	ValidatorOutstandingRewardsMethod = "validatorOutstandingRewards"
	// ValidatorCommissionMethod defines the ABI method name for the
	// ValidatorCommission query.
	ValidatorCommissionMethod = "validatorCommission"
	// ValidatorSlashesMethod defines the ABI method name for the
	// ValidatorSlashes query.
	ValidatorSlashesMethod = "validatorSlashes"
	// DelegationRewardsMethod defines the ABI method name for the
	// DelegationRewards query.
	DelegationRewardsMethod = "delegationRewards"
	// DelegationTotalRewardsMethod defines the ABI method name for the
	// DelegationTotalRewards query.
	DelegationTotalRewardsMethod = "delegationTotalRewards"
	// DelegatorValidatorsMethod defines the ABI method name for the
	// DelegatorValidators query.
	DelegatorValidatorsMethod = "delegatorValidators"
	// DelegatorWithdrawAddressMethod defines the ABI method name for the
	// DelegatorWithdrawAddress query.
	DelegatorWithdrawAddressMethod = "delegatorWithdrawAddress"
)

// ValidatorDistributionInfo returns the distribution info for a validator.
func (p Precompile) ValidatorDistributionInfo(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorDistributionInfoRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.ValidatorDistributionInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	out := new(ValidatorDistributionInfoOutput).FromResponse(res)

	return method.Outputs.Pack(out.DistributionInfo)
}

// ValidatorOutstandingRewards returns the outstanding rewards for a validator.
func (p Precompile) ValidatorOutstandingRewards(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorOutstandingRewardsRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.ValidatorOutstandingRewards(ctx, req)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(cmn.NewDecCoinsResponse(res.Rewards.Rewards))
}

// ValidatorCommission returns the commission for a validator.
func (p Precompile) ValidatorCommission(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorCommissionRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.ValidatorCommission(ctx, req)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(cmn.NewDecCoinsResponse(res.Commission.Commission))
}

// ValidatorSlashes returns the slashes for a validator.
func (p Precompile) ValidatorSlashes(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewValidatorSlashesRequest(method, args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.ValidatorSlashes(ctx, req)
	if err != nil {
		return nil, err
	}

	out := new(ValidatorSlashesOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}

// DelegationRewards returns the total rewards accrued by a delegation.
func (p Precompile) DelegationRewards(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDelegationRewardsRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}
	res, err := querier.DelegationRewards(ctx, req)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(cmn.NewDecCoinsResponse(res.Rewards))
}

// DelegationTotalRewards returns the total rewards accrued by a delegation.
func (p Precompile) DelegationTotalRewards(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDelegationTotalRewardsRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.DelegationTotalRewards(ctx, req)
	if err != nil {
		return nil, err
	}

	out := new(DelegationTotalRewardsOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}

// DelegatorValidators returns the validators a delegator is bonded to.
func (p Precompile) DelegatorValidators(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDelegatorValidatorsRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.DelegatorValidators(ctx, req)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Validators)
}

// DelegatorWithdrawAddress returns the withdraw address for a delegator.
func (p Precompile) DelegatorWithdrawAddress(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	req, err := NewDelegatorWithdrawAddressRequest(args)
	if err != nil {
		return nil, err
	}

	querier := distributionkeeper.Querier{Keeper: p.distributionKeeper}

	res, err := querier.DelegatorWithdrawAddress(ctx, req)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.WithdrawAddress)
}
