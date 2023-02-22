package tx

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/evmos/evmos/v11/app"
	cryptocodec "github.com/evmos/evmos/v11/crypto/codec"
	"github.com/evmos/evmos/v11/ethereum/eip712"
	"github.com/evmos/evmos/v11/types"
)

// CreateEIP712CosmosTx creates a cosmos tx for typed data according to EIP712.
// Also, signs the tx with the provided messages and private key.
// It returns the signed transaction and an error
func CreateEIP712CosmosTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	input CosmosTxInput,
) (sdk.Tx, error) {
	builder, err := PrepareEIP712CosmosTx(
		ctx,
		appEvmos,
		input,
	)
	return builder.GetTx(), err
}

// PrepareEIP712CosmosTx creates a cosmos tx for typed data according to EIP712.
// Also, signs the tx with the provided messages and private key.
// It returns the tx builder with the signed transaction and an error
func PrepareEIP712CosmosTx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	input CosmosTxInput,
) (client.TxBuilder, error) {
	pc, err := types.ParseChainID(input.ChainID)
	if err != nil {
		return nil, err
	}
	chainIDNum := pc.Uint64()

	from := sdk.AccAddress(input.Priv.PubKey().Address().Bytes())
	accNumber := appEvmos.AccountKeeper.GetAccount(ctx, from).GetAccountNumber()

	nonce, err := appEvmos.AccountKeeper.GetSequence(ctx, from)
	if err != nil {
		return nil, err
	}

	// GenerateTypedData TypedData
	var evmosCodec codec.ProtoCodecMarshaler
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	evmosCodec = codec.NewProtoCodec(registry)
	cryptocodec.RegisterInterfaces(registry)

	fee := legacytx.NewStdFee(input.Gas, input.Fees) //nolint: staticcheck

	data := legacytx.StdSignBytes(ctx.ChainID(), accNumber, nonce, 0, fee, input.Msgs, "", nil)
	typedData, err := eip712.WrapTxToTypedData(evmosCodec, chainIDNum, input.Msgs[0], data, &eip712.FeeDelegationOptions{
		FeePayer: from,
	})
	if err != nil {
		return nil, err
	}

	txBuilder := input.TxCfg.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errors.New("txBuilder could not be casted to authtx.ExtensionOptionsTxBuilder type")
	}

	builder.SetFeeAmount(fee.Amount)
	builder.SetGasLimit(input.Gas)

	err = builder.SetMsgs(input.Msgs...)
	if err != nil {
		return nil, err
	}

	return signCosmosEIP712Tx(
		ctx,
		appEvmos,
		builder,
		input.Priv,
		chainIDNum, typedData,
	)
}

// signCosmosEIP712Tx signs the cosmos transaction on the txBuilder provided using
// the provided private key and the typed data
func signCosmosEIP712Tx(
	ctx sdk.Context,
	appEvmos *app.Evmos,
	builder authtx.ExtensionOptionsTxBuilder,
	priv cryptotypes.PrivKey,
	chainID uint64,
	data apitypes.TypedData,
) (client.TxBuilder, error) {
	from := sdk.AccAddress(priv.PubKey().Address().Bytes())
	nonce, err := appEvmos.AccountKeeper.GetSequence(ctx, from)
	if err != nil {
		return nil, err
	}

	sigHash, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return nil, err
	}

	// Sign typedData
	keyringSigner := NewSigner(priv)
	signature, pubKey, err := keyringSigner.SignByAddress(from, sigHash)
	if err != nil {
		return nil, err
	}
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	// Add ExtensionOptionsWeb3Tx extension
	option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionsWeb3Tx{
		FeePayer:         from.String(),
		TypedDataChainID: chainID,
		FeePayerSig:      signature,
	})
	if err != nil {
		return nil, err
	}

	builder.SetExtensionOptions(option)

	sigsV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		Sequence: nonce,
	}

	err = builder.SetSignatures(sigsV2)
	if err != nil {
		return nil, err
	}

	return builder, nil
}
