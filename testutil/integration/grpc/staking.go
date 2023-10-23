// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetBalance returns the balance for the given address.
func (gqh *IntegrationHandler) GetDelegation(delegatorAddress string, validatorAddress string) (*stakingtypes.QueryDelegationResponse, error) {
	stakingClient := gqh.network.GetStakingClient()
	return stakingClient.Delegation(context.Background(), &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegatorAddress,
		ValidatorAddr: validatorAddress,
	})
}
