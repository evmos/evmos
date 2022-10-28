package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/evmos/evmos/v9/tests/e2e/chain"
	"github.com/evmos/evmos/v9/tests/e2e/util"
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
	chains         []*chain.Chain
	dkrPool        *dockertest.Pool
	dkrNet         *dockertest.Network
	hermesResource *dockertest.Resource
	initResource   *dockertest.Resource
	valResources   map[string][]*dockertest.Resource
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

	s.chains = make([]*chain.Chain, 0, 1)

	// The e2e test flow is as follows:
	//
	// 1. Configure evmos chain with two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start the networks.
	// 3. Upgrade the network
	// 4. Execute various e2e tests
	s.loadUpgradeParams()
	s.configureDockerResources(chain.ChainAID, chain.ChainBID)

	s.configureChain(chain.ChainAID)

	s.T().Logf("running validators with version %s...", s.upgradeParams.PreUpgradeVersion)
	s.runValidators(s.chains[0], 0)
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

func (s *IntegrationTestSuite) configureDockerResources(chainIDOne, chainIDTwo string) {
	var err error
	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-%s-testnet", chainIDOne, chainIDTwo))
	s.Require().NoError(err)

	s.valResources = make(map[string][]*dockertest.Resource)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.upgradeParams.SkipCleanup {
		return
	}
	s.T().Log("tearing down e2e integration test suite...")

	for _, vr := range s.valResources {
		for _, r := range vr {
			s.Require().NoError(s.dkrPool.Purge(r))
		}
	}

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))

	for _, chain := range s.chains {
		os.RemoveAll(chain.ChainMeta.DataDir)
	}

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) runValidators(c *chain.Chain, portOffset int) {
	s.T().Logf("starting Evmos %s validator containers...", c.ChainMeta.ID)
	s.valResources[c.ChainMeta.ID] = make([]*dockertest.Resource, len(c.Validators))
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.Name,
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/evmos/.evmosd", val.ConfigDir),
			},
			Repository: "tharsishq/evmos",
			Tag:        s.upgradeParams.PreUpgradeVersion,
			Cmd: []string{
				"/usr/bin/evmosd",
				"start",
				"--home",
				"/evmos/.evmosd",
			},
		}

		// expose the first validator for debugging and communication
		if val.Index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 1317+portOffset)}},
				"6060/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6060+portOffset)}},
				"6061/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6061+portOffset)}},
				"6062/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6062+portOffset)}},
				"6063/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6063+portOffset)}},
				"6064/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6064+portOffset)}},
				"6065/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6065+portOffset)}},
				"9090/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 9090+portOffset)}},
				"26656/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26656+portOffset)}},
				"26657/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26657+portOffset)}},
			}
		}

		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[c.ChainMeta.ID][i] = resource
		s.T().Logf("started Evmos %s validator container: %s", c.ChainMeta.ID, resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := rpcClient.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"Evmos node failed to produce blocks",
	)
}

func (s *IntegrationTestSuite) configureChain(chainId string) {

	uid := os.Getuid()
	user := fmt.Sprintf("%d%s%d", uid, ":", uid)

	s.T().Logf("starting e2e infrastructure for chain-id: %s", chainId)
	tmpDir, err := ioutil.TempDir("", "evmos-e2e-testnet-")

	s.T().Logf("temp directory for chain-id %v: %v", chainId, tmpDir)
	s.Require().NoError(err)

	s.initResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainId),
			Repository: "evmos-e2e-chain-init",
			Tag:        "debug",
			NetworkID:  s.dkrNet.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", tmpDir),
				fmt.Sprintf("--chain-id=%s", chainId),
			},
			User: user,
			Mounts: []string{
				fmt.Sprintf("%s:%s", tmpDir, tmpDir),
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	var newChain chain.Chain

	fileName := fmt.Sprintf("%v/%v-encode", tmpDir, chainId)
	s.T().Logf("serialized init file for chain-id %v: %v", chainId, fileName)

	// loop through the reading and unmarshaling of the init file a total of maxRetries or until error is nil
	// without this, test attempts to unmarshal file before docker container is finished writing
	for i := 0; i < maxRetries; i++ {
		encJson, _ := os.ReadFile(fileName)
		err = json.Unmarshal(encJson, &newChain)
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			s.Require().NoError(err)
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	s.chains = append(s.chains, &newChain)
	s.Require().NoError(s.dkrPool.Purge(s.initResource))
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func (s *IntegrationTestSuite) initUpgrade() {
	// submit, deposit, and vote for upgrade proposal
	v := os.Getenv("POST_UPGRADE_VERSION")
	if v == "" {
		log.Fatal("no post-upgrade version specified")
	}
	s.submitProposal(s.chains[0], v)
	s.depositProposal(s.chains[0])
	s.voteProposal(s.chains[0])
	s.fundCommunityPool(s.chains[0])

	// wait till all chains halt at upgrade height
	for i := range s.chains[0].Validators {
		s.T().Logf("waiting to reach upgrade height on %s validator container: %s", s.chains[0].ChainMeta.ID, s.valResources[s.chains[0].ChainMeta.ID][i].Container.ID)
		s.Require().Eventually(
			func() bool {
				height, _ := s.chainStatus(s.valResources[s.chains[0].ChainMeta.ID][i].Container.ID)
				if height != 50 {
					s.T().Logf("current block height is %v, waiting for block 50 container: %s", height, s.valResources[s.chains[0].ChainMeta.ID][i].Container.ID)
				}
				return height == 50
			},
			2*time.Minute,
			5*time.Second,
		)
		s.T().Logf("reached upgrade height on %s validator container: %s", s.chains[0].ChainMeta.ID, s.valResources[s.chains[0].ChainMeta.ID][i].Container.ID)
	}

}

func (s *IntegrationTestSuite) stopAllNodeContainers() {
	// remove all containers so we can upgrade them to the new version
	for i := range s.chains[0].Validators {
		s.Require().NoError(s.dkrPool.RemoveContainerByName(s.valResources[s.chains[0].ChainMeta.ID][i].Container.Name))
	}
}

func (s *IntegrationTestSuite) upgrade() {
	// upgrade containers to the locally compiled daemon
	chain := s.chains[0]
	s.T().Logf("starting upgrade for chain-id: %s...", chain.ChainMeta.ID)
	for i, val := range chain.Validators {
		runOpts := &dockertest.RunOptions{
			Name:       val.Name,
			Repository: "evmos",
			Tag:        "debug",
			NetworkID:  s.dkrNet.Network.ID,
			User:       "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/evmos/.evmosd", val.ConfigDir),
			},
			Cmd: []string{
				"/usr/bin/evmosd",
				"start",
				"--home",
				"/evmos/.evmosd",
			},
		}
		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[chain.ChainMeta.ID][i] = resource
		s.T().Logf("started Evmos upgraded %s validator container: %s", chain.ChainMeta.ID, resource.Container.ID)
	}

	// check that we are hitting blocks again
	for i := range chain.Validators {
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

func (s *IntegrationTestSuite) replaceGenesis(newGenesis []byte) {
	chain := s.chains[0]
	if len(newGenesis) == 0 {
		s.T().Logf("Do not perform genesis migration, since its not available on chain-id: %s...", chain.ChainMeta.ID)
		return
	}

	s.T().Logf("replacing genesis files for chain-id: %s...", chain.ChainMeta.ID)
	// write the updated genesis file to each validator
	for _, val := range chain.Validators {
		err := util.WriteFile(filepath.Join(val.ConfigDir, "config", "genesis.json"), newGenesis)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) replaceContainers() {
	// upgrade containers to the locally compiled daemon
	chain := s.chains[0]
	s.T().Logf("starting upgrade for chain-id: %s...", chain.ChainMeta.ID)
	for i, val := range chain.Validators {
		runOpts := &dockertest.RunOptions{
			Name:       val.Name,
			Repository: "evmos",
			Tag:        "debug",
			NetworkID:  s.dkrNet.Network.ID,
			User:       "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/evmos/.evmosd", val.ConfigDir),
			},
			Cmd: []string{
				"tail", "-f", "/dev/null",
			},
		}
		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[chain.ChainMeta.ID][i] = resource
		s.T().Logf("started Evmos upgraded %s validator container: %s", chain.ChainMeta.ID, resource.Container.ID)
	}
}
