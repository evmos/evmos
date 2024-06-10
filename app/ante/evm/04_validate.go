// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// ValidateMsg validates an Ethereum specific message type and returns an error if invalid
//
// It checks the following requirements:
// - nil MUST be passed as the from address
// - If the transaction is a contract creation or call, the corresponding operation must be enabled in the EVM parameters
func ValidateMsg(
	evmParams evmtypes.Params,
	txData evmtypes.TxData,
	from sdktypes.AccAddress,
) error {
	if from != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid from address; expected nil; got: %q", from.String())
	}

	return checkDisabledCreateCall(
		txData,
		&evmParams.AccessControl,
	)
}

// checkDisabledCreateCall checks if the transaction is a contract creation or call
// and it is disabled through governance
func checkDisabledCreateCall(
	txData evmtypes.TxData,
	permissions *evmtypes.AccessControl,
) error {
	to := txData.GetTo()
	blockCreate := permissions.Create.AccessType == evmtypes.AccessTypeRestricted
	blockCall := permissions.Call.AccessType == evmtypes.AccessTypeRestricted

	// return error if contract creation or call are disabled through governance
	// and the transaction is trying to create a contract or call a contract
	if blockCreate && to == nil {
		return errorsmod.Wrap(evmtypes.ErrCreateDisabled, "failed to create new contract")
	} else if blockCall && to != nil {
		return errorsmod.Wrap(evmtypes.ErrCallDisabled, "failed to perform a call")
	}
	return nil
}

// FIXME: this shouldn't be required if the tx was an Ethereum transaction type
func ValidateTx(tx sdktypes.Tx) (*tx.Fee, error) {
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

func CheckTxFee(txFeeInfo *tx.Fee, txFee sdktypes.Coins, txGasLimit uint64) error {
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
