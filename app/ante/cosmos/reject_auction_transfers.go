// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package cosmos

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	auctionstypes "github.com/evmos/evmos/v19/x/auctions/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// RejectAuctionTransfers prevents direct bank sends to the auctions module account
type RejectAuctionTransfers struct {
	accountKeeper evmtypes.AccountKeeper
}

func NewRejectAuctionTransfers(accountKeeper evmtypes.AccountKeeper) RejectAuctionTransfers {
	return RejectAuctionTransfers{
		accountKeeper: accountKeeper,
	}
}

// AnteHandle rejects MsgSend transactions to the auctions module account
func (rat RejectAuctionTransfers) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		if _, ok := msg.(*banktypes.MsgSend); ok {
			sendMsg := msg.(*banktypes.MsgSend)
			moduleAddress := rat.accountKeeper.GetModuleAddress(auctionstypes.ModuleName)
			if sendMsg.ToAddress == moduleAddress.String() {
				return ctx, errorsmod.Wrapf(
					errortypes.ErrUnauthorized,
					"cannot send funds directly to the auctions module account",
				)
			}
		}
	}
	return next(ctx, tx, simulate)
}
