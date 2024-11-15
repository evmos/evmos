// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package grpc

import (
	"context"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

// GetBalanceFromEVM returns the balance for the given address.
func (gqh *IntegrationHandler) GetBalanceFromEVM(address sdktypes.AccAddress) (*evmtypes.QueryBalanceResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Balance(context.Background(), &evmtypes.QueryBalanceRequest{
		Address: common.BytesToAddress(address).Hex(),
	})
}
