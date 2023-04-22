// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

// claim module event types
const (
	EventTypeClaim              = "claim"
	EventTypeMergeClaimsRecords = "merge_claims_records"

	AttributeKeyActionType             = "action"
	AttributeKeyRecipient              = "recipient"
	AttributeKeyClaimedCoins           = "claimed_coins"
	AttributeKeyFundCommunityPoolCoins = "fund_community_pool_coins"
)
