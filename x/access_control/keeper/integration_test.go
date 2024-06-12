package keeper_test

import (
	"github.com/evmos/evmos/v18/contracts"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	evmosfactory "github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"testing"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

func TestKeeperIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

var _ = Describe("Access Control module tests", func() {
	var (
		nw   *network.UnitTestNetwork
		gh   grpc.Handler
		keys keyring.Keyring
		tf   evmosfactory.TxFactory
	)

	Context("using access control functionality", func() {
		var (
			admin keyring.Key
			//user  keyring.Key
			// initialized vars
			//gasPrice = math.NewInt(700_000_000)
			//gas      = uint64(500_000)
		)

		BeforeEach(func() {
			// setup network
			keys = keyring.New(2)
			admin = keys.GetKey(0)
			user = keys.GetKey(1)

			nw = network.NewUnitTestNetwork()
			gh = grpc.NewIntegrationHandler(nw)
			tf = evmosfactory.New(nw, gh)

			Expect(nw.NextBlock()).To(BeNil())

			// create a new contract

			_, err := tf.DeployContract(
				admin.Priv,
				evmtypes.EvmTxArgs{},
				factory.ContractDeploymentData{Contract: contracts.TokenFactoryCoinContract},
			)

			Expect(err).To(BeNil())
		})
	})

})
