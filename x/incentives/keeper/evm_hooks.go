package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

var _ evmtypes.EvmHooks = (*Keeper)(nil)

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with an incentivized contract, the participants's GasUsed is
// added to its gasMeter.
func (k Keeper) PostTxProcessing(ctx sdk.Context, participant common.Address, contract *common.Address, receipt *ethtypes.Receipt) error {

	// If theres no incentive registered for the contract, do nothing
	if contract == nil || !k.IsIncentiveRegistered(ctx, *contract) {
		return nil
	}

	if err := k.addGasToIncentive(ctx, *contract, receipt.GasUsed); err != nil {
		return err
	}

	k.addGasToParticipant(ctx, *contract, participant, receipt.GasUsed)

	return nil
}

// addGasToIncentive adds gasUsed to an incentive's cumulated totalGas
func (k Keeper) addGasToIncentive(
	ctx sdk.Context,
	contract common.Address,
	gasUsed uint64,
) error {
	incentive, found := k.GetIncentive(ctx, contract)
	if !found {
		return sdkerrors.Wrapf(
			types.ErrInternalIncentive,
			"incentive for contract %v not found during addGasToIncentive()", contract,
		)
	}

	incentive.TotalGas += gasUsed
	k.SetIncentive(ctx, incentive)
	return nil
}

// addGasToParticipant adds gasUsed to a participant's gas meter's cumulative
// gas used
func (k Keeper) addGasToParticipant(
	ctx sdk.Context,
	contract, participant common.Address,
	gasUsed uint64,
) {
	previousGas, found := k.GetIncentiveGasMeter(ctx, contract, participant)
	if found {
		gasUsed += previousGas
	}

	gm := types.NewGasMeter(contract, participant, gasUsed)
	k.SetGasMeter(ctx, gm)
}
