// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"errors"
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	evmostypes "github.com/evmos/evmos/v18/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// ValidateMsg validates an Ethereum specific message type and returns an error if invalid
//
// It checks the following requirements:
// - nil MUST be passed as the from address
func ValidateMsg(
	evmParams evmtypes.Params,
	txData evmtypes.TxData,
	from sdktypes.AccAddress,
) error {
	if from != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid from address; expected nil; got: %q", from.String())
	}

	return nil
}

// Validate Permission
// - If the transaction is a contract creation
//   - Check the permission type
//   - if is permissioned, checks that the address is included in whitelist
//
// - If the transaction is a contract call
//   - if the recipient is not a contract it allows it
//   - if its a contract it:
//   - Checks the permission type
//   - if is permissioned, checks that the address is included in whitelist
func ValidatePermission(
	ctx sdk.Context,
	txData evmtypes.TxData,
	accountKeeper evmtypes.AccountKeeper,
	evmParams evmtypes.Params,
	from common.Address,
) error {
	permissions := &evmParams.AccessControl
	contractDeploy := txData.GetTo() == nil

	addr := common.Address(from.Bytes())
	if contractDeploy {
		return validateCreate(permissions, addr)
	}

	data := txData.GetData()
	toAccount := accountKeeper.GetAccount(ctx, txData.GetTo().Bytes())
	ethAcct, ok := toAccount.(evmostypes.EthAccountI)
	emptyCodeHash := crypto.Keccak256Hash(nil)
	// If its not a contract creation or contract call this check is irrelevant
	if data == nil && ok && (ethAcct.GetCodeHash() == emptyCodeHash) {
		return nil
	}
	return validateCall(permissions, addr)
}

func validateCreate(
	permissions *evmtypes.AccessControl,
	sender common.Address,
) error {
	switch permissions.Create.AccessType {
	case evmtypes.AccessTypePermissionless:
		return nil
	case evmtypes.AccessTypeRestricted:
		return errorsmod.Wrap(evmtypes.ErrCreateDisabled, "failed to create new contract")
	case evmtypes.AccessTypePermissioned:
		if len(permissions.Create.WhitelistAddresses) >= 1 && slices.Contains(permissions.Create.WhitelistAddresses, sender.Hex()) {
			return nil
		}
		return errorsmod.Wrap(evmtypes.ErrCreateDisabled, "does not have permission to create new contract")
	}
	return errors.New("undefined access type")
}

func validateCall(
	permissions *evmtypes.AccessControl,
	sender common.Address,
) error {
	switch permissions.Call.AccessType {
	case evmtypes.AccessTypePermissionless:
		return nil
	case evmtypes.AccessTypeRestricted:
		return errorsmod.Wrap(evmtypes.ErrCallDisabled, "failed to perform a call")
	case evmtypes.AccessTypePermissioned:
		if len(permissions.Call.WhitelistAddresses) >= 1 && slices.Contains(permissions.Call.WhitelistAddresses, sender.Hex()) {
			return nil
		}
		return errorsmod.Wrap(evmtypes.ErrCallDisabled, "does not have permission to perform a call")
	}
	return errors.New("undefined access type")
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
