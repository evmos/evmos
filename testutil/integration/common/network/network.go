// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"testing"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

// Network is the interface that wraps the commong methods to interact with integration test network.
//
// It was designed to avoid users to access module's keepers directly and force integration tests
// to be closer to the real user's behavior.
type Network interface {
	GetContext() sdktypes.Context
	GetChainID() string
	GetDenom() string
	GetValidators() []stakingtypes.Validator

	NextBlock() error

	// Clients
	GetAuthClient() authtypes.QueryClient
	GetStakingClient() stakingtypes.QueryClient
	GetBankClient() banktypes.QueryClient

	BroadcastTxSync(txBytes []byte) (abcitypes.ResponseDeliverTx, error)
	Simulate(txBytes []byte) (*txtypes.SimulateResponse, error)

	// GetIBCChain returns the IBC test chain.
	// NOTE: this is only used for testing IBC related functionality.
	// The idea is to depecrate this eventually.
	GetIBCChain(t *testing.T, coord *ibctesting.Coordinator) *ibctesting.TestChain
}
