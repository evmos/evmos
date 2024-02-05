package v17_test

import (
	"github.com/ethereum/go-ethereum/common"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
	"testing"
)

// TestSTRv2Migration runs the Ginkgo BDD tests for the migration logic
// associated with the Single Token Representation v2.
func TestSTRv2Migration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "STR v2 Migration Suite")
}

type ConvertERC20CoinsTestSuite struct {
	// TODO: Can be removed eventually because it's only used for the require stuff which we are not using in integration tests
	suite.Suite

	keyring testkeyring.Keyring
	network *network.UnitTestNetwork
	handler grpc.Handler
	factory testfactory.TxFactory

	// erc20Contract is the address of the deployed ERC-20 contract for testing purposes.
	erc20Contract common.Address
	// nativeTokenPair is a registered token pair for a native Coin.
	nativeTokenPair erc20types.TokenPair
	// nonNativeTokenPair is a registered token pair for an ERC-20 native asset.
	nonNativeTokenPair erc20types.TokenPair
	// wevmosContract is the address of the deployed WEVMOS contract for testing purposes.
	wevmosContract common.Address
}

// NOTE: For these tests it's mandatory to run them ORDERED!
var _ = When("testing the STR v2 migration", Ordered, func() {
	var ts *ConvertERC20CoinsTestSuite

	BeforeAll(func() {
		// NOTE: In the setup function we are creating a custom genesis state for the integration network
		// which contains balances for two accounts in different denominations.
		// There is also an ERC-20 smart contract deployed and some tokens minted for each of the accounts.
		// The balances are split between both token representations (IBC coin and ERC-20 token).
		//
		// This genesis state is the starting point to check the migration for the introduction of STR v2.
		// This should ONLY convert native coins for now, which means that the native ERC-20s should be untouched.
		// All native IBC coins should be converted to the native representation and the full balance should be returned
		// both by the bank and the ERC-20 contract.
		// There should be a new ERC-20 EVM extension registered and the ERC-20 contract should be able to be called
		// after being deleted and re-registered as a precompile.
		var err error
		ts, err = NewConvertERC20CoinsTestSuite()
		Expect(err).ToNot(HaveOccurred(), "failed to create test suite")
	})

	When("checking the genesis state of the network", Ordered, func() {
		It("should have registered a native token pair", func() {
			res, err := ts.handler.GetTokenPairs()
			Expect(err).ToNot(HaveOccurred(), "failed to get token pairs")
			Expect(res.TokenPairs).To(HaveLen(1), "unexpected number of token pairs")
			Expect(res.TokenPairs[0].Denom).To(Equal(XMPL), "expected different denom")

			// Assign the native token pair to the test suite for later use.
			ts.nativeTokenPair = res.TokenPairs[0]
		})
	})

	When("preparing the network state", Ordered, func() {
		It("should run the preparation without an error", func() {
			var err error
			ts, err = PrepareNetwork(ts)
			Expect(err).ToNot(HaveOccurred(), "failed to prepare network state")
		})

		It("should have registered another token pair", func() {
			res, err := ts.handler.GetTokenPairs()
			Expect(err).ToNot(HaveOccurred(), "failed to get token pairs")
			Expect(res.TokenPairs).To(HaveLen(2), "unexpected number of token pairs")
			Expect(res.TokenPairs).To(ContainElement(ts.nonNativeTokenPair), "non-native token pair not found")
		})
	})

	It("should migrate migrate without an error", func() {
		Expect(false).To(BeTrue(), "not implemented")
	})
})
