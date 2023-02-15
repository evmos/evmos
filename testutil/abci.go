package testutil

import (
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/utils"
)

var DefaultTxFee = sdk.NewCoin(utils.BaseDenom, sdk.NewInt(10_000_000_000_000_000)) // 0.01 EVMOS

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

// DeliverTx delivers a tx for a given set of msgs
func DeliverTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (abci.ResponseDeliverTx, error) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	var gasLimit int64 = 10_000_000
	txBuilder.SetGasLimit(uint64(gasLimit))

	var fees sdk.Coins
	if gasPrice != nil {
		fees = sdk.Coins{{Denom: utils.BaseDenom, Amount: gasPrice.MulRaw(gasLimit)}}
	} else {
		fees = sdk.Coins{DefaultTxFee}
	}

	txBuilder.SetFeeAmount(fees)
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	seq, err := appEvmos.AccountKeeper.GetSequence(ctx, accountAddress)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := appEvmos.AccountKeeper.GetAccount(ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	if err = txBuilder.SetSignatures(sigsV2...); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return BroadcastTxBytes(appEvmos, encodingConfig.TxConfig.TxEncoder(), txBuilder.GetTx())
}
