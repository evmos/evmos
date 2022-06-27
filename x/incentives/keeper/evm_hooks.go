package keeper

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v5/x/incentives/types"
)

var _ evmtypes.EvmHooks = Hooks{}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with an incentivized contract, the participants's GasUsed is
// added to its gasMeter.
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	// check if the Incentives are globally enabled
	params := h.k.GetParams(ctx)
	if !params.EnableIncentives {
		return nil
	}

	contract := msg.To()
	participant := msg.From()

	// If theres no incentive registered for the contract, do nothing
	if contract == nil || !h.k.IsIncentiveRegistered(ctx, *contract) {
		return nil
	}

	// safety check: only distribute incentives to EOAs.
	acc := h.k.accountKeeper.GetAccount(ctx, participant.Bytes())
	if acc == nil {
		return nil
	}

	ethAccount, ok := acc.(ethermint.EthAccountI)
	if ok && ethAccount.Type() == ethermint.AccountTypeContract {
		return nil
	}

	h.addGasToIncentive(ctx, *contract, receipt.GasUsed)
	h.addGasToParticipant(ctx, *contract, participant, receipt.GasUsed)

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
