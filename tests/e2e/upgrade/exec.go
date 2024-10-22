// Copyright Tharsis Labs Ltd.(Eidon-chain)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/Eidon-AI/eidon-chain/blob/main/LICENSE)

package upgrade

import (
	"fmt"
)

// E2ETxArgs contains the arguments to build a CLI transaction command from.
type E2ETxArgs struct {
	ModuleName string
	SubCommand string
	Args       []string
	ChainID    string
	From       string
}

// CreateModuleTxExec creates the execution command for an Eidon-chain transaction.
func (m *Manager) CreateModuleTxExec(txArgs E2ETxArgs) (string, error) {
	cmd := []string{
		"eidond",
		"tx",
		txArgs.ModuleName,
		txArgs.SubCommand,
	}
	cmd = append(cmd, txArgs.Args...)
	cmd = append(cmd,
		fmt.Sprintf("--chain-id=%s", txArgs.ChainID),
		"--keyring-backend=test",
		"--output=text",
		"--fees=50000000000000aeidon-chain",
		"--gas=auto",
		fmt.Sprintf("--from=%s", txArgs.From),
		"--yes",
	)
	return m.CreateExec(cmd, m.ContainerID())
}
