package upgrade

import "github.com/ory/dockertest/v3"

// Node defines evmos node params for running container
// of specific version with custom docker run arguments
type Node struct {
	repository string
	version    string

	runOptions     *dockertest.RunOptions
	withRunOptions bool
}

// NewNode creates new instance of the node and setups default dockertest RunOptions
func NewNode(repository, version string) *Node {
	return &Node{
		repository: repository,
		version:    version,
		runOptions: &dockertest.RunOptions{
			Repository: repository,
			Tag:        version,
		},
	}
}

// SetEnvVars allows to set addition container environment variables in format
// []string{ "VAR_NAME=valaue" }
func (n *Node) SetEnvVars(vars []string) {
	n.runOptions.Env = vars
	n.UseRunOptions()
}

// Mount sets container mount point, used as value for 'docker run --volume'
// https://docs.docker.com/engine/reference/builder/#volume
func (n *Node) Mount(mountPath string) {
	n.runOptions.Mounts = []string{mountPath}
	n.UseRunOptions()
}

// SetCmd sets container entry command, overriding image CMD instruction
// https://docs.docker.com/engine/reference/builder/#cmd
func (n *Node) SetCmd(cmd []string) {
	n.runOptions.Cmd = cmd
	n.UseRunOptions()
}

// UseRunOptions flags Manager to run container with additional run options
func (n *Node) UseRunOptions() {
	n.withRunOptions = true
}
