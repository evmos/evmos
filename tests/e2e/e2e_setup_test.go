package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"

	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/tharsis/evmos/v4/tests/e2e/chain"
)

var (
	// common
	maxRetries             = 10 // max retries for json unmarshalling
	validatorConfigsChainA = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
	}
	validatorConfigsChainB = []*chain.ValidatorConfig{
		{
			Pruning:            "default",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "nothing",
			PruningKeepRecent:  "0",
			PruningInterval:    "0",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
		{
			Pruning:            "custom",
			PruningKeepRecent:  "10000",
			PruningInterval:    "13",
			SnapshotInterval:   1500,
			SnapshotKeepRecent: 2,
		},
	}
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs        []string
	chains         []*chain.Chain
	dkrPool        *dockertest.Pool
	dkrNet         *dockertest.Network
	hermesResource *dockertest.Resource
	initResource   *dockertest.Resource
	valResources   map[string][]*dockertest.Resource
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
	s.configureDockerResources(chain.ChainAID, chain.ChainBID)

	s.configureChain(chain.ChainAID, validatorConfigsChainA)
	s.runValidators(s.chains[0], 0)
	s.initUpgrade()
	s.upgrade()
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
	if str := os.Getenv("EVMOS_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
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
	s.T().Logf("starting Evmos %s validator containers...", c.ChainMeta.Id)
	s.valResources[c.ChainMeta.Id] = make([]*dockertest.Resource, len(c.Validators))
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.Name,
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/evmos/.evmosd", val.ConfigDir),
			},
			Repository: "tharsishq/evmos",
			Tag:        "v3.0.2",
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

		s.valResources[c.ChainMeta.Id][i] = resource
		s.T().Logf("started Evmos %s validator container: %s", c.ChainMeta.Id, resource.Container.ID)
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

func (s *IntegrationTestSuite) configureChain(chainId string, validatorConfigs []*chain.ValidatorConfig) {

	s.T().Logf("starting e2e infrastructure for chain-id: %s", chainId)
	tmpDir, err := ioutil.TempDir("", "evmos-e2e-testnet-")

	s.T().Logf("temp directory for chain-id %v: %v", chainId, tmpDir)
	s.Require().NoError(err)

	b, err := json.Marshal(validatorConfigs)
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
				fmt.Sprintf("--config=%s", b),
			},
			User: "root:root",
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
	s.submitProposal(s.chains[0])
	s.depositProposal(s.chains[0])
	s.voteProposal(s.chains[0])

	// wait till all chains halt at upgrade height
	for i := range s.chains[0].Validators {
		s.T().Logf("waiting to reach upgrade height on %s validator container: %s", s.chains[0].ChainMeta.Id, s.valResources[s.chains[0].ChainMeta.Id][i].Container.ID)
		s.Require().Eventually(
			func() bool {
				height, _ := s.chainStatus(s.valResources[s.chains[0].ChainMeta.Id][i].Container.ID)
				if height != 75 {
					s.T().Logf("current block height is %v, waiting for block 75 container: %s", height, s.valResources[s.chains[0].ChainMeta.Id][i].Container.ID)
				}
				return height == 75
			},
			2*time.Minute,
			5*time.Second,
		)
		s.T().Logf("reached upgrade height on %s validator container: %s", s.chains[0].ChainMeta.Id, s.valResources[s.chains[0].ChainMeta.Id][i].Container.ID)
	}

	// remove all containers so we can upgrade them to the new version
	for i := range s.chains[0].Validators {
		s.Require().NoError(s.dkrPool.RemoveContainerByName(s.valResources[s.chains[0].ChainMeta.Id][i].Container.Name))
	}

}

func (s *IntegrationTestSuite) upgrade() {
	// upgrade containers to the locally compiled daemon
	chain := s.chains[0]
	s.T().Logf("starting upgrade for chain-id: %s...", chain.ChainMeta.Id)
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

		s.valResources[chain.ChainMeta.Id][i] = resource
		s.T().Logf("started Evmos upgraded %s validator container: %s", chain.ChainMeta.Id, resource.Container.ID)
	}

	// check that we are hitting blocks again
	for i := range chain.Validators {
		s.Require().Eventually(
			func() bool {
				height, _ := s.chainStatus(s.valResources[chain.ChainMeta.Id][i].Container.ID)
				if height <= 75 {
					fmt.Printf("current block height is %v, waiting to hit blocks\n", height)
				}
				return height > 75
			},
			2*time.Minute,
			5*time.Second,
		)
		s.T().Logf("upgrade successful on %s validator container: %s", chain.ChainMeta.Id, s.valResources[chain.ChainMeta.Id][i].Container.ID)
	}

}
