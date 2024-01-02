// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"github.com/cometbft/cometbft/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// deprecatedProposals is a map of the TypeURL
// of the deprecated proposal types
var deprecatedProposals = map[string]struct{}{
	"/evmos.incentives.v1.RegisterIncentiveProposal": {},
	"/evmos.incentives.v1.CancelIncentiveProposal":   {},
	"/evmos.erc20.v1.RegisterCoinProposal":           {},
}

// DeleteDeprecatedProposals deletes the RegisterCoin, RegisterIncentives & CancelIncentiveProposal
// proposals from the store because were deprecated
func DeleteDeprecatedProposals(ctx sdk.Context, gk govkeeper.Keeper, logger log.Logger) {
	gk.IterateProposals(ctx, func(proposal govtypes.Proposal) bool {
		// Check if proposal is a deprecated proposal
		msgs, err := proposal.GetMsgs()
		if err != nil {
			logger.Error("failed to get proposal messages", "error", err.Error())
			return false
		}

		for _, msg := range msgs {
			legacyContentMsg, ok := msg.(*govtypes.MsgExecLegacyContent)
			if !ok {
				continue
			}

			if _, deprecated := deprecatedProposals[legacyContentMsg.Content.TypeUrl]; !deprecated {
				continue
			}

			gk.DeleteProposal(ctx, proposal.Id)
		}
		return false
	})
}
