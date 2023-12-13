// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

// DeleteRegisterIncentivesProposals deletes the RegisterIncentives proposals from the store
// func DeleteRegisterIncentivesProposals(ctx sdk.Context, gk govkeeper.Keeper, logger log.Logger) {
//	// Delete the only RegisterIncentives proposal
//	gk.IterateProposals(ctx, func(proposal govtypes.Proposal) bool {
//		// Check if proposal is a RegisterIncentives proposal
//		msgs, err := proposal.GetMsgs()
//		if err != nil {
//			logger.Error("failed to get proposal messages", "error", err.Error())
//			return false
//		}
//
//		for _, msg := range msgs {
//			legacyContentMsg, ok := msg.(*govtypes.MsgExecLegacyContent)
//			if !ok {
//				continue
//			}
//
//			_, ok = legacyContentMsg.Content.GetCachedValue().(*incentives.RegisterIncentiveProposal)
//			if ok {
//				gk.DeleteProposal(ctx, proposal.Id)
//				return true
//			}
//		}
//		return true
//	})
// }
