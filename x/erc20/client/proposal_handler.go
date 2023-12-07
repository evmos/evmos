// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/evmos/evmos/v16/x/erc20/client/cli"
)

var (
	RegisterERC20ProposalHandler         = govclient.NewProposalHandler(cli.NewRegisterERC20ProposalCmd)
	ToggleTokenConversionProposalHandler = govclient.NewProposalHandler(cli.NewToggleTokenConversionProposalCmd)
)
