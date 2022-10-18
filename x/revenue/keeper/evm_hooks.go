package keeper

import (
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v9/x/revenue/types"
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
	developerFee := txFee.ToDec().Mul(params.DeveloperShares).TruncateInt()
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
		return sdkerrors.Wrapf(
			err,
			"fee collector account failed to distribute developer fees (%s) to withdraw address %s. contract %s",
			fees, withdrawer, contract,
		)
	}

	defer func() {
		if developerFee.IsInt64() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "distribute", "total"},
				float32(developerFee.Int64()),
				[]metrics.Label{
					telemetry.NewLabel("sender", msg.From().String()),
					telemetry.NewLabel("withdraw_address", withdrawer.String()),
					telemetry.NewLabel("contract", revenue.ContractAddress),
				},
			)
		}
	}()

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeDistributeDevRevenue,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.From().String()),
				sdk.NewAttribute(types.AttributeKeyContract, contract.String()),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, withdrawer.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, developerFee.String()),
			),
		},
	)

	return nil
}
