// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package upgrade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ethereum/go-ethereum/common"
	testnetwork "github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// safetyDelta is the number of blocks that are added to the upgrade height to make sure
// the proposal has concluded when reaching the upgrade height.
const safetyDelta = 2

// Manager defines a docker pool instance, used to build, run, interact with and stop docker containers
// running Evmos nodes.
type Manager struct {
	pool    *dockertest.Pool
	network *dockertest.Network

	// CurrentNode stores the currently running docker container
	CurrentNode *dockertest.Resource

	// CurrentVersion stores the current version of the running node
	CurrentVersion string

	// HeightBeforeStop stores the last block height that was reached before the last running node container
	// was stopped
	HeightBeforeStop int

	// proposalCounter keeps track of the number of proposals that have been submitted
	proposalCounter uint

	// ProtoCodec is the codec used to marshal/unmarshal protobuf messages
	ProtoCodec *codec.ProtoCodec

	// UpgradeHeight stores the upgrade height for the latest upgrade proposal that was submitted
	UpgradeHeight uint
}

// NewManager creates new docker pool and network and returns a populated Manager instance
func NewManager(networkName string) (*Manager, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("docker pool creation error: %w", err)
	}

	network, err := pool.CreateNetwork(networkName)
	if err != nil {
		return nil, fmt.Errorf("docker network creation error: %w", err)
	}

	nw := testnetwork.New()
	encodingConfig := nw.GetEncodingConfig()
	protoCodec, ok := encodingConfig.Codec.(*codec.ProtoCodec)
	if !ok {
		return nil, fmt.Errorf("failed to get proto codec")
	}

	return &Manager{
		pool:       pool,
		network:    network,
		ProtoCodec: protoCodec,
	}, nil
}

// BuildImage builds a docker image to run in the provided context directory
// with <name>:<version> as the image target
func (m *Manager) BuildImage(name, version, dockerFile, contextDir string, args map[string]string) error {
	buildArgs := make([]docker.BuildArg, 0, len(args))
	for k, v := range args {
		bArg := docker.BuildArg{
			Name:  k,
			Value: v,
		}
		buildArgs = append(buildArgs, bArg)
	}
	opts := docker.BuildImageOptions{
		// local Dockerfile path
		Dockerfile: dockerFile,
		BuildArgs:  buildArgs,
		// rebuild the image every time in case there were changes
		// and the image is cached
		NoCache: true,
		// name with tag, e.g. evmos:v9.0.0
		Name:         fmt.Sprintf("%s:%s", name, version),
		OutputStream: io.Discard,
		ErrorStream:  os.Stdout,
		ContextDir:   contextDir,
	}
	return m.Client().BuildImage(opts)
}

// RunNode creates a docker container from the provided node instance and runs it.
// To make sure the node started properly, get requests are sent to the JSON-RPC server repeatedly
// with a timeout of 60 seconds.
// In case the node fails to start, the container logs are returned along with the error.
func (m *Manager) RunNode(node *Node) error {
	var resource *dockertest.Resource
	var err error

	if node.withRunOptions {
		resource, err = m.pool.RunWithOptions(node.RunOptions)
	} else {
		resource, err = m.pool.Run(node.repository, node.version, []string{})
	}

	if err != nil {
		if resource == nil || resource.Container == nil {
			return fmt.Errorf(
				"can't run container\n[error]: %s",
				err.Error(),
			)
		}
		stdOut, stdErr, _ := m.GetLogs(resource.Container.ID)
		return fmt.Errorf(
			"can't run container\n\n[error stream]:\n\n%s\n\n[output stream]:\n\n%s",
			stdErr,
			stdOut,
		)
	}

	// trying to get JSON-RPC server, to make sure node started properly
	// the last returned error before deadline exceeded will be returned from .Retry()
	err = m.pool.Retry(
		func() error {
			// recreating container instance because resource.Container.State
			// does not update properly by default
			c, err := m.Client().InspectContainer(resource.Container.ID)
			if err != nil {
				return fmt.Errorf("can't inspect container: %s", err.Error())
			}
			// if node failed to start, i.e. ExitCode != 0, return container logs
			if c.State.ExitCode != 0 {
				stdOut, stdErr, _ := m.GetLogs(resource.Container.ID)
				return fmt.Errorf(
					"can't start evmos node, container exit code: %d\n\n[error stream]:\n\n%s\n\n[output stream]:\n\n%s",
					c.State.ExitCode,
					stdErr,
					stdOut,
				)
			}
			// get host:port for current container in local network
			addr := resource.GetHostPort(jrpcPort + "/tcp")
			r, err := http.Get("http://" + addr)
			if err != nil {
				return fmt.Errorf("can't get node json-rpc server: %s", err)
			}
			defer r.Body.Close()
			return nil
		},
	)
	if err != nil {
		stdOut, stdErr, _ := m.GetLogs(resource.Container.ID)
		return fmt.Errorf(
			"can't start node: %s\n\n[error stream]:\n\n%s\n\n[output stream]:\n\n%s",
			err.Error(),
			stdErr,
			stdOut,
		)
	}
	m.CurrentNode = resource
	return nil
}

// GetLogs returns the logs of the container with the provided containerID
func (m *Manager) GetLogs(containerID string) (stdOut, stdErr string, err error) {
	var outBuf, errBuf bytes.Buffer
	opts := docker.LogsOptions{
		Container:    containerID,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
		Stdout:       true,
		Stderr:       true,
	}
	err = m.Client().Logs(opts)
	if err != nil {
		return "", "", fmt.Errorf("can't get logs: %s", err)
	}
	return outBuf.String(), errBuf.String(), nil
}

// WaitNBlocks is a helper function to wait the specified number of blocks
func (m *Manager) WaitNBlocks(ctx context.Context, n int) error {
	currentHeight, err := m.GetNodeHeight(ctx)
	if err != nil {
		return err
	}
	_, err = m.WaitForHeight(ctx, currentHeight+n)
	return err
}

// WaitForHeight queries the Evmos node every second until the node will reach the specified height.
// After 5 minutes this function times out and returns an error
func (m *Manager) WaitForHeight(ctx context.Context, height int) (string, error) {
	var currentHeight int
	var err error
	ticker := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-ticker.C:
			stdOut, stdErr, errLogs := m.GetLogs(m.ContainerID())
			if errLogs != nil {
				return "", fmt.Errorf("error while getting logs: %s", errLogs.Error())
			}
			return "", fmt.Errorf(
				"can't reach height %d, due to: %v\nerror logs: %s\nout logs: %s",
				height, err, stdOut, stdErr,
			)
		default:
			currentHeight, err = m.GetNodeHeight(ctx)
			if currentHeight >= height {
				if err != nil {
					return err.Error(), nil
				}
				return "", nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// GetNodeHeight calls the Evmos CLI in the current node container to get the current block height
func (m *Manager) GetNodeHeight(ctx context.Context) (int, error) {
	cmd := []string{"evmosd", "q", "block"}
	splitIdx := 0 // split index for the lines in the output - in newer versions the output is in the second line
	useV50 := false

	// if the version is higher than v20.0.0, we need to use the json output flag
	if !EvmosVersions([]string{m.CurrentVersion, "v20.0.0"}).Less(0, 1) {
		cmd = append(cmd, "--output=json")
		splitIdx = 1
		useV50 = true
	}

	exec, err := m.CreateExec(cmd, m.ContainerID())
	if err != nil {
		return 0, fmt.Errorf("create exec error: %w", err)
	}

	outBuff, errBuff, err := m.RunExec(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("run exec error: %w", err)
	}

	if errBuff.String() != "" {
		return 0, fmt.Errorf("evmos query error: %s", errBuff.String())
	}

	// NOTE: we're splitting the output because it has the first line saying "falling back to latest height"
	splittedOutBuff := strings.Split(outBuff.String(), "\n")
	if len(splittedOutBuff) < splitIdx+1 {
		return 0, fmt.Errorf("unexpected output format for node height; got: %s", outBuff.String())
	}

	outStr := splittedOutBuff[splitIdx]
	var h int
	// parse current height number from block info
	if outStr != "<nil>" && outStr != "" {
		if useV50 {
			h, err = UnwrapBlockHeightPostV50(outStr)
		} else {
			h, err = UnwrapBlockHeightPreV50(outStr)
		}

		// check if the conversion was possible
		if err == nil {
			// if conversion was possible but the errBuff is not empty, return the height along with an error
			// this is necessary e.g. when the "duplicate proto" errors occur in the logs but the node is still
			// producing blocks
			if errBuff.String() != "" {
				return h, fmt.Errorf("%s", errBuff.String())
			}
			return h, nil
		}
	}

	return h, nil
}

type BlockHeaderPreV50 struct {
	Block struct {
		Header struct {
			Height string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}

type BlockHeaderPostV50 struct {
	Header struct {
		Height string `json:"height"`
	} `json:"header"`
}

// UnwrapBlockHeightPreV50 unwraps the block height from the output of the evmosd query block command
func UnwrapBlockHeightPreV50(input string) (int, error) {
	var blockHeader BlockHeaderPreV50
	err := json.Unmarshal([]byte(input), &blockHeader)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return strconv.Atoi(blockHeader.Block.Header.Height)
}

// UnwrapBlockHeightPostV50 unwraps the block height from the output of the evmosd query block command
func UnwrapBlockHeightPostV50(input string) (int, error) {
	var blockHeader BlockHeaderPostV50
	err := json.Unmarshal([]byte(input), &blockHeader)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return strconv.Atoi(blockHeader.Header.Height)
}

// GetNodeVersion calls the Evmos CLI in the current node container to get the
// current node version
func (m *Manager) GetNodeVersion(ctx context.Context) (string, error) {
	exec, err := m.CreateExec([]string{"evmosd", "version"}, m.ContainerID())
	if err != nil {
		return "", fmt.Errorf("create exec error: %w", err)
	}
	outBuff, errBuff, err := m.RunExec(ctx, exec)
	if err != nil {
		return "", fmt.Errorf("run exec error: %w", err)
	}
	if errBuff.String() != "" {
		return "", fmt.Errorf("evmos version error: %s", errBuff.String())
	}
	return outBuff.String(), nil
}

// GetUpgradeHeight calculates the height for the scheduled software upgrade.
//
// It checks the timeout commit that is configured on the node and checks the voting duration.
// From this information, the height at which the upgrade will be scheduled is calculated.
func (m *Manager) GetUpgradeHeight(ctx context.Context, chainID string) (uint, error) {
	currentHeight, err := m.GetNodeHeight(ctx)
	if err != nil {
		return 0, err
	}

	timeoutCommit, err := m.getTimeoutCommit(ctx)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to get timeout commit")
	}

	votingPeriod, err := m.getVotingPeriod(ctx, chainID)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to get voting period")
	}

	heightDelta := new(big.Int).Quo(votingPeriod, timeoutCommit)
	upgradeHeight := uint(currentHeight) + uint(heightDelta.Int64()) + safetyDelta // #nosec G115

	// return the height at which the upgrade will be scheduled
	return upgradeHeight, nil
}

// getTimeoutCommit returns the timeout commit duration for the current node
func (m *Manager) getTimeoutCommit(ctx context.Context) (*big.Int, error) {
	exec, err := m.CreateExec([]string{"grep", `\s*timeout_commit =`, "/root/.evmosd/config/config.toml"}, m.ContainerID())
	if err != nil {
		return common.Big0, fmt.Errorf("create exec error: %w", err)
	}

	outBuff, errBuff, err := m.RunExec(ctx, exec)
	if err != nil {
		return common.Big0, fmt.Errorf("failed to execute command: %s", err.Error())
	}

	if errBuff.String() != "" {
		return common.Big0, fmt.Errorf("evmos version error: %s", errBuff.String())
	}

	re := regexp.MustCompile(`timeout_commit = "(?P<t>\d+)s"`)
	match := re.FindStringSubmatch(outBuff.String())
	if len(match) != 2 {
		return common.Big0, fmt.Errorf("failed to parse timeout_commit: %s", outBuff.String())
	}

	tc, err := strconv.Atoi(match[1])
	if err != nil {
		return common.Big0, err
	}

	return big.NewInt(int64(tc)), nil
}

// getVotingPeriod returns the voting period for the current node
func (m *Manager) getVotingPeriod(ctx context.Context, chainID string) (*big.Int, error) {
	exec, err := m.CreateModuleQueryExec(QueryArgs{
		Module:     "gov",
		SubCommand: "params",
		ChainID:    chainID,
	})
	if err != nil {
		return common.Big0, fmt.Errorf("create exec error: %w", err)
	}

	outBuff, errBuff, err := m.RunExec(ctx, exec)
	if err != nil {
		return common.Big0, fmt.Errorf("failed to execute command: %s", err.Error())
	}

	if errBuff.String() != "" {
		return common.Big0, fmt.Errorf("evmos version error: %s", errBuff.String())
	}

	re := regexp.MustCompile(`"voting_period":"(?P<votingPeriod>\d+)s"`)
	match := re.FindStringSubmatch(outBuff.String())
	if len(match) != 2 {
		return common.Big0, fmt.Errorf("failed to parse voting_period: %s", outBuff.String())
	}

	vp, err := strconv.Atoi(match[1])
	if err != nil {
		return common.Big0, err
	}

	return big.NewInt(int64(vp)), nil
}

// ContainerID returns the docker container ID of the currently running Node
func (m *Manager) ContainerID() string {
	if m.CurrentNode == nil || m.CurrentNode.Container == nil {
		return ""
	}
	return m.CurrentNode.Container.ID
}

// The Client method returns the Docker client used by the Manager's pool
func (m *Manager) Client() *docker.Client {
	return m.pool.Client
}

// RemoveNetwork removes the Manager's used network from the pool
func (m *Manager) RemoveNetwork() error {
	return m.pool.RemoveNetwork(m.network)
}

// KillCurrentNode stops the execution of the currently used docker container
func (m *Manager) KillCurrentNode() error {
	heightBeforeStop, err := m.GetNodeHeight(context.Background())
	if err != nil && heightBeforeStop == 0 {
		return err
	}
	m.HeightBeforeStop = heightBeforeStop
	return m.pool.Client.StopContainer(m.ContainerID(), 5)
}
