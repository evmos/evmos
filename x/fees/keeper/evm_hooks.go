package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

var _ evmtypes.EvmHooks = Hooks{}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer receives
// a share from the transaction fees paid by the user.
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	// check if the fees are globally enabled
	params := h.k.GetParams(ctx)
	if !params.EnableFees {
		return nil
	}

	contract := msg.To()
	// if the contract is not registered to receive fees, do nothing
	if contract == nil || !h.k.IsFeeRegistered(ctx, *contract) {
		return nil
	}

	withdrawAddr, found := h.k.GetWithdrawal(ctx, *contract)
	if !found {
		withdrawAddr, found = h.k.GetDeployer(ctx, *contract)
	}
	if !found {
		// no registered deployer / withdraw address for the contract
		return nil
	}

	feeDistribution := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))

	evmDenom := h.k.evmKeeper.GetParams(ctx).EvmDenom
	developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
	developerCoins := sdk.Coins{{Denom: evmDenom, Amount: developerFee.TruncateInt()}}

	return h.sendFees(ctx, *contract, withdrawAddr, developerCoins)
}

// sendFees distributes the transaction fees to the contract deployer
func (h Hooks) sendFees(
	ctx sdk.Context,
	contract common.Address,
	withdrawAddr sdk.AccAddress,
	fees sdk.Coins,
) error {
	err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, h.k.feeCollectorName, withdrawAddr, fees)
	if err != nil {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"fee collector account failed to distribute developer fees: %s",
			err.Error(),
		)
		return sdkerrors.Wrapf(err, "failed to distribute %s fees", fees.String())
	}
	return nil
}
