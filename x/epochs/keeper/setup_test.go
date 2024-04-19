package keeper_test

import (
	"testing"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	evm "github.com/evmos/evmos/v17/x/evm/types"

	"github.com/evmos/evmos/v17/app"
	"github.com/evmos/evmos/v17/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.Evmos
	queryClientEvm evm.QueryClient
	queryClient    types.QueryClient
	consAddress    sdk.ConsAddress
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest()
}
