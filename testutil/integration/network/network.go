// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"math/big"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v15/x/feemarket/types"
	infltypes "github.com/evmos/evmos/v15/x/inflation/types"
	revtypes "github.com/evmos/evmos/v15/x/revenue/v1/types"
)

// Network is the interface that wraps the methods to interact with integration test network.
//
// It was designed to avoid users to access module's keepers directly and force integration tests
// to be closer to the real user's behavior.
type Network interface {
	GetContext() sdktypes.Context
	GetChainID() string
	GetDenom() string
	GetValidators() []stakingtypes.Validator
	GetValSet() *tmtypes.ValidatorSet

	NextBlock() error

	BroadcastTxSync(txBytes []byte) (abcitypes.ResponseDeliverTx, error)
	Simulate(txBytes []byte) (*txtypes.SimulateResponse, error)
}

// EvmosNetwork is the interface that wraps the methods to interact with integration test network.
//
// It was designed to avoid users to access module's keepers directly and force integration tests
// to be closer to the real user's behavior.
type EvmosNetwork interface {
	Network
	GetEIP155ChainID() *big.Int

	GetEvmClient() evmtypes.QueryClient
	GetRevenueClient() revtypes.QueryClient
	GetInflationClient() infltypes.QueryClient
	GetBankClient() banktypes.QueryClient
	GetFeeMarketClient() feemarkettypes.QueryClient
	GetAuthClient() authtypes.QueryClient
	GetStakingClient() stakingtypes.QueryClient

	// Because to update the module params on a conventional manner governance
	// would be require, we should provide an easier way to update the params
	UpdateRevenueParams(params revtypes.Params) error
	UpdateInflationParams(params infltypes.Params) error
	UpdateEvmParams(params evmtypes.Params) error
}

type BaseNetwork struct {
	cfg        Config
	ctx        sdktypes.Context
	validators []stakingtypes.Validator
	valSet     *tmtypes.ValidatorSet
	app        *baseapp.BaseApp
}

// GetContext returns the network's context
func (n *BaseNetwork) GetContext() sdktypes.Context {
	return n.ctx
}

// GetChainID returns the network's chainID
func (n *BaseNetwork) GetChainID() string {
	return n.cfg.chainID
}

// GetDenom returns the network's denom
func (n *BaseNetwork) GetDenom() string {
	return n.cfg.denom
}

// GetValidators returns the network's validators
func (n *BaseNetwork) GetValidators() []stakingtypes.Validator {
	return n.validators
}

// GetValSet returns the network's validator set
func (n *BaseNetwork) GetValSet() *tmtypes.ValidatorSet {
	return n.valSet
}

// BroadcastTxSync broadcasts the given txBytes to the network and returns the response.
// TODO - this should be change to gRPC
func (n *BaseNetwork) BroadcastTxSync(txBytes []byte) (abcitypes.ResponseDeliverTx, error) {
	req := abcitypes.RequestDeliverTx{Tx: txBytes}
	return n.app.DeliverTx(req), nil
}

// Simulate simulates the given txBytes to the network and returns the simulated response.
// TODO - this should be change to gRPC
func (n *BaseNetwork) Simulate(txBytes []byte) (*txtypes.SimulateResponse, error) {
	gas, result, err := n.app.Simulate(txBytes)
	if err != nil {
		return nil, err
	}
	return &txtypes.SimulateResponse{
		GasInfo: &gas,
		Result:  result,
	}, nil
}
