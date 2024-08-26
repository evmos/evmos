// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

const (
	// ModuleName the name of the module
	ModuleName = "auctions"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName

	// AuctionCollectorName the name of the asset auction collector module account
	AuctionCollectorName = "auction_collector"
)

// prefix bytes for the auctions persistent store
const (
	prefixAuctionRound = iota + 1
	prefixBid
)

// KVStore key prefixes
var (
	KeyPrefixAuctionRound = []byte{prefixAuctionRound}
	KeyPrefixBid          = []byte{prefixBid}
)
