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
