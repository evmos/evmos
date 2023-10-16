// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"context"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// GetBalance returns the balance for the given address.
func (gqh *IntegrationHandler) GetBalance(address sdktypes.AccAddress, denom string) (*banktypes.QueryBalanceResponse, error) {
	bankClient := gqh.network.GetBankClient()
	return bankClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: address.String(),
		Denom:   denom,
	})
}
