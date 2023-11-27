// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// SignatureVerification checks that the registered chain id is the same as the one on the message, and
// that the signer address matches the one defined on the message.
// It's not skipped for RecheckTx, because it set `From` address which is critical from other ante handler to work.
// Failure in RecheckTx will prevent tx to be included into block, especially when CheckTx succeed, in which case user
// won't see the error message.
func SignatureVerification(
	msg *evmtypes.MsgEthereumTx,
	signer ethtypes.Signer,
	allowUnprotectedTxs bool,
) error {
	ethTx := msg.AsTransaction()

	if !allowUnprotectedTxs && !ethTx.Protected() {
		return errorsmod.Wrapf(
			errortypes.ErrNotSupported,
			"rejected unprotected Ethereum transaction. Please EIP155 sign your transaction to protect it against replay-attacks")
	}

	sender, err := signer.Sender(ethTx)
	if err != nil {
		return errorsmod.Wrapf(
			errortypes.ErrorInvalidSigner,
			"couldn't retrieve sender address from the ethereum transaction: %s",
			err.Error(),
		)
	}

	// set up the sender to the transaction field if not already
	msg.From = sender.Hex()
	return nil
}
