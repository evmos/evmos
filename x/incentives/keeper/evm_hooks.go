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
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmostypes "github.com/evmos/evmos/v11/types"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"

	"github.com/evmos/evmos/v11/x/incentives/types"
)

var _ evmtypes.EvmHooks = Hooks{}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with an incentivized contract, the participants's GasUsed is
// added to its gasMeter.
func (k Keeper) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	// check if the Incentives are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableIncentives {
		return nil
	}

	contract := msg.To()
	participant := msg.From()

	// If theres no incentive registered for the contract, do nothing
	if contract == nil || !k.IsIncentiveRegistered(ctx, *contract) {
		return nil
	}

	// safety check: only distribute incentives to EOAs.
	acc := k.accountKeeper.GetAccount(ctx, participant.Bytes())
	if acc == nil {
		return nil
	}

	ethAccount, ok := acc.(evmostypes.EthAccountI)
	if ok && ethAccount.Type() == evmostypes.AccountTypeContract {
		return nil
	}

	k.addGasToIncentive(ctx, *contract, receipt.GasUsed)
	k.addGasToParticipant(ctx, *contract, participant, receipt.GasUsed)

	defer func() {
		telemetry.IncrCounter(
			1,
			"tx", "msg", "ethereum_tx", types.ModuleName, "total",
		)

		if receipt.GasUsed != 0 {
			telemetry.IncrCounter(
				float32(receipt.GasUsed),
				"tx", "msg", "ethereum_tx", types.ModuleName, "gas_used", "total",
			)
		}
	}()

	return nil
}

// addGasToIncentive adds gasUsed to an incentive's cumulated totalGas
func (k Keeper) addGasToIncentive(
	ctx sdk.Context,
	contract common.Address,
	gasUsed uint64,
) {
	// NOTE: existence of contract incentive is already checked
	incentive, _ := k.GetIncentive(ctx, contract)
	incentive.TotalGas += gasUsed
	k.SetIncentive(ctx, incentive)
}

// addGasToParticipant adds gasUsed to a participant's gas meter's cumulative
// gas used
func (k Keeper) addGasToParticipant(
	ctx sdk.Context,
	contract, participant common.Address,
	gasUsed uint64,
) {
	previousGas, found := k.GetGasMeter(ctx, contract, participant)
	if found {
		gasUsed += previousGas
	}

	gm := types.NewGasMeter(contract, participant, gasUsed)
	k.SetGasMeter(ctx, gm)
}
