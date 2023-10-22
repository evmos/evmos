// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/encoding"
)

// NextFn is a no-op function that returns the context and no error in order to mock
// the next function in the AnteHandler chain.
//
// It can be used in unit tests when calling a decorator's AnteHandle method, e.g.
// `dec.AnteHandle(ctx, tx, false, NextFn)`
func NextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

// ValidateAnteForMsgs is a helper function, which takes in an AnteDecorator as well as 1 or
// more messages, builds a transaction containing these messages, and returns any error that
// the AnteHandler might return.
func ValidateAnteForMsgs(ctx sdk.Context, dec sdk.AnteDecorator, msgs ...sdk.Msg) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return err
	}

	tx := txBuilder.GetTx()

	// Call Ante decorator
	_, err = dec.AnteHandle(ctx, tx, false, NextFn)
	return err
}
