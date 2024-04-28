// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package grpc

import (
	"context"

	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
)

// GetTokenPair returns the ERC-20 token pair corresponding to the given token.
func (gqh *IntegrationHandler) GetTokenPair(token string) (*erc20types.QueryTokenPairResponse, error) {
	erc20client := gqh.network.GetERC20Client()
	return erc20client.TokenPair(context.Background(), &erc20types.QueryTokenPairRequest{Token: token})
}

// GetTokenPairs returns the ERC-20 token pairs.
func (gqh *IntegrationHandler) GetTokenPairs() (*erc20types.QueryTokenPairsResponse, error) {
	erc20Client := gqh.network.GetERC20Client()
	return erc20Client.TokenPairs(context.Background(), &erc20types.QueryTokenPairsRequest{})
}
