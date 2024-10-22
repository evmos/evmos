package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/Eidon-AI/eidon-chain/v20/app"
	eidond "github.com/Eidon-AI/eidon-chain/v20/cmd/eidond"
	"github.com/Eidon-AI/eidon-chain/v20/utils"
)

func TestInitCmd(t *testing.T) {
	err := app.InitializeAppConfiguration("eidon-chain_9001-1")
	require.NoError(t, err)
	target := t.TempDir()

	rootCmd, _ := eidond.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",       // Test the init cmd
		"eidon-chain-test", // Moniker
		fmt.Sprintf("--home=%s", target),
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, utils.TestnetChainID+"-1"),
	})

	err = svrcmd.Execute(rootCmd, "eidond", app.DefaultNodeHome)
	require.NoError(t, err)
}

func TestAddKeyLedgerCmd(t *testing.T) {
	rootCmd, _ := eidond.NewRootCmd()
	rootCmd.SetArgs([]string{
		"keys",
		"add",
		"dev0",
		fmt.Sprintf("--%s", flags.FlagUseLedger),
	})

	err := svrcmd.Execute(rootCmd, "EVMOSD", app.DefaultNodeHome)
	require.Error(t, err)
}
