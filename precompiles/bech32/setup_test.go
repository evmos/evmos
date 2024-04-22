package bech32_test

import (
	"testing"

	"github.com/evmos/evmos/v18/precompiles/bech32"

	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

// PrecompileTestSuite is the implementation of the TestSuite interface for ERC20 precompile
// unit tests.
type PrecompileTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	keyring testkeyring.Keyring

	precompile *bech32.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	s.keyring = keyring
	s.network = integrationNetwork

	precompile, err := bech32.NewPrecompile(6000)
	s.Require().NoError(err, "failed to create bech32 precompile")

	s.precompile = precompile
}
