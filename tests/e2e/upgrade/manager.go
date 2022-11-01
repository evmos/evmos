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

type Manager struct {
	pool    *dockertest.Pool
	network *dockertest.Network

	CurrentNode     *dockertest.Resource
	proposalCounter uint
}

func NewManager() (*Manager, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("docker pool creation error: %w", err)
	}

	network, err := pool.CreateNetwork("evmos-local")
	if err != nil {
		return nil, fmt.Errorf("docker network creation error: %w", err)
	}
	return &Manager{
		pool:    pool,
		network: network,
	}, nil
}

func (m *Manager) RunNode(node *Node) error {
	if node.withRunOptions {
		resource, err := m.pool.RunWithOptions(&node.runOptions)
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
	// sleep to let container start to prevent querying panics
	time.Sleep(5 * time.Second)
	return nil
}

func (m *Manager) RemoveNetwork() error {
	return m.pool.RemoveNetwork(m.network)
}

func (m *Manager) KillCurrentNode() error {
	time.Sleep(5 * time.Second)
	return m.pool.Purge(m.CurrentNode)
}

func (m *Manager) ContainerID() string {
	return m.CurrentNode.Container.ID
}

func (m *Manager) Client() *docker.Client {
	return m.pool.Client
}

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
	index := strings.Index(outBuff.String(), "\"height\":")
	qq := outBuff.String()[index+10 : index+12]
	h, _ := strconv.Atoi(qq)
	if errBuff.String() != "" {
		return 0, fmt.Errorf("evmos query error: %s", errBuff.String())
	}
	return h, nil
}

func (m *Manager) ExportState(targetDir string) error {
	cmd := exec.Command(
		"docker",
		"cp",
		fmt.Sprintf("%s:/root/.evmosd", m.ContainerID()),
		targetDir,
	)
	return cmd.Run()
}
