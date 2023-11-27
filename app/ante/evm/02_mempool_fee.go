// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

// CheckMempoolFee ensures that the provided fees meet a minimum threshold for the validator.
// This check only for local mempool purposes, and thus it is only run on (Re)CheckTx.
// The logic is also skipped if the London hard fork and EIP-1559 are enabled.
// The function checks checks if the provided fee is at least as large as the local validator's
func CheckMempoolFee(fee, mempoolMinGasPrice, gasLimit sdkmath.LegacyDec, isLondon bool) error {
	if isLondon {
		return nil
	}

	requiredFee := mempoolMinGasPrice.Mul(gasLimit)

	if fee.LT(requiredFee) {
		return errorsmod.Wrapf(
			errortypes.ErrInsufficientFee,
			"insufficient fee; got: %s required: %s",
			fee, requiredFee,
		)
	}

	return nil
}
