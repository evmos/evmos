package distribution_test

import (
	"testing"

	"github.com/evmos/evmos/v16/precompiles/distribution"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"

	// //nolint:revive // dot imports are fine for Ginkgo
	// . "github.com/onsi/ginkgo/v2"
	// //nolint:revive // dot imports are fine for Ginkgo
	// . "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *distribution.Precompile
	bondDenom  string
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	// RegisterFailHandler(Fail)
	// RunSpecs(t, "Distribution Precompile Suite")
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	ctx := integrationNetwork.GetContext()
	sk := integrationNetwork.App.StakingKeeper
	bondDenom, err := sk.BondDenom(ctx)
	s.Require().NoError(err, "failed to get bond denom")
	s.Require().NotEmpty(bondDenom, "bond denom cannot be empty")

	s.bondDenom = bondDenom
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = integrationNetwork
	s.precompile = s.setupDistrPrecompile()
}

// setupBankPrecompile is a helper function to set up an instance of the Bank precompile for
// a given token denomination.
func (s *PrecompileTestSuite) setupDistrPrecompile() *distribution.Precompile {
	precompile, err := distribution.NewPrecompile(
		s.network.App.DistrKeeper,
		s.network.App.StakingKeeper,
		s.network.App.AuthzKeeper,
	)

	s.Require().NoError(err, "failed to create bank precompile")

	return precompile
}
