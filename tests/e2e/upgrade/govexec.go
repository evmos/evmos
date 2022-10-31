package upgrade

import (
	"bytes"
	"context"
	"log"

	"github.com/ory/dockertest/v3/docker"
)

func (m *Manager) RunExec(ctx context.Context, execID string) (bytes.Buffer, bytes.Buffer, error) {
	var outBuf, errBuf bytes.Buffer
	err := m.pool.Client.StartExec(execID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	log.Println(outBuf.String(), errBuf.String(), err)
	return outBuf, errBuf, err
}

func (m *Manager) CreateProposalExec(ctx context.Context, targetVersion string) (string, error) {
	exec, err := m.pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    m.ContainerID(),
		User:         "root",
		Cmd: []string{
			"evmosd",
			"tx", "gov", "submit-proposal",
			"software-upgrade", targetVersion,
			"--title=\"TEST\"",
			"--description=\"Test upgrade proposal\"",
			"--upgrade-height=50",
			"--upgrade-info=\"\"",
			"--chain-id=evmos_9000-1",
			"--from=mykey", "-b=block",
			"--yes", "--keyring-backend=test",
			"--log_format=json", "--fees=20aevmos",
			"--gas=auto",
		},
	})

	return exec.ID, err
}

func (m *Manager) CreateDepositProposalExec(ctx context.Context) (string, error) {
	exec, err := m.pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    m.ContainerID(),
		User:         "root",
		Cmd: []string{
			"evmosd",
			"tx",
			"gov",
			"deposit",
			"1",
			"10000000aevmos",
			"--from=mykey",
			"--chain-id=evmos_9000-1",
			"-b=block",
			"--yes",
			"--keyring-backend=test",
			"--fees=20aevmos",
			"--gas=auto",
		},
	})
	return exec.ID, err
}

func (m *Manager) CreateVoteProposalExec(ctx context.Context) (string, error) {
	exec, err := m.pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    m.ContainerID(),
		User:         "root",
		Cmd: []string{
			"evmosd",
			"tx",
			"gov",
			"vote",
			"1",
			"yes",
			"--from=mykey",
			"--chain-id=evmos_9000-1",
			"-b=block",
			"--yes",
			"--keyring-backend=test",
			"--fees=20aevmos",
			"--gas=auto",
		},
	})
	return exec.ID, err
}
