package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/berachain/core/app"
	berachaind "github.com/berachain/core/cmd/berachaind"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := berachaind.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",       // Test the init cmd
		"evmos-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, "bera_9000-1"),
	})

	err := svrcmd.Execute(rootCmd, app.DefaultNodeHome)
	require.NoError(t, err)
}
