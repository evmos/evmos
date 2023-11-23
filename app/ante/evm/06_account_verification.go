// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/x/evm/keeper"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// EthAccountVerificationDecorator validates an account balance checks
type EthAccountVerificationDecorator struct {
	ak        evmtypes.AccountKeeper
	evmKeeper EVMKeeper
}

// NewEthAccountVerificationDecorator creates a new EthAccountVerificationDecorator
func NewEthAccountVerificationDecorator(ak evmtypes.AccountKeeper, ek EVMKeeper) EthAccountVerificationDecorator {
	return EthAccountVerificationDecorator{
		ak:        ak,
		evmKeeper: ek,
	}
}

// AnteHandle validates checks that the sender balance is greater than the total transaction cost.
// The account will be set to store if it doesn't exist, i.e. cannot be found on store.
// This AnteHandler decorator will fail if:
// - any of the msgs is not a MsgEthereumTx
// - from address is empty
// - account balance is lower than the transaction cost
func (avd EthAccountVerificationDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		_, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		fromAddr := common.BytesToAddress(from)
		account := avd.evmKeeper.GetAccount(ctx, fromAddr)
		if err := VerifyAccountBalance(ctx, avd.ak, account, fromAddr, txData); err != nil {
			return ctx, err
		}
	}
	return next(ctx, tx, simulate)
}

// VerifyAccountBalance checks that the sender balance is greater than the total transaction cost.
func VerifyAccountBalance(
	ctx sdk.Context,
	accountKeeper evmtypes.AccountKeeper,
	account *statedb.Account,
	from common.Address,
	txData evmtypes.TxData,
) error {
	// check whether the sender address is EOA
	if account != nil && account.IsContract() {
		return errorsmod.Wrapf(
			errortypes.ErrInvalidType,
			"the sender is not EOA: address %s", from,
		)
	}

	if account == nil {
		acc := accountKeeper.NewAccountWithAddress(ctx, from.Bytes())
		accountKeeper.SetAccount(ctx, acc)
		account = statedb.NewEmptyAccount()
	}

	if err := keeper.CheckSenderBalance(sdkmath.NewIntFromBigInt(account.Balance), txData); err != nil {
		return errorsmod.Wrap(err, "failed to check sender balance")
	}

	return nil
}
