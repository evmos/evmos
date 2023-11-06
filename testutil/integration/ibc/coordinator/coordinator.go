// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package coordinator

import (
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	evmosibc "github.com/evmos/evmos/v15/ibc/testing"
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
	// GetDummyChainsIds returns the chainIDs for all dummy chains.
	GetDummyChainsIds() []string
	// SetDefaultSignerForChain sets the default signer for the chain with the given chainID.
	SetDefaultSignerForChain(chainID string, priv cryptotypes.PrivKey, acc authtypes.AccountI)
	// Setup constructs a TM client, connection, and channel on both chains provided. It will
	// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
	// for both chains. The channels created are connected to the ibc-transfer application.
	Setup(src, dst string) IBCConnection
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
// NOTE: When using the coordinator, it is important to commit blocks through the coordinator and not
// through the network interface directly. This is because the coordinator does not keep the context in
// sync with the network interface.
type IntegrationCoordinator struct {
	coord          *ibctesting.Coordinator
	dummyChainsIds []string
}

// NewIntegrationCoordinator returns a new IntegrationCoordinator with N TestChain's.
func NewIntegrationCoordinator(t *testing.T, preConfiguredChains []network.Network) *IntegrationCoordinator {
	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: GlobalTime,
	}
	ibcChains := getIBCChains(t, coord, preConfiguredChains)
	dummyChains, dummyChainsIds := generateDummyChains(t, coord, AmountOfDummyChains)
	totalChains := mergeMaps(ibcChains, dummyChains)
	coord.Chains = totalChains
	return &IntegrationCoordinator{
		coord:          coord,
		dummyChainsIds: dummyChainsIds,
	}
}

// GetChain returns the TestChain for a given chainID.
func (c *IntegrationCoordinator) GetChain(chainID string) ibcchain.Chain {
	return c.coord.Chains[chainID]
}

// GetDummyChainsIds returns the chainIDs for all dummy chains.
func (c *IntegrationCoordinator) GetDummyChainsIds() []string {
	return c.dummyChainsIds
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

// SetDefaultSignerForChain sets the default signer for the chain with the given chainID.
func (c *IntegrationCoordinator) SetDefaultSignerForChain(chainID string, priv cryptotypes.PrivKey, acc authtypes.AccountI) {
	chain := c.coord.GetChain(chainID)
	chain.SenderPrivKey = priv
	chain.SenderAccount = acc
	chain.SenderAccounts = []ibctesting.SenderAccount{{SenderPrivKey: priv, SenderAccount: acc}}
}

// Setup constructs a TM client, connection, and channel on both chains provided. It will
// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
// for both chains. The channels created are connected to the ibc-transfer application.
func (c *IntegrationCoordinator) Setup(a, b string) IBCConnection {
	chainA := c.coord.GetChain(a)
	chainB := c.coord.GetChain(b)

	path := evmosibc.NewTransferPath(chainA, chainB)
	evmosibc.SetupPath(c.coord, path)

	return IBCConnection{
		EndpointA: Endpoint{
			ChainID:      a,
			ClientID:     path.EndpointA.ClientID,
			ConnectionID: path.EndpointA.ConnectionID,
			ChannelID:    path.EndpointA.ChannelID,
			PortID:       path.EndpointA.ChannelConfig.PortID,
		},
		EndpointB: Endpoint{
			ChainID:      b,
			ClientID:     path.EndpointB.ClientID,
			ConnectionID: path.EndpointB.ConnectionID,
			ChannelID:    path.EndpointB.ChannelID,
			PortID:       path.EndpointB.ChannelConfig.PortID,
		},
	}
}

// CommitNBlocks commits n blocks on the chain with the given chainID.
func (c *IntegrationCoordinator) CommitNBlocks(chainID string, n uint64) error {
	chain := c.coord.GetChain(chainID)
	c.coord.CommitNBlocks(chain, n)
	return nil
}
