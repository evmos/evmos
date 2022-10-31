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

	CurrentNode *dockertest.Resource
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

func (m *Manager) RunNode(repository, version string) error {
	resource, err := m.pool.Run(repository, version, []string{})
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	m.CurrentNode = resource
	return nil
}

func (m *Manager) RunMountedNode(repository, version, mountPath string) error {
	resource, err := m.pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: repository,
			Tag:        version,
			User:       "root",
			Mounts: []string{
				mountPath,
			},
			Cmd: []string{
				"evmosd",
				"start",
			},
		},
	)
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	m.CurrentNode = resource
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
	exec, err := m.pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    m.ContainerID(),
		User:         "root",
		Cmd: []string{
			"evmosd",
			"q",
			"block",
		},
	})
	if err != nil {
		return 0, fmt.Errorf("create exec error: %w", err)
	}
	outBuff, errBuff, err := m.RunExec(ctx, exec.ID)
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

func (m *Manager) ExportState() error {
	cmd := exec.Command(
		"docker",
		"cp",
		fmt.Sprintf("%s:/root/.evmosd", m.ContainerID()),
		"./build",
	)
	return cmd.Run()
}
