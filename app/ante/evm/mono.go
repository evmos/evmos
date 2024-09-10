// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	anteutils "github.com/evmos/evmos/v19/app/ante/utils"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
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

type DecoratorUtils struct {
	EvmParams          evmtypes.Params
	EthConfig          *params.ChainConfig
	Rules              params.Rules
	Signer             ethtypes.Signer
	BaseFee            *big.Int
	EvmDenom           string
	MempoolMinGasPrice sdkmath.LegacyDec
	GlobalMinGasPrice  sdkmath.LegacyDec
	BlockTxIndex       uint64
	TxGasLimit         uint64
	GasWanted          uint64
	MinPriority        int64
	TxFee              sdk.Coins
}

// NewMonoDecorator creates a new MonoDecorator
func NewMonoDecorator(
	accountKeeper evmtypes.AccountKeeper,
	bankKeeper evmtypes.BankKeeper,
	feeMarketKeeper FeeMarketKeeper,
	evmKeeper EVMKeeper,
	distributionKeeper anteutils.DistributionKeeper,
	stakingKeeper anteutils.StakingKeeper,
	maxGasWanted uint64,
) MonoDecorator {
	return MonoDecorator{
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		feeMarketKeeper:    feeMarketKeeper,
		evmKeeper:          evmKeeper,
		distributionKeeper: distributionKeeper,
		stakingKeeper:      stakingKeeper,
		maxGasWanted:       maxGasWanted,
	}
}

// NewMonoDecoratorUtils returns a new DecoratorUtils instance.
//
// These utilities are extracted once at the beginning of the ante handle process,
// and are used throughout the entire decorator chain.
// This avoids redundant calls to the keeper and thus improves speed of transaction processing.
func NewMonoDecoratorUtils(
	ctx sdk.Context,
	ek EVMKeeper,
	fmk FeeMarketKeeper,
) (*DecoratorUtils, error) {
	evmParams := ek.GetParams(ctx)
	chainCfg := evmParams.GetChainConfig()
	ethCfg := chainCfg.EthereumConfig(ek.ChainID())
	blockHeight := big.NewInt(ctx.BlockHeight())
	rules := ethCfg.Rules(blockHeight, true)
	baseFee := ek.GetBaseFee(ctx, ethCfg)
	feeMarketParams := fmk.GetParams(ctx)

	if rules.IsLondon && baseFee == nil {
		return nil, errorsmod.Wrap(
			evmtypes.ErrInvalidBaseFee,
			"base fee is supported but evm block context value is nil",
		)
	}

	return &DecoratorUtils{
		EvmParams:          evmParams,
		EthConfig:          ethCfg,
		Rules:              rules,
		Signer:             ethtypes.MakeSigner(ethCfg, blockHeight),
		BaseFee:            baseFee,
		MempoolMinGasPrice: ctx.MinGasPrices().AmountOf(evmParams.EvmDenom),
		GlobalMinGasPrice:  feeMarketParams.MinGasPrice,
		EvmDenom:           evmParams.EvmDenom,
		BlockTxIndex:       ek.GetTxIndexTransient(ctx),
		TxGasLimit:         0,
		GasWanted:          0,
		MinPriority:        int64(math.MaxInt64),
		TxFee:              sdk.Coins{},
	}, nil
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

	// 2. get utils
	decUtils, err := NewMonoDecoratorUtils(ctx, md.evmKeeper, md.feeMarketKeeper)
	if err != nil {
		return ctx, err
	}

	msgs := tx.GetMsgs()
	if msgs == nil {
		return ctx, errorsmod.Wrap(errortypes.ErrUnknownRequest, "invalid transaction. Transaction without messages")
	}

	// Use the lowest priority of all the messages as the final one.
	for i, msg := range msgs {
		ethMsg, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		feeAmt := txData.Fee()
		gas := txData.GetGas()
		fee := sdkmath.LegacyNewDecFromBigInt(feeAmt)
		gasLimit := sdkmath.LegacyNewDecFromBigInt(new(big.Int).SetUint64(gas))

		// 2. mempool inclusion fee
		if ctx.IsCheckTx() && !simulate {
			if err := CheckMempoolFee(fee, decUtils.MempoolMinGasPrice, gasLimit, decUtils.Rules.IsLondon); err != nil {
				return ctx, err
			}
		}

		// 3. min gas price (global min fee)
		if txData.TxType() == ethtypes.DynamicFeeTxType && decUtils.BaseFee != nil {
			feeAmt = txData.EffectiveFee(decUtils.BaseFee)
			fee = sdkmath.LegacyNewDecFromBigInt(feeAmt)
		}

		if err := CheckGlobalFee(fee, decUtils.GlobalMinGasPrice, gasLimit); err != nil {
			return ctx, err
		}

		// 4. validate msg contents
		err = ValidateMsg(
			decUtils.EvmParams,
			txData,
			from,
		)
		if err != nil {
			return ctx, err
		}

		// 5. signature verification
		if err := SignatureVerification(
			ethMsg,
			decUtils.Signer,
			decUtils.EvmParams.AllowUnprotectedTxs,
		); err != nil {
			return ctx, err
		}

		// NOTE: sender address has been verified and cached
		from = ethMsg.GetFrom()

		// 6. account balance verification
		fromAddr := common.HexToAddress(ethMsg.From)
		// TODO: Use account from AccountKeeper instead
		account := md.evmKeeper.GetAccount(ctx, fromAddr)
		if err := VerifyAccountBalance(
			ctx,
			md.accountKeeper,
			account,
			fromAddr,
			txData,
		); err != nil {
			return ctx, err
		}

		// 7. can transfer
		coreMsg, err := ethMsg.AsMessage(decUtils.Signer, decUtils.BaseFee)
		if err != nil {
			return ctx, errorsmod.Wrapf(
				err,
				"failed to create an ethereum core.Message from signer %T", decUtils.Signer,
			)
		}

		if err := CanTransfer(
			ctx,
			md.evmKeeper,
			coreMsg,
			decUtils.BaseFee,
			decUtils.EthConfig,
			decUtils.EvmParams,
			decUtils.Rules.IsLondon,
		); err != nil {
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

		if err := CheckVesting(
			ctx,
			md.bankKeeper,
			acc,
			accountExpenses,
			value,
			decUtils.EvmDenom,
		); err != nil {
			return ctx, err
		}

		// 9. gas consumption
		msgFees, err := evmkeeper.VerifyFee(
			txData,
			decUtils.EvmDenom,
			decUtils.BaseFee,
			decUtils.Rules.IsHomestead,
			decUtils.Rules.IsIstanbul,
			ctx.IsCheckTx(),
		)
		if err != nil {
			return ctx, err
		}

		err = ConsumeFeesAndEmitEvent(
			ctx,
			&ConsumeGasKeepers{
				Bank:         md.bankKeeper,
				Distribution: md.distributionKeeper,
				Evm:          md.evmKeeper,
				Staking:      md.stakingKeeper,
			},
			msgFees,
			from,
		)
		if err != nil {
			return ctx, err
		}

		gasWanted := UpdateCumulativeGasWanted(
			ctx,
			txData.GetGas(),
			md.maxGasWanted,
			decUtils.GasWanted,
		)
		decUtils.GasWanted = gasWanted

		minPriority := GetMsgPriority(
			txData,
			decUtils.MinPriority,
			decUtils.BaseFee,
		)
		decUtils.MinPriority = minPriority

		txFee := UpdateCumulativeTxFee(
			decUtils.TxFee,
			txData.Fee(),
			decUtils.EvmDenom,
		)
		decUtils.TxFee = txFee
		decUtils.TxGasLimit += gas

		// 10. increment sequence
		if err := IncrementNonce(ctx, md.accountKeeper, acc, txData.GetNonce()); err != nil {
			return ctx, err
		}

		// 11. gas wanted
		if err := CheckGasWanted(ctx, md.feeMarketKeeper, tx, decUtils.Rules.IsLondon); err != nil {
			return ctx, err
		}

		// 12. emit events
		txIdx := uint64(i) //nolint:gosec // G115 G701
		EmitTxHashEvent(ctx, ethMsg, decUtils.BlockTxIndex, txIdx)
	}

	if err := CheckTxFee(txFeeInfo, decUtils.TxFee, decUtils.TxGasLimit); err != nil {
		return ctx, err
	}

	ctx, err = CheckBlockGasLimit(ctx, decUtils.GasWanted, decUtils.MinPriority)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
