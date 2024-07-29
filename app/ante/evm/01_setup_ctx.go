// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	evmante "github.com/evmos/evmos/v19/x/evm/ante"
)

var _ sdktypes.AnteDecorator = &EthSetupContextDecorator{}

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

func (esc EthSetupContextDecorator) AnteHandle(ctx sdktypes.Context, tx sdktypes.Tx, simulate bool, next sdktypes.AnteHandler) (newCtx sdktypes.Context, err error) {
	newCtx, err = SetupContext(ctx, tx, esc.evmKeeper)
	if err != nil {
		return ctx, err
	}
	return next(newCtx, tx, simulate)
}

func SetupContext(ctx sdktypes.Context, tx sdktypes.Tx, evmKeeper EVMKeeper) (sdktypes.Context, error) {
	// all transactions must implement GasTx
	_, ok := tx.(authante.GasTx)
	if !ok {
		return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType, "invalid transaction type %T, expected GasTx", tx)
	}

	// We need to set up an empty gas config so that the gas is consistent with Ethereum.
	newCtx := evmante.BuildEvmExecutionCtx(ctx).
		WithGasMeter(sdktypes.NewInfiniteGasMeter())
	// Reset transient gas used to prepare the execution of current cosmos tx.
	// Transient gas-used is necessary to sum the gas-used of cosmos tx, when it contains multiple eth msgs.
	evmKeeper.ResetTransientGasUsed(ctx)

	return newCtx, nil
}
