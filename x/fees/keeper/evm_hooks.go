package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
func (h Hooks) PostTxProcessing(ctx sdk.Context, from common.Address, contract *common.Address, receipt *ethtypes.Receipt) error {
	// check if the fees are globally enabled
	params := h.k.GetParams(ctx)
	if !params.EnableFees {
		return nil
	}

	// If theres no fees registered for the contract, do nothing
	if contract == nil || !h.k.IsFeeRegistered(ctx, *contract) {
		return nil
	}
	cfg, err := h.k.evmKeeper.EVMConfig(ctx)
	if err != nil {
		return err
	}

	feeContract, ok := h.k.GetFee(ctx, *contract)
	if !ok {
		return nil
	}

	withdrawAddr, err := sdk.AccAddressFromBech32(feeContract.WithdrawAddress)
	if err != nil {
		return err
	}

	feeDistribution := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(cfg.BaseFee))
	developerFee := sdk.NewDecFromInt(feeDistribution).Mul(h.k.GetParams(ctx).DeveloperPercentage)
	developerCoins := sdk.Coins{sdk.NewCoin(cfg.Params.EvmDenom, developerFee.TruncateInt())}

	// TODO burn 100% - devP - validatorP

	return h.addFeesToOwner(ctx, *contract, withdrawAddr, developerCoins)
}

// addGasToParticipant adds gasUsed to a participant's gas meter's cumulative
// gas used
func (h Hooks) addFeesToOwner(
	ctx sdk.Context,
	contract common.Address,
	withdrawAddr sdk.AccAddress,
	fees sdk.Coins,
) error {
	err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, h.k.feeCollectorName, withdrawAddr, fees)
	if err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "fee collector account failed to distribute developer fees: %s", err.Error())
		return sdkerrors.Wrapf(err, "failed to distribute %s fees", fees.String())
	}
	return nil
}
