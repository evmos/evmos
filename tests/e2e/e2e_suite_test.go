package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"

	"github.com/evmos/evmos/v9/tests/e2e/upgrade"
	"github.com/evmos/evmos/v9/tests/e2e/util"
)

const (
	localRepository = "evmos"
	initialTag      = "initial"
)

var (
	// common
	maxRetries = 10 // max retries for json unmarshalling
)

type upgradeParams struct {
	PreUpgradeVersion  string
	PostUpgradeVersion string
	MigrateGenesis     bool
	SkipCleanup        bool
}

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs        []string
	upgradeManager *upgrade.Manager
	hermesResource *dockertest.Resource
	upgradeParams  upgradeParams
}

type status struct {
	LatestHeight string `json:"latest_block_height"`
}

type syncInfo struct {
	SyncInfo status `json:"SyncInfo"`
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var err error
	s.upgradeManager, err = upgrade.NewManager()
	s.NoError(err, "upgrade manager creation error")

}

func (s *IntegrationTestSuite) loadUpgradeParams() {
	preV := os.Getenv("PRE_UPGRADE_VERSION")
	if preV == "" {
		s.Fail("no pre-upgrade version specified")
	}
	s.upgradeParams.PreUpgradeVersion = preV

	postV := os.Getenv("POST_UPGRADE_VERSION")
	if postV == "" {
		s.Fail("no post-upgrade version specified")
	}
	s.upgradeParams.PostUpgradeVersion = postV

	migrateGenFlag := os.Getenv("MIGRATE_GENESIS")
	migrateGenesis, err := strconv.ParseBool(migrateGenFlag)
	s.Require().NoError(err, "invalid migrate genesis flag")
	s.upgradeParams.MigrateGenesis = migrateGenesis

	skipFlag := os.Getenv("E2E_SKIP_CLEANUP")
	skipCleanup, err := strconv.ParseBool(skipFlag)
	s.Require().NoError(err, "invalid skip cleanup flag")
	s.upgradeParams.SkipCleanup = skipCleanup
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.upgradeParams.SkipCleanup {
		return
	}
	s.T().Log("tearing down e2e integration test suite...")

	s.Require().NoError(s.upgradeManager.KillCurrentNode())

	s.Require().NoError(s.upgradeManager.RemoveNetwork())
	// TODO: cleanup ./build/

}

func (s *IntegrationTestSuite) initUpgrade() {

}

func (s *IntegrationTestSuite) upgrade() {

		s.Require().Eventually(
			func() bool {
				height, _ := s.chainStatus(s.valResources[chain.ChainMeta.ID][i].Container.ID)
				s.Require().Greater(height, 50)
				if height <= 70 {
					fmt.Printf("current block height is %v, waiting to hit blocks\n", height)
				}
				return height > 70
			},
			2*time.Minute,
			5*time.Second,
		)
		s.T().Logf("upgrade successful on %s validator container: %s", chain.ChainMeta.ID, s.valResources[chain.ChainMeta.ID][i].Container.ID)
	}

}



