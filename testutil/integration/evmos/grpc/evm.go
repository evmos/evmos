// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetEvmAccount returns the EVM account for the given address.
func (gqh *IntegrationHandler) GetEvmAccount(address common.Address) (*evmtypes.QueryAccountResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Account(context.Background(), &evmtypes.QueryAccountRequest{
		Address: address.String(),
	})
}

// EstimateGas returns the estimated gas for the given call args.
func (gqh *IntegrationHandler) EstimateGas(args []byte, gasCap uint64) (*evmtypes.EstimateGasResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	res, err := evmClient.EstimateGas(context.Background(), &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: gasCap,
	})
	if err != nil {
		return nil, err
	}
	// handle case where there's a revert related error
	if len(res.VmError) > 0 {
		if res.VmError != vm.ErrExecutionReverted.Error() {
			return nil, status.Error(codes.Internal, res.VmError)
		}
		if len(res.Ret) == 0 {
			return nil, errors.New(res.VmError)
		}
		return nil, evmtypes.NewExecErrorWithReason(res.Ret)
	}

	return res, err
}

// GetEvmParams returns the EVM module params.
func (gqh *IntegrationHandler) GetEvmParams() (*evmtypes.QueryParamsResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Params(context.Background(), &evmtypes.QueryParamsRequest{})
}
