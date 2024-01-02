// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	anteutils "github.com/evmos/evmos/v16/app/ante/utils"
	"github.com/evmos/evmos/v16/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// // EthGasConsumeDecorator validates enough intrinsic gas for the transaction and
// // gas consumption.
// type EthGasConsumeDecorator struct {
// 	bankKeeper         anteutils.BankKeeper
// 	distributionKeeper anteutils.DistributionKeeper
// 	evmKeeper          EVMKeeper
// 	stakingKeeper      anteutils.StakingKeeper
// 	maxGasWanted       uint64
// }

// // AnteHandle validates that the Ethereum tx message has enough to cover intrinsic gas
// // (during CheckTx only) and that the sender has enough balance to pay for the gas cost.
// // If the balance is not sufficient, it will be attempted to withdraw enough staking rewards
// // for the payment.
// //
// // Intrinsic gas for a transaction is the amount of gas that the transaction uses before the
// // transaction is executed. The gas is a constant value plus any cost incurred by additional bytes
// // of data supplied with the transaction.
// //
// // This AnteHandler decorator will fail if:
// // - the message is not a MsgEthereumTx
// // - sender account cannot be found
// // - transaction's gas limit is lower than the intrinsic gas
// // - user has neither enough balance nor staking rewards to deduct the transaction fees (gas_limit * gas_price)
// // - transaction or block gas meter runs out of gas
// // - sets the gas meter limit
// // - gas limit is greater than the block gas meter limit
// func (egcd EthGasConsumeDecorator) AnteHandle(ctx sdktypes.Context, tx sdktypes.Tx, simulate bool, next sdktypes.AnteHandler) (sdktypes.Context, error) {
// 	gasWanted := uint64(0)
// 	// gas consumption limit already checked during CheckTx so there's no need to
// 	// verify it again during ReCheckTx
// 	if ctx.IsReCheckTx() {
// 		// Use new context with gasWanted = 0
// 		// Otherwise, there's an error on txmempool.postCheck (tendermint)
// 		// that is not bubbled up. Thus, the Tx never runs on DeliverMode
// 		// Error: "gas wanted -1 is negative"
// 		// For more information, see issue #1554
// 		// https://github.com/evmos/ethermint/issues/1554
// 		newCtx := ctx.WithGasMeter(types.NewInfiniteGasMeterWithLimit(gasWanted))
// 		return next(newCtx, tx, simulate)
// 	}
//
// 	evmParams := egcd.evmKeeper.GetParams(ctx)
// 	evmDenom := evmParams.GetEvmDenom()
// 	chainCfg := evmParams.GetChainConfig()
// 	ethCfg := chainCfg.EthereumConfig(egcd.evmKeeper.ChainID())
//
// 	blockHeight := big.NewInt(ctx.BlockHeight())
// 	homestead := ethCfg.IsHomestead(blockHeight)
// 	istanbul := ethCfg.IsIstanbul(blockHeight)
//
// 	// Use the lowest priority of all the messages as the final one.
// 	minPriority := int64(math.MaxInt64)
// 	baseFee := egcd.evmKeeper.GetBaseFee(ctx, ethCfg)
//
// 	for _, msg := range tx.GetMsgs() {
// 		_, txData, from, err := evmtypes.UnpackEthMsg(msg)
// 		if err != nil {
// 			return ctx, err
// 		}
//
// 		gasWanted, minPriority, err = ConsumeGas(
// 			ctx,
// 			egcd.bankKeeper,
// 			egcd.distributionKeeper,
// 			egcd.evmKeeper,
// 			egcd.stakingKeeper,
// 			from,
// 			txData,
// 			minPriority,
// 			gasWanted,
// 			egcd.maxGasWanted,
// 			evmDenom,
// 			baseFee,
// 			homestead,
// 			istanbul,
// 		)
//
// 		if err != nil {
// 			return ctx, err
// 		}
// 	}
//
// 	newCtx, err := CheckBlockGasLimit(ctx, gasWanted, minPriority)
// 	if err != nil {
// 		return ctx, err
// 	}
//
// 	return next(newCtx, tx, simulate)
// }

// func GetAndVerifyFees(
// 	ctx sdktypes.Context,
// 	txData evmtypes.TxData,
// 	decUtils *DecoratorUtils,
// ) (sdktypes.Coins, error) {
// 	return fees, nil
// }

// UpdateComulativeGasWanted updates the cumulative gas wanted
func UpdateComulativeGasWanted(
	ctx sdktypes.Context,
	msgGasWanted uint64,
	maxTxGasWanted uint64,
	cumulativeGasWanted uint64,
) uint64 {
	if ctx.IsCheckTx() && maxTxGasWanted != 0 {
		// We can't trust the tx gas limit, because we'll refund the unused gas.
		if msgGasWanted > maxTxGasWanted {
			cumulativeGasWanted += maxTxGasWanted
		} else {
			cumulativeGasWanted += msgGasWanted
		}
	} else {
		cumulativeGasWanted += msgGasWanted
	}
	return cumulativeGasWanted
}

type ConsumeGasKeepers struct {
	Bank         anteutils.BankKeeper
	Distribution anteutils.DistributionKeeper
	Evm          EVMKeeper
	Staking      anteutils.StakingKeeper
}

// ConsumeFeesAndEmitEvent deduces fees from sender and emits the event
func ConsumeFeesAndEmitEvent(
	ctx sdktypes.Context,
	keepers *ConsumeGasKeepers,
	fees sdktypes.Coins,
	from sdktypes.AccAddress,
) error {
	if err := deductFees(
		ctx,
		keepers,
		fees,
		from,
	); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdktypes.NewEvent(
			sdktypes.EventTypeTx,
			sdktypes.NewAttribute(sdktypes.AttributeKeyFee, fees.String()),
		),
	)
	return nil
}

// deductFee checks if the fee payer has enough funds to pay for the fees and deducts them.
// If the spendable balance is not enough, it tries to claim enough staking rewards to cover the fees.
func deductFees(
	ctx sdktypes.Context,
	keepers *ConsumeGasKeepers,
	fees sdktypes.Coins,
	feePayer sdktypes.AccAddress,
) error {
	if fees.IsZero() {
		return nil
	}

	// If the account balance is not sufficient, try to withdraw enough staking rewards
	if err := anteutils.ClaimStakingRewardsIfNecessary(
		ctx,
		keepers.Bank,
		keepers.Distribution,
		keepers.Staking,
		feePayer,
		fees,
	); err != nil {
		return err
	}

	if err := keepers.Evm.DeductTxCostsFromUserBalance(
		ctx,
		fees,
		common.BytesToAddress(feePayer),
	); err != nil {
		return errorsmod.Wrapf(err, "failed to deduct transaction costs from user balance")
	}
	return nil
}

// GetMsgPriority returns the priority of a Eth Tx capped by the minimum priority
func GetMsgPriority(
	txData evmtypes.TxData,
	minPriority int64,
	baseFee *big.Int,
) int64 {
	priority := evmtypes.GetTxPriority(txData, baseFee)

	if priority < minPriority {
		minPriority = priority
	}
	return minPriority
}

// TODO: (@fedekunze) Why is this necessary? This seems to be a duplicate from the CheckGasWanted function.
func CheckBlockGasLimit(ctx sdktypes.Context, gasWanted uint64, minPriority int64) (sdktypes.Context, error) {
	blockGasLimit := types.BlockGasLimit(ctx)

	// return error if the tx gas is greater than the block limit (max gas)

	// NOTE: it's important here to use the gas wanted instead of the gas consumed
	// from the tx gas pool. The latter only has the value so far since the
	// EthSetupContextDecorator, so it will never exceed the block gas limit.
	if gasWanted > blockGasLimit {
		return ctx, errorsmod.Wrapf(
			errortypes.ErrOutOfGas,
			"tx gas (%d) exceeds block gas limit (%d)",
			gasWanted,
			blockGasLimit,
		)
	}

	// Set tx GasMeter with a limit of GasWanted (i.e. gas limit from the Ethereum tx).
	// The gas consumed will be then reset to the gas used by the state transition
	// in the EVM.

	// FIXME: use a custom gas configuration that doesn't add any additional gas and only
	// takes into account the gas consumed at the end of the EVM transaction.
	ctx = ctx.
		WithGasMeter(types.NewInfiniteGasMeterWithLimit(gasWanted)).
		WithPriority(minPriority)

	return ctx, nil
}

// UpdateCumulativeTxFee updates the cumulative transaction fee
func UpdateCumulativeTxFee(
	cumulativeTxFee sdktypes.Coins,
	msgFee *big.Int,
	denom string,
) sdktypes.Coins {
	return cumulativeTxFee.Add(
		sdktypes.Coin{
			Denom:  denom,
			Amount: sdkmath.NewIntFromBigInt(msgFee),
		},
	)
}
