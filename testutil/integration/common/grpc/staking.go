// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetDelegation returns the delegation for the given delegator and validator addresses.
func (gqh *IntegrationHandler) GetDelegation(delegatorAddress string, validatorAddress string) (*stakingtypes.QueryDelegationResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.Delegation(context.Background(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegatorAddress,
		ValidatorAddr: validatorAddress,
	})
}

// GetValidatorDelegations returns the delegations to a given validator.
func (gqh *IntegrationHandler) GetValidatorDelegations(validatorAddress string) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.ValidatorDelegations(context.Background(), &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validatorAddress,
	})
}

// GetDelegatorDelegations returns the delegations to a given delegator.
func (gqh *IntegrationHandler) GetDelegatorDelegations(delegatorAddress string) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.DelegatorDelegations(context.Background(), &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegatorAddress,
	})
}

// GetRedelegations returns the redelegations to a given delegator and validators.
func (gqh *IntegrationHandler) GetRedelegations(delegatorAddress, srcValidator, dstValidator string) (*stakingtypes.QueryRedelegationsResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.Redelegations(context.Background(), &stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr:    delegatorAddress,
		SrcValidatorAddr: srcValidator,
		DstValidatorAddr: dstValidator,
	})
}

// GetValidatorUnbondingDelegations returns the unbonding delegations to a given validator.
func (gqh *IntegrationHandler) GetValidatorUnbondingDelegations(validatorAddress string) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.ValidatorUnbondingDelegations(context.Background(), &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validatorAddress,
	})
}

// GetDelegatorUnbondingDelegations returns all the unbonding delegations for given delegator.
func (gqh *IntegrationHandler) GetDelegatorUnbondingDelegations(delegatorAddress string) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.DelegatorUnbondingDelegations(context.Background(), &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: delegatorAddress,
	})
}

// GetValidators returns the list of all bonded validators.
func (gqh *IntegrationHandler) GetBondedValidators() (*stakingtypes.QueryValidatorsResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})
}
