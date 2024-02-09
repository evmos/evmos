package werc20_test

import (

	//nolint:revive // dot imports are fine for Ginkgo
	"testing"

	. "github.com/onsi/ginkgo/v2"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// var s *PrecompileTestSuite

func TestPrecompileTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Precompile Suite")
}

// func (s *PrecompileTestSuite) SetupTest() {
// 	keyring := testkeyring.New(2)
// 	integrationNetwork := network.NewUnitTestNetwork(
// 		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
// 	)
// 	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
// 	txFactory := factory.New(integrationNetwork, grpcHandler)

// 	// s.bondDenom = integrationNetwork.GetDenom()
// 	s.factory = txFactory
// 	s.grpcHandler = grpcHandler
// 	s.keyring = keyring
// 	s.network = integrationNetwork
// }
