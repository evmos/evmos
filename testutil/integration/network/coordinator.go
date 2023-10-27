package network

import (
	"testing"
	"time"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

type Coordinator interface {
	IncrementTime()
	IncrementTimeBy(time.Duration)
	UpdateTime()
	UpdateTimeForChain(string)
	Setup(*ibctesting.Path)
	// Maybe this should return a IBCChain Interface
	GetChain(string) *ibctesting.TestChain
	GetChains() map[string]*ibctesting.TestChain
	SetEvmosChains(chains []EvmosIBCNetwork)
}

var _ Coordinator = (*IntegrationCoordinator)(nil)

type IntegrationCoordinator struct {
	coord *ibctesting.Coordinator
}

func newCoordinator(cfg IBCNetworkConfig) *IntegrationCoordinator {
	// Dummy testing object since we don't need it
	t := &testing.T{}
	// TODO - Start the network
	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: time.Date(time.Now().Year()+1, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	coord.Chains = createChains(coord, cfg.numberOfChains)

	return &IntegrationCoordinator{
		coord: coord,
	}
}

func createChains(coord *ibctesting.Coordinator, amountOfChains int) map[string]*ibctesting.TestChain {
	chains := make(map[string]*ibctesting.TestChain, amountOfChains)
	for i := 1; i <= amountOfChains; i++ {
		chainID := ibctesting.GetChainID(i)
		chains[chainID] = ibctesting.NewTestChain(coord.T, coord, chainID)
	}
	return chains
}

func (c *IntegrationCoordinator) IncrementTime() {
	c.coord.IncrementTime()
}

func (c *IntegrationCoordinator) IncrementTimeBy(duration time.Duration) {
	c.coord.IncrementTimeBy(duration)
}

func (c *IntegrationCoordinator) UpdateTime() {
	c.coord.UpdateTime()
}

func (c *IntegrationCoordinator) UpdateTimeForChain(chainID string) {
	chain := c.coord.Chains[chainID]
	c.coord.UpdateTimeForChain(chain)
}

// Setup constructs a TM client, connection, and channel on both chains provided. It will
// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
// for both chains. The channels created are connected to the ibc-transfer application.
func (c *IntegrationCoordinator) Setup(path *ibctesting.Path) {
	c.coord.Setup(path)
}

func (c *IntegrationCoordinator) GetChain(chainID string) *ibctesting.TestChain {
	return c.coord.Chains[chainID]
}

func (c *IntegrationCoordinator) GetChains() map[string]*ibctesting.TestChain {
	return c.coord.Chains
}

func (c *IntegrationCoordinator) SetEvmosChains(chains []EvmosIBCNetwork) {
	for _, chain := range chains {
		c.coord.Chains[chain.GetChainID()] = chain.getIBCChain(c.coord)
	}
}
