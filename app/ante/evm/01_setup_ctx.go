// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	evmante "github.com/evmos/evmos/v20/x/evm/ante"
)

var _ sdk.AnteDecorator = &EthSetupContextDecorator{}

// EthSetupContextDecorator is adapted from SetUpContextDecorator from cosmos-sdk, it ignores gas consumption
// by setting the gas meter to infinite
type EthSetupContextDecorator struct {
	evmKeeper EVMKeeper
}

func NewEthSetUpContextDecorator(evmKeeper EVMKeeper) EthSetupContextDecorator {
	return EthSetupContextDecorator{
		evmKeeper: evmKeeper,
	}
}

func (esc EthSetupContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	newCtx, err = SetupContextAndResetTransientGas(ctx, tx, esc.evmKeeper)
	if err != nil {
		return ctx, err
	}
	return next(newCtx, tx, simulate)
}

// SetupContextAndResetTransientGas modifies the context to be used in the
// execution of the ante handler associated with an EVM transaction. Previous
// gas consumed is reset in the transient store.
func SetupContextAndResetTransientGas(ctx sdk.Context, tx sdk.Tx, evmKeeper EVMKeeper) (sdk.Context, error) {
	// all transactions must implement GasTx
	_, ok := tx.(authante.GasTx)
	if !ok {
		return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType, "invalid transaction type %T, expected GasTx", tx)
	}

	// To have gas consumption consistent with Ethereum, we need to:
	//     1. Set an empty gas config for both KV and transient store.
	//     2. Set an infinite gas meter.
	newCtx := evmante.BuildEvmExecutionCtx(ctx).
		WithGasMeter(storetypes.NewInfiniteGasMeter())

	// Reset transient gas used to prepare the execution of current cosmos tx.
	// Transient gas-used is necessary to sum the gas-used of cosmos tx, when it contains multiple eth msgs.
	// TODO: add more context here to explain why gas used is reset. Not clear
	// from docstring.
	evmKeeper.ResetTransientGasUsed(ctx)

	return newCtx, nil
}
