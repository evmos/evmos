// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package upgrade

import "github.com/ory/dockertest/v3"

const jrpcPort = "8545"

// Node represents an Evmos node in the context of the upgrade tests. It contains
// fields to store the used repository, version as well as custom run options for
// dockertest.
type Node struct {
	repository string
	version    string

	RunOptions     *dockertest.RunOptions
	withRunOptions bool
}

// NewNode creates a new instance of the node with a set of sensible default RunOptions
// for dockertest.
func NewNode(repository, version string) *Node {
	return &Node{
		repository: repository,
		version:    version,
		RunOptions: &dockertest.RunOptions{
			User:       "0:0",
			Repository: repository,
			Tag:        version,
			// exposing JSON-RPC port by default to ping node after start
			ExposedPorts: []string{jrpcPort},
		},
	}
}

// SetEnvVars allows to set additional container environment variables by passing a slice
// of strings that each fit the pattern "VAR_NAME=value".
func (n *Node) SetEnvVars(vars []string) {
	n.RunOptions.Env = vars
	n.UseRunOptions()
}

// Mount sets the container mount point, which is used as the value for 'docker run --volume'.
//
// See https://docs.docker.com/engine/reference/builder/#volume
func (n *Node) Mount(mountPath string) {
	n.RunOptions.Mounts = []string{mountPath}
	n.UseRunOptions()
}

// SetCmd sets the container entry command and overrides the image CMD instruction.
//
// See https://docs.docker.com/engine/reference/builder/#cmd
func (n *Node) SetCmd(cmd []string) {
	n.RunOptions.Cmd = cmd
	n.UseRunOptions()
}

// UseRunOptions sets a flag to allow the node Manager to run the container with additional run options.
func (n *Node) UseRunOptions() {
	n.withRunOptions = true
}
