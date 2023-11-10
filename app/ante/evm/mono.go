// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	"math"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	anteutils "github.com/evmos/evmos/v15/app/ante/utils"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

var _ sdk.AnteDecorator = &EthSetupContextDecorator{}

// MonoDecorator is a single decorator that handles all the prechecks for
// ethereum transactions.
type MonoDecorator struct {
	accountKeeper      evmtypes.AccountKeeper
	bankKeeper         evmtypes.BankKeeper
	feeMarketKeeper    FeeMarketKeeper
	evmKeeper          EVMKeeper
	distributionKeeper anteutils.DistributionKeeper
	stakingKeeper      anteutils.StakingKeeper
	maxGasWanted       uint64
}

// NewMonoDecorator creates a new MonoDecorator
func NewMonoDecorator(
	accountKeeper evmtypes.AccountKeeper,
	bankKeeper evmtypes.BankKeeper,
	feeMarketKeeper FeeMarketKeeper,
	evmKeeper EVMKeeper,
) MonoDecorator {
	return MonoDecorator{
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		feeMarketKeeper: feeMarketKeeper,
		evmKeeper:       evmKeeper,
	}
}

// AnteHandle handles the entire decorator chain using a mono decorator.
func (md MonoDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	accountExpenses := make(map[string]*EthVestingExpenseTracker)

	var txFeeInfo *txtypes.Fee
	if !ctx.IsReCheckTx() {
		txFeeInfo, err = ValidateTx(tx)
		if err != nil {
			return ctx, err
		}
	}

	// 1. setup ctx
	ctx, err = SetupContext(ctx, tx, md.evmKeeper)
	if err != nil {
		return ctx, err
	}

	evmParams := md.evmKeeper.GetParams(ctx)
	chainCfg := evmParams.GetChainConfig()
	ethCfg := chainCfg.EthereumConfig(md.evmKeeper.ChainID())
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))
	allowUnprotectedTxs := evmParams.GetAllowUnprotectedTxs()
	blockHeight := big.NewInt(ctx.BlockHeight())
	isLondon := ethCfg.IsLondon(blockHeight)
	isHomestead := ethCfg.IsHomestead(blockHeight)
	isIstanbul := ethCfg.IsIstanbul(blockHeight)

	baseFee := md.evmKeeper.GetBaseFee(ctx, ethCfg)
	// skip check as the London hard fork and EIP-1559 are enabled
	if baseFee != nil {
		// FIXME: skip to the next sub handler
		return next(ctx, tx, simulate)
	}

	if isLondon && baseFee == nil {
		return ctx, errorsmod.Wrap(
			evmtypes.ErrInvalidBaseFee,
			"base fee is supported but evm block context value is nil",
		)
	}

	txFee := sdk.Coins{}
	txGasLimit := uint64(0)
	gasWanted := uint64(0)
	evmDenom := evmParams.EvmDenom
	// TODO: use AmountOfNoValidation instead
	mempoolMinGasPrice := ctx.MinGasPrices().AmountOf(evmDenom)
	globalMinGasPrice := md.feeMarketKeeper.GetParams(ctx).MinGasPrice
	blockTxIndex := md.evmKeeper.GetTxIndexTransient(ctx)

	// Use the lowest priority of all the messages as the final one.
	minPriority := int64(math.MaxInt64)

	for i, msg := range tx.GetMsgs() {
		ethMsg, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		feeAmt := txData.Fee()
		gas := txData.GetGas()

		fee := sdk.NewDecFromBigInt(feeAmt)
		gasLimit := sdk.NewDecFromBigInt(new(big.Int).SetUint64(gas))
		requiredMempoolFee := mempoolMinGasPrice.Mul(gasLimit)
		requiredGlobalFee := globalMinGasPrice.Mul(gasLimit)

		// 2. mempool inclusion fee
		if ctx.IsCheckTx() && !simulate && fee.LT(requiredMempoolFee) {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrInsufficientFee,
				"insufficient mempool inclusion fee; got: %s required: %s",
				fee.TruncateInt().String(), requiredMempoolFee.TruncateInt().String(),
			)
		}

		// 3. min gas price (global min fee)

		if txData.TxType() != ethtypes.LegacyTxType {
			feeAmt = txData.EffectiveFee(baseFee)
			fee = sdk.NewDecFromBigInt(feeAmt)
		}

		if requiredGlobalFee.IsPositive() && fee.LT(requiredGlobalFee) {
			return ctx, errorsmod.Wrapf(
				errortypes.ErrInsufficientFee,
				"provided fee < minimum global fee (%s < %s). Please increase the priority tip (for EIP-1559 txs) or the gas prices (for access list or legacy txs)", //nolint:lll
				fee.TruncateInt().String(), requiredGlobalFee.TruncateInt().String(),
			)
		}

		// 4. validate basic
		txFee, txGasLimit, err = TxFee(
			txData.GetTo(),
			from,
			txGasLimit,
			gas,
			evmParams.EnableCreate,
			evmParams.EnableCall,
			baseFee,
			feeAmt,
			txData.TxType(),
			evmDenom,
			txFee,
		)
		if err != nil {
			return ctx, err
		}

		// 5. signature verification
		if err := SignatureVerification(ethMsg, signer, allowUnprotectedTxs); err != nil {
			return ctx, err
		}

		// NOTE: sender address has been verified and cached
		from = ethMsg.GetFrom()

		// 6. account balance verification
		fromAddr := common.HexToAddress(ethMsg.From)
		// // TODO: Use account from AccountKeeper instead
		account := md.evmKeeper.GetAccount(ctx, fromAddr)
		if err := VerifyAccountBalance(ctx, md.accountKeeper, account, fromAddr, txData); err != nil {
			return ctx, err
		}

		// 7. can transfer
		coreMsg, err := ethMsg.AsMessage(signer, baseFee)
		if err != nil {
			return ctx, errorsmod.Wrapf(
				err,
				"failed to create an ethereum core.Message from signer %T", signer,
			)
		}

		if err := CanTransfer(ctx, md.evmKeeper, coreMsg, baseFee, ethCfg, evmParams, isLondon); err != nil {
			return ctx, err
		}

		// 8. vesting
		value := txData.GetValue()
		acc := md.accountKeeper.GetAccount(ctx, from)
		if acc == nil {
			// safety check: shouldn't happen
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownAddress,
				"account %s does not exist", acc)
		}

		if err := CheckVesting(ctx, md.bankKeeper, acc, accountExpenses, value, evmDenom); err != nil {
			return ctx, err
		}

		// 9. gas consumption

		gasWanted, minPriority, err = ConsumeGas(
			ctx,
			md.bankKeeper,
			md.distributionKeeper,
			md.evmKeeper,
			md.stakingKeeper,
			from,
			txData,
			minPriority,
			gasWanted,
			md.maxGasWanted,
			evmDenom,
			baseFee,
			isHomestead,
			isIstanbul,
		)

		if err != nil {
			return ctx, err
		}

		// 10. increment sequence
		if err := IncrementNonce(ctx, md.accountKeeper, acc, txData.GetNonce()); err != nil {
			return ctx, err
		}

		// 11. gas wanted
		if err := CheckGasWanted(ctx, md.feeMarketKeeper, tx, isLondon); err != nil {
			return ctx, err
		}

		// 12. emit events
		txIdx := uint64(i) // nosec: G701
		EmitTxHashEvent(ctx, ethMsg, blockTxIndex, txIdx)
	}

	if err := CheckTxFee(txFeeInfo, txFee, txGasLimit); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
