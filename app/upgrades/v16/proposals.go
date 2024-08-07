// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

// DeleteIncentivesProposals deletes the RegisterIncentives & CancelIncentiveProposal proposals from the store
// because the module was deprecated
func DeleteIncentivesProposals(sdk.Context, govkeeper.Keeper, log.Logger) {
	// MODULE WAS ALREADY DELETED
	// AND MIGRATIO PERFORMED
}
