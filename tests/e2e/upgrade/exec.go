package upgrade

import (
	"fmt"
)

// E2eTxArgs contains the arguments to build a CLI transaction command from.
type E2eTxArgs struct {
	ModuleName string
	SubCommand string
	Args       []string
	ChainID    string
	From       string
}

// CreateModuleTxExec creates the execution command for an Evmos transaction.
func (m *Manager) CreateModuleTxExec(txArgs E2eTxArgs) (string, error) {
	cmd := []string{
		"evmosd",
		"tx",
		txArgs.ModuleName,
		txArgs.SubCommand,
	}
	cmd = append(cmd, txArgs.Args...)
	cmd = append(cmd,
		fmt.Sprintf("--chain-id=%s", txArgs.ChainID),
		"--keyring-backend=test",
		"--log_format=json",
		"--fees=500aevmos",
		"--gas=auto",
		fmt.Sprintf("--from=%s", txArgs.From),
	)
	return m.CreateExec(cmd, m.ContainerID())
}
