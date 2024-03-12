// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper_test

import (
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSTRv2Tracking(t *testing.T) {
	// Run Ginkgo BDD tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "STRv2 Tracking Tests")
}

var _ = Describe("STRv2 Tracking", func() {
	var (
		keyring testkeyring.Keyring
		network *testnetwork.UnitTestNetwork
		handler grpc.Handler
		factory testfactory.TxFactory
	)

	BeforeAll(func() {
		keyring = testkeyring.New(2)
		network = testnetwork.NewUnitTestNetwork(
			testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		handler = grpc.NewIntegrationHandler(network)
		factory = testfactory.New(network, handler)
	})

	When("sending an EVM transaction", func() {
		Context("which interacts with a registered native token pair ERC-20 contract", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})

			It("should not fail if the address is already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})
})
