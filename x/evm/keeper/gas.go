// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/evmos/evmos/v16/x/evm/types"
)

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, accessList ethtypes.AccessList, isContractCreation bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if isContractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		nonZeroGas := params.TxDataNonZeroGasEIP2028

		if (math.MaxUint64-gas)/nonZeroGas < nz {
			return 0, core.ErrGasUintOverflow
		}
		gas += nz * nonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, core.ErrGasUintOverflow
		}
		gas += z * params.TxDataZeroGas
	}
	if accessList != nil {
		gas += uint64(len(accessList)) * params.TxAccessListAddressGas
		gas += uint64(accessList.StorageKeys()) * params.TxAccessListStorageKeyGas
	}
	return gas, nil
}

// RefundGas transfers the leftover gas to the sender of the message, caped to half of the total gas
// consumed in the transaction. Additionally, the function sets the total gas consumed to the value
// returned by the EVM execution, thus ignoring the previous intrinsic gas consumed during in the
// AnteHandler.
func (k *Keeper) RefundGas(ctx sdk.Context, msg core.Message, leftoverGas uint64, denom string) error {
	// Return EVM tokens for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(leftoverGas), msg.GasPrice())

	switch remaining.Sign() {
	case -1:
		// negative refund errors
		return errorsmod.Wrapf(types.ErrInvalidRefund, "refunded amount value cannot be negative %d", remaining.Int64())
	case 1:
		// positive amount refund
		refundedCoins := sdk.Coins{sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(remaining))}

		// refund to sender from the fee collector module account, which is the escrow account in charge of collecting tx fees

		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, msg.From().Bytes(), refundedCoins)
		if err != nil {
			err = errorsmod.Wrapf(errortypes.ErrInsufficientFunds, "fee collector account failed to refund fees: %s", err.Error())
			return errorsmod.Wrapf(err, "failed to refund %d leftover gas (%s)", leftoverGas, refundedCoins.String())
		}
	default:
		// no refund, consume gas and update the tx gas meter
	}

	return nil
}

// ResetGasMeterAndConsumeGas reset first the gas meter consumed value to zero and set it back to the new value
// 'gasUsed'
func (k *Keeper) ResetGasMeterAndConsumeGas(ctx sdk.Context, gasUsed uint64) {
	// reset the gas count
	ctx.GasMeter().RefundGas(ctx.GasMeter().GasConsumed(), "reset the gas count")
	ctx.GasMeter().ConsumeGas(gasUsed, "apply evm transaction")
}

// GasToRefund calculates the amount of gas the state machine should refund to the sender. It is
// capped by the refund quotient value.
// Note: do not pass 0 to refundQuotient
func GasToRefund(availableRefund, gasConsumed, refundQuotient uint64) uint64 {
	// Apply refund counter
	refund := gasConsumed / refundQuotient
	if refund > availableRefund {
		return availableRefund
	}
	return refund
}
