package upgrade

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// Manager defines a docker pool instance, used to run and interact with evmos
// node containers. It enables run, query, execute cli commands and purge.
type Manager struct {
	pool    *dockertest.Pool
	network *dockertest.Network

	CurrentNode     *dockertest.Resource
	proposalCounter uint
}

// NewManager creates new docker pool and network
// returns Manager instance
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

// RunNode create docker container from provided node instance and run it
func (m *Manager) RunNode(node *Node) error {
	var resource *dockertest.Resource
	var err error

	if node.withRunOptions {
		resource, err = m.pool.RunWithOptions(node.runOptions)
	} else {
		resource, err = m.pool.Run(node.repository, node.version, []string{})
	}

	if err != nil {
		return err
	}
	// trying to get JSON-RPC server, to make sure node started properly
	err = m.pool.Retry(
		func() error {
			_, err := http.Get(fmt.Sprintf("http://localhost:%s", resource.GetPort("8545/tcp")))
			return err
		},
	)
	if err != nil {
		return fmt.Errorf("could not connect to JSON-RPC server: %s", err)
	}

	m.CurrentNode = resource
	return nil
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

// Makes system call to current node container environment with evmosd cli command to get current block height
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

// ContainerID returns current running container ID
func (m *Manager) ContainerID() string {
	return m.CurrentNode.Container.ID
}

// Docker client
func (m *Manager) Client() *docker.Client {
	return m.pool.Client
}

func (m *Manager) RemoveNetwork() error {
	return m.pool.RemoveNetwork(m.network)
}

func (m *Manager) KillCurrentNode() error {
	return m.pool.Client.StopContainer(m.ContainerID(), 5)
}
