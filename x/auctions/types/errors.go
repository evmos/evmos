// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrAuctionDisabled            = errorsmod.Register(ModuleName, 1, "Burn Auction is disabled")
	ErrBidMustBeHigherThanCurrent = errorsmod.Register(ModuleName, 2, "Bid must be higher than current one")
	ErrInvalidDenom               = errorsmod.Register(ModuleName, 3, "Invalid denom")
)
