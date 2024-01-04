package upgrade

import (
	"fmt"
	"strings"
)

// E2eTxArgs contains the arguments to build a CLI transaction command from.
type E2eTxArgs struct {
	moduleName string
	subCommand string
	args       []string
	chainID    string
	from       string
}

// CreateModuleTxExec creates the execution command for an Evmos transaction.
func (m *Manager) CreateModuleTxExec(txArgs E2eTxArgs) (string, error) {
	cmd := []string{
		"evmosd",
		"tx",
		txArgs.moduleName,
		txArgs.subCommand,
		strings.Join(txArgs.args, " "),
		fmt.Sprintf("--chain-id=%s", txArgs.chainID),
		"--keyring-backend=test",
		"--log_format=json",
		"--fees=500aevmos",
		"--gas=auto",
		fmt.Sprintf("--from=%s", txArgs.from),
	}
	return m.CreateExec(cmd, m.ContainerID())
}
