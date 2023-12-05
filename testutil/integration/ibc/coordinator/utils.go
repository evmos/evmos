// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package coordinator

import (
	"strconv"
	"testing"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/evmos/evmos/v16/testutil/integration/common/network"
)

// getIBCChains returns a map of TestChain's for the given network interface.
func getIBCChains(t *testing.T, coord *ibctesting.Coordinator, chains []network.Network) map[string]*ibctesting.TestChain {
	ibcChains := make(map[string]*ibctesting.TestChain)
	for _, chain := range chains {
		ibcChains[chain.GetChainID()] = chain.GetIBCChain(t, coord)
	}
	return ibcChains
}

// generateDummyChains returns a map of dummy chains to complement IBC connections for integration tests.
func generateDummyChains(t *testing.T, coord *ibctesting.Coordinator, numberOfChains int) (map[string]*ibctesting.TestChain, []string) {
	ibcChains := make(map[string]*ibctesting.TestChain)
	ids := make([]string, numberOfChains)
	for i := 1; i <= numberOfChains; i++ {
		chainID := "dummychain-" + strconv.Itoa(i)
		ids[i-1] = chainID
		ibcChains[chainID] = ibctesting.NewTestChain(t, coord, chainID)
	}
	return ibcChains, ids
}

// mergeMaps merges two maps of TestChain's.
func mergeMaps(m1, m2 map[string]*ibctesting.TestChain) map[string]*ibctesting.TestChain {
	for k, v := range m2 {
		m1[k] = v
	}
	return m1
}
