// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package grpc

import (
	"fmt"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"golang.org/x/exp/slices"
)

// GetGovParams returns the gov params from the gov module.
func (gqh *IntegrationHandler) GetGovParams(paramsType string) (*govtypes.QueryParamsResponse, error) {
	possibleTypes := []string{"deposit", "tallying", "voting"}
	if !slices.Contains(possibleTypes, paramsType) {
		return nil, fmt.Errorf("invalid params type: %s\npossible types: %s", paramsType, possibleTypes)
	}

	govClient := gqh.network.GetGovClient()
	return govClient.Params(gqh.network.GetContext(), &govtypes.QueryParamsRequest{ParamsType: paramsType})
}

// GetProposal returns the proposal from the gov module.
func (gqh *IntegrationHandler) GetProposal(proposalID uint64) (*govtypes.QueryProposalResponse, error) {
	govClient := gqh.network.GetGovClient()
	return govClient.Proposal(gqh.network.GetContext(), &govtypes.QueryProposalRequest{ProposalId: proposalID})
}
