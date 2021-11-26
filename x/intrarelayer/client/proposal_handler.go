package cli

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// TODO: add REST
var (
	RegisterCoinProposalHandler  = govclient.NewProposalHandler(NewRegisterCoinProposalCmd, nil)
	RegisterERC20ProposalHandler = govclient.NewProposalHandler(NewRegisterERC20ProposalCmd, nil)
	// TODO: add other proposal types
)
