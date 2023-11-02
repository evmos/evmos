// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/x/evm/statedb"
)

// UnitTestNetwork is the implementation of the Network interface for unit tests.
// It embeds the IntegrationNetwork struct to reuse its methods and
// makes the App public for easier testing.
type UnitTestNetwork struct {
	IntegrationNetwork
	App *app.Evmos
}

var _ Network = &UnitTestNetwork{}

// NewUnitTestNetwork configures and initializes a new Evmos Network instance with
// the given configuration options. If no configuration options are provided
// it uses the default configuration.
//
// It panics if an error occurs.
// Note: Only uses for Unit Tests
func NewUnitTestNetwork(opts ...ConfigOption) *UnitTestNetwork {
	network := New(opts...)
	return &UnitTestNetwork{
		IntegrationNetwork: *network,
		App:                network.app,
	}
}

// GetStateDB returns the state database for the current block.
func (n *UnitTestNetwork) GetStateDB() *statedb.StateDB {
	headerHash := n.GetContext().HeaderHash()
	return statedb.New(
		n.GetContext(),
		n.App.EvmKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash.Bytes())),
	)
}
