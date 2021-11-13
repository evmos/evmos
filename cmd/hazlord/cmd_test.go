package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/hazlorlabs/hsc/app"
	hazlord "github.com/hazlorlabs/hsc/cmd/hazlord"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := hazlord.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",        // Test the init cmd
		"hazlor-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, "hazlor_7878-1"),
	})

	err := svrcmd.Execute(rootCmd, app.DefaultNodeHome)
	require.NoError(t, err)
}
