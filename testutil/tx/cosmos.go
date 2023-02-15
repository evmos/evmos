package tx

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/utils"
)

var DefaultTxFee = sdk.NewCoin(utils.BaseDenom, sdk.NewInt(10_000_000_000_000_000)) // 0.01 EVMOS

func PrepareCosmosTx(
	ctx sdk.Context,
	txCfg client.TxConfig,
	appEvmos *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (authsigning.Tx, error) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	txBuilder := txCfg.NewTxBuilder()

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
		return nil, err
	}

	seq, err := appEvmos.AccountKeeper.GetSequence(ctx, accountAddress)
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  txCfg.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := appEvmos.AccountKeeper.GetAccount(ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		txCfg.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, txCfg,
		seq,
	)
	if err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	if err = txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}
	return txBuilder.GetTx(), nil
}
