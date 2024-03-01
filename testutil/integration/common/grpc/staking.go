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
