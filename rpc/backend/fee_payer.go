package backend

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

func (b *Backend) feePayerTx(clientCtx client.Context, ethereumMsg *evmtypes.MsgEthereumTx, evmDenom string) (tx authsigning.Tx, err error) {
	if b.feePayerPrivKey == nil {
		panic("no fee payer priv key")
	}
	privKey := *b.feePayerPrivKey
	pubKey := privKey.PubKey()

	// Add the extension options to the transaction for the ethereum message
	txBuilder, ok := clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		err = fmt.Errorf("unsupported builder: %T", b)
		return
	}
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		return
	}
	txBuilder.SetExtensionOptions(option)

	// Set fees from the ethereum message
	txData, err := evmtypes.UnpackTxData(ethereumMsg.Data)
	if err != nil {
		return
	}
	fees := make(sdk.Coins, 0, 1)
	feeAmt := sdkmath.NewIntFromBigInt(txData.Fee())
	if feeAmt.Sign() > 0 {
		fees = append(fees, sdk.NewCoin(evmDenom, feeAmt))
	}
	txBuilder.SetFeeAmount(fees)

	// Set gas limit from the ethereum message
	txBuilder.SetGasLimit(ethereumMsg.GetGas())

	// A valid msg should have empty From field
	ethereumMsg.From = ""

	// Set message in the transaction
	err = txBuilder.SetMsgs(ethereumMsg)
	if err != nil {
		return
	}

	// Add the fee payer information
	feepayerAddress := sdk.AccAddress(pubKey.Address())
	txBuilder.SetFeePayer(feepayerAddress)

	// Query the account number and sequence from the remote chain
	accountRetriever := authtypes.AccountRetriever{}
	accountNumber, sequence, err := accountRetriever.GetAccountNumberSequence(clientCtx, feepayerAddress)
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	// Make sure AuthInfo is complete before signing
	sigData := signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sigV2 := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: sequence,
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return
	}

	// Sign and set signatures
	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
	sig, err := clienttx.SignWithPrivKey(
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder,
		&privKey,
		clientCtx.TxConfig,
		sequence,
	)
	if err != nil {
		err = fmt.Errorf("failed to sign transaction: %w", err)
		return
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		err = fmt.Errorf("failed to set signatures: %w", err)
		return
	}

	tx = txBuilder.GetTx()
	return
}
