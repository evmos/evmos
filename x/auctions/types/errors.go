// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrAuctionDisabled            = errorsmod.Register(ModuleName, 1, "auctions are disabled")
	ErrBidMustBeHigherThanCurrent = errorsmod.Register(ModuleName, 2, "bid must be higher than current one")
	ErrInvalidDenom               = errorsmod.Register(ModuleName, 3, "invalid denom")
	ErrInvalidRound               = errorsmod.Register(ModuleName, 4, "invalid round")
	ErrNegativeBid                = errorsmod.Register(ModuleName, 5, "negative bid")
)
