// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package upgrade

import (
	"fmt"
)

// QueryArgs is a struct to hold the relevant information for the query commands.
type QueryArgs struct {
	// Module is the module name to query
	Module string
	// SubCommand is the subcommand to query
	SubCommand string
	// Args are the arguments to the query command
	Args []string
	// ChainID is the chain ID to query
	ChainID string
}

// Validate performs basic validation on the QueryArgs.
func (q QueryArgs) Validate() error {
	if q.Module == "" {
		return fmt.Errorf("module cannot be empty")
	}
	if q.SubCommand == "" {
		return fmt.Errorf("subcommand cannot be empty")
	}
	if q.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	return nil
}

// CreateModuleQueryExec creates a Evmos module query.
func (m *Manager) CreateModuleQueryExec(args QueryArgs) (string, error) {
	// Check that valid args were provided
	if err := args.Validate(); err != nil {
		return "", err
	}

	// Build the query command
	cmd := []string{
		"evmosd",
		"q",
		args.Module,
		args.SubCommand,
	}

	if len(args.Args) > 0 {
		cmd = append(cmd, args.Args...)
	}

	cmd = append(cmd,
		fmt.Sprintf("--chain-id=%s", args.ChainID),
		"--output=json",
	)

	return m.CreateExec(cmd, m.ContainerID())
}
