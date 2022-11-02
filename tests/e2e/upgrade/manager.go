package upgrade

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// Manager defines docker pool instance, used to run and interact with evmos
// node containers: run, query, execute cli commands and purge
type Manager struct {
	pool    *dockertest.Pool
	network *dockertest.Network

	CurrentNode     *dockertest.Resource
	proposalCounter uint
}

func NewManager(networkName string) (*Manager, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("docker pool creation error: %w", err)
	}

	network, err := pool.CreateNetwork(networkName)
	if err != nil {
		return nil, fmt.Errorf("docker network creation error: %w", err)
	}
	return &Manager{
		pool:    pool,
		network: network,
	}, nil
}

func (m *Manager) RunNode(node *Node) error {
	// sleep to let container start to prevent querying panics
	defer time.Sleep(5 * time.Second)
	if node.withRunOptions {
		resource, err := m.pool.RunWithOptions(node.runOptions)
		if err != nil {
			return err
		}
		m.CurrentNode = resource
		return nil
	}
	resource, err := m.pool.Run(node.repository, node.version, []string{})
	if err != nil {
		return err
	}
	m.CurrentNode = resource
	return nil
}

func (m *Manager) RemoveNetwork() error {
	return m.pool.RemoveNetwork(m.network)
}

func (m *Manager) KillCurrentNode() error {
	return m.pool.Client.StopContainer(m.ContainerID(), 5)
}

func (m *Manager) ContainerID() string {
	return m.CurrentNode.Container.ID
}

// Docker client
func (m *Manager) Client() *docker.Client {
	return m.pool.Client
}

// WaitForHeight query evmos node every second until node will reach specified height
// for 5 minutes, after time exceed returns error
func (m *Manager) WaitForHeight(ctx context.Context, height int) error {
	var currentHeight int
	var err error
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			return fmt.Errorf("can't reach height %d, due: %w", height, err)
		default:
			currentHeight, err = m.nodeHeight(ctx)
			if currentHeight >= height {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (m *Manager) nodeHeight(ctx context.Context) (int, error) {
	exec, err := m.CreateExec([]string{"evmosd", "q", "block"}, m.ContainerID())
	if err != nil {
		return 0, fmt.Errorf("create exec error: %w", err)
	}
	outBuff, errBuff, err := m.RunExec(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("run exec error: %w", err)
	}
	outStr := outBuff.String()
	var h int
	// parse current height number from block info
	if outStr != "<nil>" && outStr != "" {
		index := strings.Index(outBuff.String(), "\"height\":")
		qq := outStr[index+10 : index+12]
		h, _ = strconv.Atoi(qq)
	}
	if errBuff.String() != "" {
		return 0, fmt.Errorf("evmos query error: %s", errBuff.String())
	}
	return h, nil
}

// ExportState execute 'docker cp' command to copy container .evmosd dir
// to specified local dir
// https://docs.docker.com/engine/reference/commandline/cp/
func (m *Manager) ExportState(targetDir string) error {
	/* #nosec G204 */
	cmd := exec.Command(
		"docker",
		"cp",
		fmt.Sprintf("%s:/root/.evmosd", m.ContainerID()),
		targetDir,
	)
	return cmd.Run()
}
