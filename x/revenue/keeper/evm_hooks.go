// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v10/x/revenue/types"
)

var _ evmtypes.EvmHooks = Hooks{}

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer (or, if set,
// the withdraw address) receives a share from the transaction fees paid by the
// transaction sender.
func (k Keeper) PostTxProcessing(
	ctx sdk.Context,
	msg core.Message,
	receipt *ethtypes.Receipt,
) error {
	// check if the fees are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableRevenue {
		return nil
	}

	contract := msg.To()
	if contract == nil {
		return nil
	}

	// if the contract is not registered to receive fees, do nothing
	revenue, found := k.GetRevenue(ctx, *contract)
	if !found {
		return nil
	}

	withdrawer := revenue.GetWithdrawerAddr()
	if len(withdrawer) == 0 {
		withdrawer = revenue.GetDeployerAddr()
	}

	txFee := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	developerFee := (params.DeveloperShares).MulInt(txFee).TruncateInt()
	evmDenom := k.evmKeeper.GetParams(ctx).EvmDenom
	fees := sdk.Coins{{Denom: evmDenom, Amount: developerFee}}

	// distribute the fees to the contract deployer / withdraw address
	err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		k.feeCollectorName,
		withdrawer,
		fees,
	)
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"fee collector account failed to distribute developer fees (%s) to withdraw address %s. contract %s",
			fees, withdrawer, contract,
		)
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventDistributeRevenue{
		Sender:            msg.From().String(),
		Contract:          contract.String(),
		WithdrawerAddress: withdrawer.String(),
		Amount:            developerFee.String(),
	})

	if err != nil {
		k.Logger(ctx).Error(err.Error())
	}

	return nil
}
