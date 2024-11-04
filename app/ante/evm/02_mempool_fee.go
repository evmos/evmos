// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

// CheckMempoolFee checks if the provided fee is at least as large as the local
// validator's configured value. The fee computation assumes that both price and fee are
// represented in 18 decimals.
func CheckMempoolFee(fee, mempoolMinGasPrice, gasLimit sdkmath.LegacyDec, isLondon bool) error {
	if isLondon {
		return nil
	}

	requiredFee := mempoolMinGasPrice.Mul(gasLimit)

	if fee.LT(requiredFee) {
		return errorsmod.Wrapf(
			errortypes.ErrInsufficientFee,
			"got: %s, minimum required: %s",
			fee, requiredFee,
		)
	}

	return nil
}
