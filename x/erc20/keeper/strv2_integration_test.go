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

type STRv2TrackingSuite struct {
	keyring testkeyring.Keyring
	network *testnetwork.UnitTestNetwork
	handler grpc.Handler
	factory testfactory.TxFactory
}

const (
	deployerIdx = 0
)

var _ = Describe("STRv2 Tracking", func() {
	var s *STRv2TrackingSuite

	BeforeEach(func() {
		s = SetupTest()

		deployer := s.keyring.GetKey(deployerIdx)

		nativeERC20Addr, err := s.DeployERC20Contract(deployer, ERC20ConstructorArgs{
			Name:     "TestToken",
			Symbol:   "TTK",
			Decimals: 18,
		})
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC-20 contract")

		_, err = s.network.App.Erc20Keeper.RegisterERC20(s.network.GetContext(), nativeERC20Addr)
		Expect(err).ToNot(HaveOccurred(), "failed to register token pair")
	})

	When("sending an EVM transaction", func() {
		Context("which interacts with a registered native token pair ERC-20 contract", func() {
			Context("in a direct call to the token pair contract", func() {
				It("should add the address to the store if it is not already stored", func() {
					Expect(true).To(BeFalse(), "not implemented")
				})

				It("should not fail if the address is already stored", func() {
					Expect(true).To(BeFalse(), "not implemented")
				})
			})

			Context("in a call to the token pair contract from another contract", func() {
				It("should add the address to the store if it is not already stored", func() {
					Expect(true).To(BeFalse(), "not implemented")
				})

				It("should not fail if the address is already stored", func() {
					Expect(true).To(BeFalse(), "not implemented")
				})
			})
		})

		Context("which interacts with a registered non-native token pair ERC-20 contract", func() {
			It("should not add the address to the store", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})

		Context("which interacts with an unregistered ERC-20 contract", func() {
			It("should not add the address to the store", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})

	When("when receiving an incoming IBC transfer", func() {
		Context("for a registered IBC asset", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})

			It("should not fail if the address is already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})

		Context("for an unregistered IBC asset", func() {
			It("should not add the address to the store", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})

	When("sending an IBC transfer", func() {
		Context("for a registered IBC asset", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})

			It("should not fail if the address is already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})

		Context("for an unregistered IBC asset", func() {
			It("should not add the address to the store", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})

	When("manually converting", func() {
		Context("a registered coin into its ERC-20 representation", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})

			It("should not fail if the address is already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})

		Context("a registered ERC-20 representation into its native coin", func() {
			It("should add the address to the store if it is not already stored", func() {
				Expect(true).To(BeFalse(), "not implemented")
			})
		})
	})
})
