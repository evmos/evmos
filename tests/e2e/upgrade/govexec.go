package upgrade

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ory/dockertest/v3/docker"
)

// RunExec runs the provided docker exec call
func (m *Manager) RunExec(ctx context.Context, exec string) (outBuf bytes.Buffer, errBuf bytes.Buffer, err error) {
	err = m.pool.Client.StartExec(exec, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	return
}

// CreateExec creates docker exec command for specified container
func (m *Manager) CreateExec(cmd []string, containerID string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exec, err := m.pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		User:         "root",
		Container:    containerID,
		Cmd:          cmd,
	})
	if err != nil {
		return "", err
	}
	return exec.ID, nil
}

// CreateSubmitProposalExec creates a gov tx to submit an upgrade proposal to the chain
func (m *Manager) CreateSubmitProposalExec(targetVersion, chainID string, upgradeHeight uint) (string, error) {
	cmd := []string{
		"evmosd",
		"tx", "gov", "submit-proposal",
		"software-upgrade", targetVersion,
		"--title=\"TEST\"",
		"--description=\"Test upgrade proposal\"",
		fmt.Sprintf("--upgrade-height=%d", upgradeHeight),
		"--upgrade-info=\"\"",
		fmt.Sprintf("--chain-id=%s", chainID),
		"--from=mykey", "-b=block",
		"--yes", "--keyring-backend=test",
		"--log_format=json", "--fees=20aevmos",
		"--gas=auto",
	}
	// increment proposal counter to use proposal number for deposit && voting
	m.proposalCounter++
	return m.CreateExec(cmd, m.ContainerID())
}

// CreateDepositProposalExec creates a gov tx to deposit for the current upgrade proposal
func (m *Manager) CreateDepositProposalExec(chainID string) (string, error) {
	cmd := []string{
		"evmosd",
		"tx",
		"gov",
		"deposit",
		fmt.Sprint(m.proposalCounter),
		"10000000aevmos",
		"--from=mykey",
		fmt.Sprintf("--chain-id=%s", chainID),
		"-b=block",
		"--yes",
		"--keyring-backend=test",
		"--fees=20aevmos",
		"--gas=auto",
	}

	return m.CreateExec(cmd, m.ContainerID())
}

// CreateVoteProposalExec creates gov tx to vote 'yes' on the current upgrade proposal
func (m *Manager) CreateVoteProposalExec(chainID string) (string, error) {
	cmd := []string{
		"evmosd",
		"tx",
		"gov",
		"vote",
		fmt.Sprint(m.proposalCounter),
		"yes",
		"--from=mykey",
		fmt.Sprintf("--chain-id=%s", chainID),
		"-b=block",
		"--yes",
		"--keyring-backend=test",
		"--fees=20aevmos",
		"--gas=auto",
	}
	return m.CreateExec(cmd, m.ContainerID())
}
