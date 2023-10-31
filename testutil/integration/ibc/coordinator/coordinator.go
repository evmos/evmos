// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package coordinator

import (
	"testing"
	"time"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/evmos/evmos/v15/testutil/integration/common/network"
	ibcchain "github.com/evmos/evmos/v15/testutil/integration/ibc/chain"
)

// Coordinator is the interface that defines the methods that are used to
// coordinate the execution of the IBC relayer.
type Coordinator interface {
	// IncrementTime iterates through all the TestChain's and increments their current header time
	// by 5 seconds.
	IncrementTime()
	// UpdateTime updates all clocks for the TestChains to the current global time.
	UpdateTime()
	// UpdateTimeForChain updates the clock for a specific chain.
	UpdateTimeForChain(chainID string)
	// GetChain returns the TestChain for a given chainID.
	GetChain(chainID string) ibcchain.Chain
	// Setup constructs a TM client, connection, and channel on both chains provided. It will
	// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
	// for both chains. The channels created are connected to the ibc-transfer application.
	Setup(path *ibctesting.Path)
	// CommitNBlocks commits n blocks on the chain with the given chainID.
	CommitNBlocks(chainID string, n uint64) error
}

// TODO: Replace for a config
var (
	AmountOfDummyChains = 2
	GlobalTime          = time.Date(time.Now().Year()+1, 1, 2, 0, 0, 0, 0, time.UTC)
)

var _ Coordinator = (*IntegrationCoordinator)(nil)

// IntegrationCoordinator is a testing struct which contains N TestChain's. It handles keeping all chains
// in sync with regards to time.
type IntegrationCoordinator struct {
	coord *ibctesting.Coordinator
}

// NewIntegrationCoordinator returns a new IntegrationCoordinator with N TestChain's.
func NewIntegrationCoordinator(t *testing.T, preConfiguredChains []network.Network) *IntegrationCoordinator {
	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: GlobalTime,
	}
	ibcChains := getIBCChains(t, coord, preConfiguredChains)
	dummyChains := generateDummyChains(t, coord, AmountOfDummyChains)
	totalChains := mergeMaps(ibcChains, dummyChains)
	coord.Chains = totalChains
	return &IntegrationCoordinator{
		coord: coord,
	}
}

// getIBCChains returns a map of TestChain's for the given network interface.
func getIBCChains(t *testing.T, coord *ibctesting.Coordinator, chains []network.Network) map[string]*ibctesting.TestChain {
	ibcChains := make(map[string]*ibctesting.TestChain)
	for _, chain := range chains {
		ibcChains[chain.GetChainID()] = chain.GetIBCChain(t, coord)
	}
	return ibcChains
}

// generateDummyChains returns a map of dummy chains to complement IBC connections for integration tests.
func generateDummyChains(t *testing.T, coord *ibctesting.Coordinator, numberOfChains int) map[string]*ibctesting.TestChain {
	ibcChains := make(map[string]*ibctesting.TestChain)
	for i := 1; i <= numberOfChains; i++ {
		chainID := ibctesting.GetChainID(i)
		ibcChains[chainID] = ibctesting.NewTestChain(t, coord, chainID)
	}
	return ibcChains
}

// mergeMaps merges two maps of TestChain's.
func mergeMaps(m1, m2 map[string]*ibctesting.TestChain) map[string]*ibctesting.TestChain {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}

// GetChain returns the TestChain for a given chainID.
func (c *IntegrationCoordinator) GetChain(chainID string) ibcchain.Chain {
	return c.coord.Chains[chainID]
}

// IncrementTime iterates through all the TestChain's and increments their current header time
// by 5 seconds.
func (c *IntegrationCoordinator) IncrementTime() {
	c.coord.IncrementTime()
}

// UpdateTime updates all clocks for the TestChains to the current global time.
func (c *IntegrationCoordinator) UpdateTime() {
	c.coord.UpdateTime()
}

// UpdateTimeForChain updates the clock for a specific chain.
func (c *IntegrationCoordinator) UpdateTimeForChain(chainID string) {
	chain := c.coord.GetChain(chainID)
	c.coord.UpdateTimeForChain(chain)
}

// Setup constructs a TM client, connection, and channel on both chains provided. It will
// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
// for both chains. The channels created are connected to the ibc-transfer application.
func (c *IntegrationCoordinator) Setup(path *ibctesting.Path) {
	c.coord.Setup(path)
}

// CommitNBlocks commits n blocks on the chain with the given chainID.
func (c *IntegrationCoordinator) CommitNBlocks(chainID string, n uint64) error {
	chain := c.coord.GetChain(chainID)
	c.coord.CommitNBlocks(chain, n)
	return nil
}
