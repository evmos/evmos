// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"errors"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// EthValidateBasicDecorator is adapted from ValidateBasicDecorator from cosmos-sdk, it ignores ErrNoSignatures
type EthValidateBasicDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthValidateBasicDecorator creates a new EthValidateBasicDecorator
func NewEthValidateBasicDecorator(ek EVMKeeper) EthValidateBasicDecorator {
	return EthValidateBasicDecorator{
		evmKeeper: ek,
	}
}

// AnteHandle handles basic validation of tx
func (vbd EthValidateBasicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to validate basic on recheck tx, call next antehandler
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	txFeeInfo, err := ValidateTx(tx)
	if err != nil {
		return ctx, err
	}

	txFee := sdk.Coins{}
	txGasLimit := uint64(0)

	evmParams := vbd.evmKeeper.GetParams(ctx)
	chainCfg := evmParams.GetChainConfig()
	chainID := vbd.evmKeeper.ChainID()
	ethCfg := chainCfg.EthereumConfig(chainID)
	baseFee := vbd.evmKeeper.GetBaseFee(ctx, ethCfg)
	enableCreate := evmParams.GetEnableCreate()
	enableCall := evmParams.GetEnableCall()
	evmDenom := evmParams.GetEvmDenom()

	for _, msg := range tx.GetMsgs() {
		_, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		txFee, txGasLimit, err = CheckDisabledCreateCallAndUpdateTxFee(
			txData.GetTo(),
			from,
			txGasLimit,
			txData.GetGas(),
			enableCreate,
			enableCall,
			baseFee,
			txData.Fee(),
			txData.TxType(),
			evmDenom,
			txFee,
		)
		if err != nil {
			return ctx, err
		}
	}

	if err := CheckTxFee(txFeeInfo, txFee, txGasLimit); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// FIXME: split this function into multiple ones
// CheckDisabledCreateCallAndUpdateTxFee checks if contract creation or call are disabled through governance
// and updates the transaction fee by adding the message fee to the cumulative transaction fee
func CheckDisabledCreateCallAndUpdateTxFee(
	to *common.Address,
	from sdk.AccAddress,
	txGasLimit, gasLimit uint64,
	enableCreate, enableCall bool,
	baseFee *big.Int,
	msgFee *big.Int,
	txType byte,
	denom string,
	txFee sdk.Coins,
) (sdk.Coins, uint64, error) {
	// Validate `From` field
	if from != nil {
		return nil, 0, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid from address %s, expect empty string", from)
	}

	txGasLimit += gasLimit

	// return error if contract creation or call are disabled through governance
	if !enableCreate && to == nil {
		return nil, 0, errorsmod.Wrap(evmtypes.ErrCreateDisabled, "failed to create new contract")
	} else if !enableCall && to != nil {
		return nil, 0, errorsmod.Wrap(evmtypes.ErrCallDisabled, "failed to call contract")
	}

	if baseFee == nil && txType == ethtypes.DynamicFeeTxType {
		return nil, 0, errorsmod.Wrap(ethtypes.ErrTxTypeNotSupported, "dynamic fee tx not supported")
	}

	txFee = txFee.Add(sdk.Coin{Denom: denom, Amount: sdkmath.NewIntFromBigInt(msgFee)})
	return txFee, txGasLimit, nil
}

// FIXME: this shouldn't be required if the tx was an Ethereum transaction type
func ValidateTx(tx sdk.Tx) (*tx.Fee, error) {
	err := tx.ValidateBasic()
	// ErrNoSignatures is fine with eth tx
	if err != nil && !errors.Is(err, errortypes.ErrNoSignatures) {
		return nil, errorsmod.Wrap(err, "tx basic validation failed")
	}

	// For eth type cosmos tx, some fields should be verified as zero values,
	// since we will only verify the signature against the hash of the MsgEthereumTx.Data
	wrapperTx, ok := tx.(protoTxProvider)
	if !ok {
		return nil, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid tx type %T, didn't implement interface protoTxProvider", tx)
	}

	protoTx := wrapperTx.GetProtoTx()
	body := protoTx.Body
	if body.Memo != "" || body.TimeoutHeight != uint64(0) || len(body.NonCriticalExtensionOptions) > 0 {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest,
			"for eth tx body Memo TimeoutHeight NonCriticalExtensionOptions should be empty")
	}

	if len(body.ExtensionOptions) != 1 {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "for eth tx length of ExtensionOptions should be 1")
	}

	authInfo := protoTx.AuthInfo
	if len(authInfo.SignerInfos) > 0 {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "for eth tx AuthInfo SignerInfos should be empty")
	}

	if authInfo.Fee.Payer != "" || authInfo.Fee.Granter != "" {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "for eth tx AuthInfo Fee payer and granter should be empty")
	}

	sigs := protoTx.Signatures
	if len(sigs) > 0 {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "for eth tx Signatures should be empty")
	}

	return authInfo.Fee, nil
}

func CheckTxFee(txFeeInfo *tx.Fee, txFee sdk.Coins, txGasLimit uint64) error {
	if txFeeInfo == nil {
		return nil
	}

	if !txFeeInfo.Amount.IsEqual(txFee) {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid AuthInfo Fee Amount (%s != %s)", txFeeInfo.Amount, txFee)
	}

	if txFeeInfo.GasLimit != txGasLimit {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid AuthInfo Fee GasLimit (%d != %d)", txFeeInfo.GasLimit, txGasLimit)
	}

	return nil
}
