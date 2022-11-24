package upgrade

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

// BuildImage build docker image by provided path with <name>:<version> as name target
func (m *Manager) BuildImage(name, version, dockerFile, contextDir string, args map[string]string) error {
	buildArgs := []docker.BuildArg{}
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
		// name with tag, e.g. evmos:v9.0.0
		Name:         fmt.Sprintf("%s:%s", name, version),
		OutputStream: io.Discard,
		ErrorStream:  os.Stdout,
		ContextDir:   contextDir,
	}
	return m.Client().BuildImage(opts)
}

// RunNode create docker container from provided node instance and run it.
// Retry to get node JRPC server for 60 seconds to make sure node started properly.
// On node start error returns container logs.
func (m *Manager) RunNode(node *Node) error {
	var resource *dockertest.Resource
	var err error

	if node.withRunOptions {
		resource, err = m.pool.RunWithOptions(node.RunOptions)
	} else {
		resource, err = m.pool.Run(node.repository, node.version, []string{})
	}

	if err != nil {
		return err
	}

	// trying to get JSON-RPC server, to make sure node started properly
	// the last returned error before deadline exceeded will be returned from .Retry()
	err = m.pool.Retry(
		func() error {
			// recreating container instance because resource.Container.State
			// does not updateing properly by default
			c, err := m.Client().InspectContainer(resource.Container.ID)
			if err != nil {
				return err
			}
			// if node failed to start, i.e. ExitCode != 0, return container logs
			if c.State.ExitCode != 0 {
				var outBuf, errBuf bytes.Buffer
				// no error check becuse success logs retieving returns error anyways
				_ = m.Client().Logs(docker.LogsOptions{
					Container:    resource.Container.ID,
					OutputStream: &outBuf,
					ErrorStream:  &errBuf,
					Stdout:       true,
					Stderr:       true,
				})
				return fmt.Errorf(
					"can't start evmos node, container exit code: %d\n\n[error stream]:\n\n%s\n\n[output stream]:\n\n%s",
					c.State.ExitCode,
					errBuf.String(),
					outBuf.String(),
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
		return err
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
