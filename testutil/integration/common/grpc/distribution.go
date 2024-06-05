// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"context"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// GetDelegationTotalRewards returns the total delegation rewards for the given delegator.
func (gqh *IntegrationHandler) GetDelegationTotalRewards(delegatorAddress string) (*distrtypes.QueryDelegationTotalRewardsResponse, error) {
	distrClient := gqh.network.GetDistrClient()
	return distrClient.DelegationTotalRewards(context.Background(), &distrtypes.QueryDelegationTotalRewardsRequest{
		DelegatorAddress: delegatorAddress,
	})
}

// GetDelegationRewards returns the  delegation rewards for the given delegator and validator.
func (gqh *IntegrationHandler) GetDelegationRewards(delegatorAddress string, validatorAddress string) (*distrtypes.QueryDelegationRewardsResponse, error) {
	distrClient := gqh.network.GetDistrClient()
	return distrClient.DelegationRewards(context.Background(), &distrtypes.QueryDelegationRewardsRequest{
		DelegatorAddress: delegatorAddress,
		ValidatorAddress: validatorAddress,
	})
}

// GetDelegatorWithdrawAddr returns the withdraw address the given delegator.
func (gqh *IntegrationHandler) GetDelegatorWithdrawAddr(delegatorAddress string) (*distrtypes.QueryDelegatorWithdrawAddressResponse, error) {
	distrClient := gqh.network.GetDistrClient()
	return distrClient.DelegatorWithdrawAddress(context.Background(), &distrtypes.QueryDelegatorWithdrawAddressRequest{
		DelegatorAddress: delegatorAddress,
	})
}

// GetValidatorCommission returns the commission for the given validator.
func (gqh *IntegrationHandler) GetValidatorCommission(validatorAddress string) (*distrtypes.QueryValidatorCommissionResponse, error) {
	distrClient := gqh.network.GetDistrClient()
	return distrClient.ValidatorCommission(context.Background(), &distrtypes.QueryValidatorCommissionRequest{
		ValidatorAddress: validatorAddress,
	})
}

// GetValidatorOutstandingRewards returns the  delegation rewards for the given delegator and validator.
func (gqh *IntegrationHandler) GetValidatorOutstandingRewards(validatorAddress string) (*distrtypes.QueryValidatorOutstandingRewardsResponse, error) {
	distrClient := gqh.network.GetDistrClient()
	return distrClient.ValidatorOutstandingRewards(context.Background(), &distrtypes.QueryValidatorOutstandingRewardsRequest{
		ValidatorAddress: validatorAddress,
	})
}

// GetCommunityPool queries the community pool coins.
func (gqh *IntegrationHandler) GetCommunityPool() (*distrtypes.QueryCommunityPoolResponse, error) {
	distrClient := gqh.network.GetDistrClient()
	return distrClient.CommunityPool(context.Background(), &distrtypes.QueryCommunityPoolRequest{})
}
