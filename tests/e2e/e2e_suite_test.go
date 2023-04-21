package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/evmos/evmos/v13/tests/e2e/upgrade"
	"github.com/evmos/evmos/v13/utils"
)

const (
	// defaultManagerNetwork defines the network used by the upgrade manager
	defaultManagerNetwork = "evmos-local"

	// blocksAfterUpgrade defines how many blocks must be produced after an upgrade is
	// considered successful
	blocksAfterUpgrade = 5

	// relatedBuildPath defines the path where the build data is stored
	relatedBuildPath = "../../build/"

	// upgradeHeightDelta defines the number of blocks after the proposal and the scheduled upgrade
	upgradeHeightDelta = 10

	// upgradePath defines the relative path from this folder to the upgrade folder
	upgradePath = "../../app/upgrades"

	// registryDockerFile builds the image using the docker image registry
	registryDockerFile = "./upgrade/Dockerfile.init"

	// repoDockerFile builds the image from the repository (used when the images are not pushed to the registry, e.g. main)
	repoDockerFile = "./Dockerfile.repo"
)

type IntegrationTestSuite struct {
	suite.Suite

	upgradeManager *upgrade.Manager
	upgradeParams  upgrade.Params
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var err error

	s.upgradeParams, err = upgrade.LoadUpgradeParams(upgradePath)
	s.Require().NoError(err, "can't load upgrade params")

	s.upgradeManager, err = upgrade.NewManager(defaultManagerNetwork)
	s.Require().NoError(err, "upgrade manager creation error")
	if _, err := os.Stat(relatedBuildPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(relatedBuildPath, os.ModePerm)
		s.Require().NoError(err, "can't create build tmp dir")
	}
}

// runInitialNode builds a docker image capable of running an Evmos node with the given version.
// After a successful build, it runs the container and checks if the node can produce blocks.
func (s *IntegrationTestSuite) runInitialNode(version upgrade.VersionConfig) {
	err := s.upgradeManager.BuildImage(
		version.ImageName,
		version.ImageTag,
		registryDockerFile,
		".",
		map[string]string{"INITIAL_VERSION": version.ImageTag},
	)
	s.Require().NoError(err, "can't build container with Evmos version: %s", version.ImageTag)

	node := upgrade.NewNode(version.ImageName, version.ImageTag)
	node.SetEnvVars([]string{fmt.Sprintf("CHAIN_ID=%s", s.upgradeParams.ChainID)})

	err = s.upgradeManager.RunNode(node)
	s.Require().NoError(err, "can't run node with Evmos version: %s", version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// wait until node starts and produce some blocks
	_, err = s.upgradeManager.WaitForHeight(ctx, s.upgradeManager.HeightBeforeStop+5)
	s.Require().NoError(err)

	s.T().Logf("successfully started node with version: [%s]", version.ImageTag)
}

// runNodeWithCurrentChanges builds a docker image using the current branch of the Evmos repository.
// Before running the node, runs a script to modify some configurations for the tests
// (e.g.: gov proposal voting period, setup accounts, balances, etc..)
// After a successful build, runs the container.
func (s *IntegrationTestSuite) runNodeWithCurrentChanges() {
	const (
		name    = "e2e-test/evmos"
		version = "latest"
	)
	// get the current branch name
	// to run the tests against the last changes
	branch, err := getCurrentBranch()
	s.Require().NoError(err)

	err = s.upgradeManager.BuildImage(
		name,
		version,
		repoDockerFile,
		".",
		map[string]string{"BRANCH_NAME": branch},
	)
	s.Require().NoError(err, "can't build container for e2e test")

	node := upgrade.NewNode(name, version)
	node.SetEnvVars([]string{fmt.Sprintf("CHAIN_ID=%s", s.upgradeParams.ChainID)})

	err = s.upgradeManager.RunNode(node)
	s.Require().NoError(err, "can't run node Evmos using branch %s", branch)
}

// proposeUpgrade submits an upgrade proposal to the chain that schedules an upgrade to
// the given target version.
func (s *IntegrationTestSuite) proposeUpgrade(name, target string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// calculate upgrade height for the proposal
	nodeHeight, err := s.upgradeManager.GetNodeHeight(ctx)
	s.Require().NoError(err, "can't get block height from running node")
	s.upgradeManager.UpgradeHeight = uint(nodeHeight + upgradeHeightDelta)

	// if Evmos is lower than v10.x.x no need to use the legacy proposal
	currentVersion, err := s.upgradeManager.GetNodeVersion(ctx)
	s.Require().NoError(err, "can't get current Evmos version")
	isLegacyProposal := upgrade.CheckLegacyProposal(currentVersion)

	// create the proposal
	exec, err := s.upgradeManager.CreateSubmitProposalExec(
		name,
		s.upgradeParams.ChainID,
		s.upgradeManager.UpgradeHeight,
		isLegacyProposal,
		"--fees=10000000000000000aevmos",
		"--gas=500000",
	)
	s.Require().NoErrorf(
		err,
		"can't create the proposal to upgrade Evmos to %s at height %d with name %s",
		target, s.upgradeManager.UpgradeHeight, name,
	)

	outBuf, errBuf, err := s.upgradeManager.RunExec(ctx, exec)
	s.Require().NoErrorf(
		err,
		"failed to submit proposal to upgrade Evmos to %s at height %d\nstdout: %s,\nstderr: %s",
		target, s.upgradeManager.UpgradeHeight, outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned non code 0:\nstdout: %s\nstderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf(
		"successfully submitted upgrade proposal: height: %d, name: %s",
		s.upgradeManager.UpgradeHeight,
		name,
	)
}

// voteForProposal votes for the upgrade proposal with the given id.
func (s *IntegrationTestSuite) voteForProposal(id int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exec, err := s.upgradeManager.CreateVoteProposalExec(s.upgradeParams.ChainID, id, "--fees=10000000000000000aevmos", "--gas=500000")
	s.Require().NoError(err, "can't create vote for proposal exec")
	outBuf, errBuf, err := s.upgradeManager.RunExec(ctx, exec)
	s.Require().NoErrorf(
		err,
		"failed to vote for proposal tx;\nstdout: %s,\nstderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Truef(
		strings.Contains(outBuf.String(), "code: 0"),
		"tx returned non code 0:\nstdout: %s\nstderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("successfully voted for upgrade proposal")
}

// upgrade upgrades the node to the given version using the given repo. The repository
// can either be a local path or a remote repository.
func (s *IntegrationTestSuite) upgrade(targetRepo, targetVersion string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.T().Log("wait for node to reach upgrade height...")
	// wait for proposed upgrade height
	_, err := s.upgradeManager.WaitForHeight(ctx, int(s.upgradeManager.UpgradeHeight))
	s.Require().NoError(err, "can't reach upgrade height")
	buildDir := strings.Split(s.upgradeParams.MountPath, ":")[0]

	s.T().Log("exporting state to local...")
	// export node .evmosd to local build/
	err = s.upgradeManager.ExportState(buildDir)
	s.Require().NoError(err, "can't export node container state to local")

	s.T().Log("killing node before upgrade...")
	err = s.upgradeManager.KillCurrentNode()
	s.Require().NoError(err, "can't kill current node")

	s.T().Logf("starting upgraded node with version: [%s]", targetVersion)

	node := upgrade.NewNode(targetRepo, targetVersion)
	node.Mount(s.upgradeParams.MountPath)
	node.SetCmd([]string{"evmosd", "start", fmt.Sprintf("--chain-id=%s", s.upgradeParams.ChainID)})
	err = s.upgradeManager.RunNode(node)
	s.Require().NoError(err, "can't mount and run upgraded node container")

	s.T().Logf("node started! waiting for node to produce %d blocks", blocksAfterUpgrade)

	s.T().Logf("executing all module queries")
	s.executeQueries()

	// make sure node produce blocks after upgrade
	s.T().Logf("height to wait for is %d", int(s.upgradeManager.UpgradeHeight)+blocksAfterUpgrade)
	// make sure node produces blocks after upgrade
	errLogs, err := s.upgradeManager.WaitForHeight(ctx, int(s.upgradeManager.UpgradeHeight)+blocksAfterUpgrade)
	if err == nil && errLogs != "" {
		s.T().Logf(
			"even though the node is producing blocks, there are error messages contained in the logs:\n%s",
			errLogs,
		)
	}
	s.Require().NoError(err, "node does not produce blocks after upgrade")

	if targetVersion != upgrade.LocalVersionTag {
		s.T().Log("checking node version...")
		version, err := s.upgradeManager.GetNodeVersion(ctx)
		s.Require().NoError(err, "can't get node version")

		version = strings.TrimSpace(version)
		targetVersion = strings.TrimPrefix(targetVersion, "v")
		s.Require().Equal(targetVersion, version,
			"unexpected node version after upgrade:\nexpected: %s\nactual: %s",
			targetVersion, version,
		)
		s.T().Logf("node version is correct: %s", version)
	}
}

// executeQueries executes all the module queries
func (s *IntegrationTestSuite) executeQueries() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chainID := utils.TestnetChainID + "-1"
	testCases := []struct {
		name       string
		moduleName string
		subCommand string
	}{
		{"inflation: params", "inflation", "params"},
		{"inflation: circulating-supply", "inflation", "circulating-supply"},
		{"inflation: inflation-rate", "inflation", "inflation-rate"},
		{"inflation: period", "inflation", "period"},
		{"inflation: skipped-epochs", "inflation", "skipped-epochs"},
		{"inflation: epoch-mint-provision", "inflation", "epoch-mint-provision"},
		{"erc20: params", "erc20", "params"},
		{"erc20: token-pairs", "erc20", "token-pairs"},
		{"evm: params", "evm", "params"},
		{"feemarket: params", "feemarket", "params"},
		{"feemarket: base-fee", "feemarket", "base-fee"},
		{"feemarket: block-gas", "feemarket", "block-gas"},
		{"feemarket: block-gas", "feemarket", "block-gas"},
		{"revenue: params", "revenue", "params"},
		{"revenue: contracts", "revenue", "contracts"},
		{"incentives: params", "incentives", "params"},
		{"incentives: allocation-meters", "incentives", "allocation-meters"},
		{"incentives: incentives", "incentives", "incentives"},
	}

	for _, tc := range testCases {
		s.T().Logf("executing %s", tc.name)
		exec, err := s.upgradeManager.CreateModuleQueryExec(tc.moduleName, tc.subCommand, chainID)
		s.Require().NoError(err)

		_, errBuf, err := s.upgradeManager.RunExec(ctx, exec)
		s.Require().NoError(err)
		s.Require().Empty(errBuf.String())
	}
	s.T().Logf("executed all queries successfully")
}

// TearDownSuite kills the running container, removes the network and mount path
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.upgradeParams.SkipCleanup {
		s.T().Logf("skipping cleanup... container %s will be left running", s.upgradeManager.ContainerID())
		return
	}

	s.T().Log("tearing down e2e integration test suite...")
	s.T().Log("killing node...")
	err := s.upgradeManager.KillCurrentNode()
	s.Require().NoError(err, "can't kill current node")

	s.T().Log("removing network...")
	s.Require().NoError(s.upgradeManager.RemoveNetwork(), "can't remove network")

	s.T().Log("removing mount path...")
	s.Require().NoError(os.RemoveAll(strings.Split(s.upgradeParams.MountPath, ":")[0]), "can't remove mount path")
}
