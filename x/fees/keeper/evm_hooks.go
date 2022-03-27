package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

var _ evmtypes.EvmHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with an incentivized contract, the owner's GasUsed is
// added to its gasMeter.
func (h Hooks) PostTxProcessing(ctx sdk.Context, owner common.Address, contract *common.Address, receipt *ethtypes.Receipt) error {
	fmt.Println("---PostTxProcessing", owner, contract, receipt.GasUsed)
	// check if the fees are globally enabled
	params := h.k.GetParams(ctx)
	if !params.EnableFees {
		return nil
	}

	// If theres no fees registered for the contract, do nothing
	if contract == nil || !h.k.IsFeeRegistered(ctx, *contract) {
		return nil
	}

	h.addFeesToOwner(ctx, *contract, owner, receipt.GasUsed)

	return nil
}

// addGasToParticipant adds gasUsed to a participant's gas meter's cumulative
// gas used
func (h Hooks) addFeesToOwner(
	ctx sdk.Context,
	contract, owner common.Address,
	fees uint64,
) {
	fmt.Println("--addFeesToOwner", contract, owner, fees)
	// transfer fees from module to owner
}
