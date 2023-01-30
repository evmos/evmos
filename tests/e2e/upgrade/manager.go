// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

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

// Manager defines a docker pool instance, used to build, run, interact with and stop docker containers
// running Evmos nodes.
type Manager struct {
	pool    *dockertest.Pool
	network *dockertest.Network

	// CurrentNode stores the currently running docker container
	CurrentNode *dockertest.Resource

	// HeightBeforeStop stores the last block height that was reached before the last running node container
	// was stopped
	HeightBeforeStop int

	// proposalCounter keeps track of the number of proposals that have been submitted
	proposalCounter uint

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
	return &Manager{
		pool:    pool,
		network: network,
	}, nil
}

// BuildImage builds a docker image to run in the provided context directory
// with <name>:<version> as the name target
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

// WaitForHeight queries the Evmos node every second until the node will reach the specified height.
// After 5 minutes this function times out and returns an error
func (m *Manager) WaitForHeight(ctx context.Context, height int) error {
	var currentHeight int
	var err error
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			stdOut, stdErr, errLogs := m.GetLogs(m.ContainerID())
			if errLogs != nil {
				return fmt.Errorf("error while getting logs: %s", errLogs.Error())
			}
			return fmt.Errorf(
				"can't reach height %d, due to: %s\nerror logs: %s\nout logs: %s",
				height, err.Error(), stdOut, stdErr,
			)
		default:
			currentHeight, err = m.GetNodeHeight(ctx)
			if currentHeight >= height {
				return nil
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// GetNodeHeight calls the Evmos CLI in the current node container to get the current block height
func (m *Manager) GetNodeHeight(ctx context.Context) (int, error) {
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

// ContainerID returns the docker container ID of the currently running Node
func (m *Manager) ContainerID() string {
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
	if err != nil {
		return err
	}
	m.HeightBeforeStop = heightBeforeStop
	return m.pool.Client.StopContainer(m.ContainerID(), 5)
}
