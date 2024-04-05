// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"github.com/cometbft/cometbft/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	incentives "github.com/evmos/evmos/v17/x/incentives/types"
)

// DeleteIncentivesProposals deletes the RegisterIncentives & CancelIncentiveProposal proposals from the store
// because the module was deprecated
func DeleteIncentivesProposals(ctx sdk.Context, gk govkeeper.Keeper, logger log.Logger) {
	// Delete the only incentives module proposals
	gk.IterateProposals(ctx, func(proposal govtypes.Proposal) bool {
		// Check if proposal is a RegisterIncentives or CancelIncentiveProposal proposal
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

			_, ok = legacyContentMsg.Content.GetCachedValue().(*incentives.RegisterIncentiveProposal)
			if ok {
				gk.DeleteProposal(ctx, proposal.Id)
				continue
			}

			_, ok = legacyContentMsg.Content.GetCachedValue().(*incentives.CancelIncentiveProposal)
			if ok {
				gk.DeleteProposal(ctx, proposal.Id)
				continue
			}
		}
		return false
	})
}
