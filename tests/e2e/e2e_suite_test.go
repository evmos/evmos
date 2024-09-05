package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/evmos/evmos/v19/tests/e2e/upgrade"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/utils"
	"github.com/stretchr/testify/suite"
)

const (
	// defaultManagerNetwork defines the network used by the upgrade manager
	defaultManagerNetwork = "evmos-local"

	// blocksAfterUpgrade defines how many blocks must be produced after an upgrade is
	// considered successful
	blocksAfterUpgrade = 5

	// relatedBuildPath defines the path where the build data is stored
	relatedBuildPath = "../../build/"

	// upgradePath defines the relative path from this folder to the upgrade folder
	upgradePath = "../../app/upgrades"

	// registryDockerFile builds the image using the docker image registry
	registryDockerFile = "./upgrade/Dockerfile.init"
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

// proposeUpgrade submits an upgrade proposal to the chain that schedules an upgrade to
// the given target version.
func (s *IntegrationTestSuite) proposeUpgrade(name, target string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// calculate upgrade height for the proposal
	upgradeHeight, err := s.upgradeManager.GetUpgradeHeight(ctx, s.upgradeParams.ChainID)
	s.Require().NoError(err, "can't get upgrade height")
	s.upgradeManager.UpgradeHeight = upgradeHeight

	// if Evmos is lower than v10.x.x no need to use the legacy proposal
	currentVersion, err := s.upgradeManager.GetNodeVersion(ctx)
	s.Require().NoError(err, "can't get current Evmos version")
	proposalVersion := upgrade.CheckUpgradeProposalVersion(currentVersion)

	// create the proposal
	exec, err := s.upgradeManager.CreateSubmitProposalExec(
		name,
		s.upgradeParams.ChainID,
		s.upgradeManager.UpgradeHeight,
		proposalVersion,
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
func (s *IntegrationTestSuite) upgrade(targetVersion upgrade.VersionConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.T().Logf("wait for node to reach upgrade height %d...", s.upgradeManager.UpgradeHeight)
	// wait for proposed upgrade height
	_, err := s.upgradeManager.WaitForHeight(ctx, int(s.upgradeManager.UpgradeHeight)) //#nosec G115
	s.Require().NoError(err, "can't reach upgrade height")
	dirs := strings.Split(s.upgradeParams.MountPath, ":")
	buildDir := dirs[0]
	rootDir := dirs[1]

	// check that the proposal has passed before stopping the node
	s.checkProposalPassed(ctx)

	s.T().Log("exporting state to local...")
	// export node .evmosd to local build/
	err = s.upgradeManager.ExportState(buildDir)
	s.Require().NoError(err, "can't export node container state to local")

	s.T().Log("killing node before upgrade...")
	err = s.upgradeManager.KillCurrentNode()
	s.Require().NoError(err, "can't kill current node")

	s.T().Logf("starting upgraded node with version: [%s]", targetVersion)

	// NOTE: after the upgrade, the current version needs to be updated to make sure that the correct CLI commands
	// for the given version are used.
	//
	// this is e.g. relevant for retrieving the current node height from the block header
	if targetVersion.ImageTag == upgrade.LocalVersionTag {
		// NOTE: the upgrade name is the latest version from the app/upgrades folder to upgrade to
		s.upgradeManager.CurrentVersion = targetVersion.UpgradeName
	} else {
		s.upgradeManager.CurrentVersion = targetVersion.ImageTag
	}

	node := upgrade.NewNode(targetVersion.ImageName, targetVersion.ImageTag)
	node.Mount(s.upgradeParams.MountPath)
	node.SetCmd([]string{"evmosd", "start", fmt.Sprintf("--chain-id=%s", s.upgradeParams.ChainID), fmt.Sprintf("--home=%s.evmosd", rootDir)})
	err = s.upgradeManager.RunNode(node)
	s.Require().NoError(err, "can't mount and run upgraded node container")

	s.T().Logf("node started! waiting for node to produce %d blocks", blocksAfterUpgrade)

	s.T().Log("executing all module queries")
	s.executeQueries()

	s.T().Log("executing sample transactions")
	s.executeTransactions()

	// make sure node produce blocks after upgrade
	s.T().Logf("height to wait for is %d", int(s.upgradeManager.UpgradeHeight)+blocksAfterUpgrade) // #nosec G115
	// make sure node produces blocks after upgrade
	errLogs, err := s.upgradeManager.WaitForHeight(ctx, int(s.upgradeManager.UpgradeHeight)+blocksAfterUpgrade) // #nosec G115
	if err == nil && errLogs != "" {
		s.T().Logf(
			"even though the node is producing blocks, there are error messages contained in the logs:\n%s",
			errLogs,
		)
	}
	s.Require().NoError(err, "node does not produce blocks after upgrade")

	if targetVersion.ImageTag != upgrade.LocalVersionTag {
		s.T().Log("checking node version...")
		version, err := s.upgradeManager.GetNodeVersion(ctx)
		s.Require().NoError(err, "can't get node version")

		version = strings.TrimSpace(version)
		targetVersion.ImageTag = strings.TrimPrefix(targetVersion.ImageTag, "v")
		s.Require().Equal(targetVersion, version,
			"unexpected node version after upgrade:\nexpected: %s\nactual: %s",
			targetVersion, version,
		)
		s.T().Logf("node version is correct: %s", version)
	}
}

// checkProposalPassed queries the (most recent) upgrade proposal and checks that it has passed.
//
// NOTE: This was a problem in the past, where the upgrade height was reached before the proposal actually passed.
// This is a safety check to make sure this doesn't happen again, as this was not obvious from the log output.
func (s *IntegrationTestSuite) checkProposalPassed(ctx context.Context) {
	exec, err := s.upgradeManager.CreateModuleQueryExec(upgrade.QueryArgs{
		Module:     "gov",
		SubCommand: "proposals",
		ChainID:    s.upgradeParams.ChainID,
	})
	s.Require().NoError(err, "can't create query proposals exec")

	outBuf, errBuf, err := s.upgradeManager.RunExec(ctx, exec)
	s.Require().NoErrorf(
		err,
		"failed to query proposals;\nstdout: %s,\nstderr: %s", outBuf.String(), errBuf.String(),
	)

	nw := network.New()
	encodingConfig := nw.GetEncodingConfig()
	protoCodec, ok := encodingConfig.Codec.(*codec.ProtoCodec)
	s.Require().True(ok, "encoding config codec is not a proto codec")

	var proposalsRes govtypes.QueryProposalsResponse
	err = protoCodec.UnmarshalJSON(outBuf.Bytes(), &proposalsRes)
	s.Require().NoError(err, "can't unmarshal proposals response\n%s", outBuf.String())
	s.Require().GreaterOrEqual(len(proposalsRes.Proposals), 1, "no proposals found")

	// check that the most recent proposal has passed
	proposal := proposalsRes.Proposals[len(proposalsRes.Proposals)-1]
	s.Require().Equal(govtypes.ProposalStatus_PROPOSAL_STATUS_PASSED.String(), proposal.Status.String(), "expected proposal to have passed already")
}

// executeQueries executes all the module queries to check they are still working after the upgrade.
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
	}

	for _, tc := range testCases {
		s.T().Logf("executing %s", tc.name)
		exec, err := s.upgradeManager.CreateModuleQueryExec(upgrade.QueryArgs{
			Module:     tc.moduleName,
			SubCommand: tc.subCommand,
			ChainID:    chainID,
		})
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
