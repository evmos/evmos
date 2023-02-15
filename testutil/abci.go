package testutil

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/testutil/tx"
)

// Commit commits a block at a given time. Reminder: At the end of each
// Tendermint Consensus round the following methods are run
//  1. BeginBlock
//  2. DeliverTx
//  3. EndBlock
//  4. Commit
func Commit(ctx sdk.Context, app *app.Evmos, t time.Duration) sdk.Context {
	header := ctx.BlockHeader()
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: header.Height})
	_ = app.Commit()

	header.Height++
	header.Time = header.Time.Add(t)
	app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	return app.BaseApp.NewContext(false, header)
}

// DeliverTx delivers a cosmos tx for a given set of msgs
func DeliverTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (abci.ResponseDeliverTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
	tx, err := tx.PrepareCosmosTx(ctx, txConfig, appEvmos, priv, gasPrice, msgs...)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}
	return BroadcastTxBytes(appEvmos, txConfig.TxEncoder(), tx)
}

// DeliverEthTx generates and broadcasts a Cosmos Tx populated with MsgEthereumTx messages.
// If a private key is provided, it will attempt to sign all messages with the given private key,
// otherwise, it will assume the messages have already been signed.
func DeliverEthTx(
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	msgs ...sdk.Msg,
) (abci.ResponseDeliverTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig

	tx, err := prepareEthTx(txConfig, appEvmos, priv, msgs...)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}
	return BroadcastTxBytes(appEvmos, txConfig.TxEncoder(), tx)
}

// CheckTx checks a cosmos tx for a given set of msgs
func CheckTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (abci.ResponseCheckTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig

	tx, err := tx.PrepareCosmosTx(ctx, txConfig, appEvmos, priv, gasPrice, msgs...)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}
	bz, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}
	req := abci.RequestCheckTx{Tx: bz}

	res := appEvmos.CheckTx(req)
	if res.Code != 0 {
		return abci.ResponseCheckTx{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, res.Log)
	}
	return res, nil
}

// CheckEthTx checks a Ethereum tx for a given set of msgs
func CheckEthTx(
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	msgs ...sdk.Msg,
) (abci.ResponseCheckTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig

	tx, err := prepareEthTx(txConfig, appEvmos, priv, msgs...)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}
	bz, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}
	req := abci.RequestCheckTx{Tx: bz}

	res := appEvmos.CheckTx(req)
	if res.Code != 0 {
		return abci.ResponseCheckTx{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, res.Log)
	}
	return res, nil
}

// BroadcastTxBytes encodes a transaction and calls DeliverTx on the app.
func BroadcastTxBytes(app *app.Evmos, txEncoder sdk.TxEncoder, tx sdk.Tx) (abci.ResponseDeliverTx, error) {
	// bz are bytes to be broadcasted over the network
	bz, err := txEncoder(tx)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	req := abci.RequestDeliverTx{Tx: bz}
	res := app.BaseApp.DeliverTx(req)
	if res.Code != 0 {
		return abci.ResponseDeliverTx{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, res.Log)
	}

	return res, nil
}
