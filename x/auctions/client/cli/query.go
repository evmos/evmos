package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/evmos/evmos/v18/x/auctions/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for the inflation module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the auctions module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	return cmd
}
