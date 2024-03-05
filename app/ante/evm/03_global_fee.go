// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

// For dynamic transactions, GetFee() uses the GasFeeCap value, which
// is the maximum gas price that the signer can pay. In practice, the
// signer can pay less, if the block's BaseFee is lower. So, in this case,
// we use the EffectiveFee. If the feemarket formula results in a BaseFee
// that lowers EffectivePrice until it is < MinGasPrices, the users must
// increase the GasTipCap (priority fee) until EffectivePrice > MinGasPrices.
// Transactions with MinGasPrices * gasUsed < tx fees < EffectiveFee are rejected
// by the feemarket AnteHandle
func CheckGlobalFee(fee, globalMinGasPrice, gasLimit math.LegacyDec) error {
	if globalMinGasPrice.IsZero() {
		return nil
	}

	requiredFee := globalMinGasPrice.Mul(gasLimit)

	if fee.LT(requiredFee) {
		return errorsmod.Wrapf(
			errortypes.ErrInsufficientFee,
			"provided fee < minimum global fee (%s < %s). Please increase the priority tip (for EIP-1559 txs) or the gas prices (for access list or legacy txs)", //nolint:lll
			fee.TruncateInt().String(), requiredFee.TruncateInt().String(),
		)
	}

	return nil
}
