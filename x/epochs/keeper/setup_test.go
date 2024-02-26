package keeper_test

import (
	"testing"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	. "github.com/onsi/ginkgo/v2"

	"github.com/evmos/evmos/v16/x/epochs/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"
)

const (
	day             = time.Hour * 24
	week            = time.Hour * 24 * 7
	month           = time.Hour * 24 * 31
	monthIdentifier = "month"
)

// KeeperTestSuite is the implementation of the test suite for the
// Epochs module.
type KeeperTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

// SetupTest is the setup function for epoch module tests. If epochsInfo is provided empty
// the default genesis for the epoch module is used.
func (s *KeeperTestSuite) SetupTest(epochsInfo []types.EpochInfo) sdktypes.Context {
	keys := keyring.New(1)

	customGenesis := network.CustomGenesisState{}
	epochsGenesis := types.DefaultGenesisState()

	if len(epochsInfo) > 0 {
		epochsGenesis = types.NewGenesisState(epochsInfo)
	}

	customGenesis[types.ModuleName] = epochsGenesis

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	)

	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	s.keyring = keys
	s.network = nw
	s.handler = gh
	s.factory = tf

	return nw.GetContext()
}
