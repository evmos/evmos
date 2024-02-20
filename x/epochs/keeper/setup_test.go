package keeper_test

import (
	"testing"

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

func (s *KeeperTestSuite) SetupTest() {
	keys := keyring.New(1)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

    // nw.NextBlock()
    
	identifiers := []string{types.WeekEpochID, types.DayEpochID}
	for _, identifier := range identifiers {
        ctx := nw.GetContext()
		epoch, found := nw.App.EpochsKeeper.GetEpochInfo(ctx, identifier)
		s.Require().True(found)
		epoch.StartTime = ctx.BlockTime()
		epoch.CurrentEpochStartHeight = ctx.BlockHeight()
		nw.App.EpochsKeeper.SetEpochInfo(ctx, epoch)
	}

	s.keyring = keys
	s.network = nw
	s.handler = gh
	s.factory = tf
}
