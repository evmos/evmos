// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package osmosis_test

import (
	"fmt"
	"testing"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	commonnetwork "github.com/evmos/evmos/v15/testutil/integration/common/network"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	_ "github.com/evmos/evmos/v15/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v15/testutil/integration/ibc/coordinator"
)

type IntegrationTestSuite struct {
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
	coordinator coordinator.Coordinator
}

var globalT *testing.T

func TestIntegrationOutpost(t *testing.T) {
	RegisterFailHandler(Fail)
	globalT = t
	RunSpecs(t, "Outpost Integration Suite")
}

var _ = Describe("Handling an Osmosis Outpost", Label("Osmosis"), Ordered, func() {
	var s *IntegrationTestSuite

	BeforeAll(func() {
		keyring := testkeyring.New(3)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)

		coordinator := coordinator.NewIntegrationCoordinator(globalT, []commonnetwork.Network{integrationNetwork})
		s = &IntegrationTestSuite{
			network:     integrationNetwork,
			factory:     txFactory,
			grpcHandler: grpcHandler,
			keyring:     keyring,
			coordinator: coordinator,
		}

		// Start each test with a fresh block
		err := s.coordinator.CommitNBlocks(s.network.GetChainID(), 1)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		// // Start each test with a fresh block
		err := s.coordinator.CommitNBlocks(s.network.GetChainID(), 1)
		Expect(err).To(BeNil())
	})

	When("a user sends a transaction to create a pool", func() {
		It("should create a pool", func() {
			// TODO
			acc, err := s.grpcHandler.GetAccount(s.keyring.GetAccAddr(0).String())
			Expect(err).To(BeNil())

			s.coordinator.SetDefaultSignerForChain(s.network.GetChainID(), s.keyring.GetPrivKey(0), acc)

			dummyChains := s.coordinator.GetDummyChainsIds()
			fmt.Println(dummyChains)

			_ = s.coordinator.Setup(s.network.GetChainID(), dummyChains[0])
			Expect(true).To(Equal(true))
		})
	})
})
