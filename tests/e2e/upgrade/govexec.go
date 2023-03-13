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
func (m *Manager) CreateSubmitProposalExec(targetVersion, chainID string, upgradeHeight uint, legacy bool, flags ...string) (string, error) {
	var upgradeInfo, proposalType string
	if legacy {
		upgradeInfo = "--no-validate"
		proposalType = "submit-legacy-proposal"
	} else {
		upgradeInfo = "--upgrade-info=\"\""
		proposalType = "submit-proposal"
	}
	cmd := []string{
		"evmosd",
		"tx",
		"gov",
		proposalType,
		"software-upgrade",
		targetVersion,
		"--title=\"TEST\"",
		"--deposit=10000000aevmos",
		"--description=\"Test upgrade proposal\"",
		fmt.Sprintf("--upgrade-height=%d", upgradeHeight),
		upgradeInfo,
		fmt.Sprintf("--chain-id=%s", chainID),
		"--from=mykey",
		"-b=block",
		"--yes",
		"--keyring-backend=test",
		"--log_format=json",
	}
	cmd = append(cmd, flags...)
	// increment proposal counter to use proposal number for deposit && voting
	m.proposalCounter++
	return m.CreateExec(cmd, m.ContainerID())
}

// CreateDepositProposalExec creates a gov tx to deposit for the proposal with the given id
func (m *Manager) CreateDepositProposalExec(chainID string, id int) (string, error) {
	cmd := []string{
		"evmosd",
		"tx",
		"gov",
		"deposit",
		fmt.Sprint(id),
		"10000000aevmos",
		"--from=mykey",
		fmt.Sprintf("--chain-id=%s", chainID),
		"-b=block",
		"--yes",
		"--keyring-backend=test",
		"--log_format=json",
		"--fees=500aevmos",
		"--gas=500000",
	}

	return m.CreateExec(cmd, m.ContainerID())
}

// CreateVoteProposalExec creates gov tx to vote 'yes' on the proposal with the given id
func (m *Manager) CreateVoteProposalExec(chainID string, id int, flags ...string) (string, error) {
	cmd := []string{
		"evmosd",
		"tx",
		"gov",
		"vote",
		fmt.Sprint(id),
		"yes",
		"--from=mykey",
		fmt.Sprintf("--chain-id=%s", chainID),
		"-b=block",
		"--yes",
		"--keyring-backend=test",
		"--log_format=json",
	}
	cmd = append(cmd, flags...)
	return m.CreateExec(cmd, m.ContainerID())
}
