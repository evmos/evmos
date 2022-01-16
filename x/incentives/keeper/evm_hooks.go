package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

var _ evmtypes.EvmHooks = Hooks{}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with an incentivized contract, the participants's GasUsed is
// added to its gasMeter.
func (h Hooks) PostTxProcessing(ctx sdk.Context, participant common.Address, contract *common.Address, receipt *ethtypes.Receipt) error {
	// If theres no incentive registered for the contract, do nothing
	if contract == nil || !h.k.IsIncentiveRegistered(ctx, *contract) {
		return nil
	}

	h.addGasToIncentive(ctx, *contract, receipt.GasUsed)
	h.addGasToParticipant(ctx, *contract, participant, receipt.GasUsed)

	return nil
}

// addGasToIncentive adds gasUsed to an incentive's cumulated totalGas
func (h Hooks) addGasToIncentive(
	ctx sdk.Context,
	contract common.Address,
	gasUsed uint64,
) {
	// NOTE: existence of contract incentive is already checked
	incentive, _ := h.k.GetIncentive(ctx, contract)
	incentive.TotalGas += gasUsed
	h.k.SetIncentive(ctx, incentive)
}

// addGasToParticipant adds gasUsed to a participant's gas meter's cumulative
// gas used
func (h Hooks) addGasToParticipant(
	ctx sdk.Context,
	contract, participant common.Address,
	gasUsed uint64,
) {
	previousGas, found := h.k.GetGasMeter(ctx, contract, participant)
	if found {
		gasUsed += previousGas
	}

	gm := types.NewGasMeter(contract, participant, gasUsed)
	h.k.SetGasMeter(ctx, gm)
}
