package upgrade

import "github.com/ory/dockertest/v3"

var baseCmd = []string{"evmosd", "start"}

type Node struct {
	repository string
	version    string

	runOptions     *dockertest.RunOptions
	withRunOptions bool
}

func NewNode(repository, version string) *Node {
	return &Node{
		repository: repository,
		version:    version,
		runOptions: &dockertest.RunOptions{
			Repository: repository,
			Tag:        version,
			Cmd:        baseCmd,
		},
	}
}

func (n *Node) Mount(mountPath string) *Node {
	n.runOptions.Mounts = []string{mountPath}
	n.UseRunOptions()
	return n
}

func (n *Node) Cmd(cmd []string) *Node {
	n.runOptions.Cmd = cmd
	n.UseRunOptions()
	return n
}

func (n *Node) UseRunOptions() {
	n.withRunOptions = true
}
