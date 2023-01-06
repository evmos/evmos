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

import "github.com/ory/dockertest/v3"

const jrpcPort = "8545"

// Node defines evmos node params for running container
// of specific version with custom docker run arguments
type Node struct {
	repository string
	version    string

	RunOptions     *dockertest.RunOptions
	withRunOptions bool
}

// NewNode creates new instance of the node and setups default dockertest RunOptions
func NewNode(repository, version string) *Node {
	return &Node{
		repository: repository,
		version:    version,
		RunOptions: &dockertest.RunOptions{
			Repository: repository,
			Tag:        version,
			// exposing JSON-RPC port by default to ping node after start
			ExposedPorts: []string{jrpcPort},
		},
	}
}

// SetEnvVars allows to set addition container environment variables in format
// []string{ "VAR_NAME=valaue" }
func (n *Node) SetEnvVars(vars []string) {
	n.RunOptions.Env = vars
	n.UseRunOptions()
}

// Mount sets the container mount point, which is used as the value for 'docker run --volume'
//
// See https://docs.docker.com/engine/reference/builder/#volume
func (n *Node) Mount(mountPath string) {
	n.RunOptions.Mounts = []string{mountPath}
	n.UseRunOptions()
}

// SetCmd sets the container entry command and overrides the image CMD instruction
//
// See https://docs.docker.com/engine/reference/builder/#cmd
func (n *Node) SetCmd(cmd []string) {
	n.RunOptions.Cmd = cmd
	n.UseRunOptions()
}

// UseRunOptions sets a flag to allow the node Manager to run the container with additional run options
func (n *Node) UseRunOptions() {
	n.withRunOptions = true
}
