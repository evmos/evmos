package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/tharsis/evmos/x/incentives/client/cli"
	// "github.com/tharsis/evmos/x/incentives/client/rest"
)

// TODO: REST
var (
	RegisterIncentiveProposalHandler = govclient.NewProposalHandler(cli.NewRegisterIncentiveProposalCmd, nil)
	CancelIncentiveProposalHandler   = govclient.NewProposalHandler(cli.NewCancelIncentiveProposalCmd, nil)
)
