package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"encoding/hex"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostxtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// PrepareEthTx signs the transaction with the provided MsgEthereumTx.
func (suite *ChainIntegrationTestSuite) PrepareEthTx(
	signer *itutiltypes.TestAccount,
	ethMsg *evmtypes.MsgEthereumTx,
) (authsigning.Tx, error) {
	suite.Require().NotNil(signer)

	txBuilder := suite.EncodingConfig.TxConfig.NewTxBuilder()

	txFee := sdk.Coins{}
	txGasLimit := uint64(0)

	// Sign messages and compute gas/fees.
	err := ethMsg.Sign(suite.EthSigner, itutiltypes.NewSigner(signer.PrivateKey))
	if err != nil {
		return nil, err
	}

	ethMsg.From = ""

	txGasLimit += ethMsg.GetGas()
	txFee = txFee.Add(sdk.Coin{Denom: suite.ChainConstantsConfig.GetMinDenom(), Amount: sdkmath.NewIntFromBigInt(ethMsg.GetFee())})

	if err := txBuilder.SetMsgs(ethMsg); err != nil {
		return nil, err
	}

	// Set the extension
	var option *codectypes.Any
	option, err = codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		return nil, err
	}

	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, errorsmod.Wrapf(errorsmod.Error{}, "could not set extensions for Ethereum tx")
	}

	builder.SetExtensionOptions(option)

	txBuilder.SetGasLimit(txGasLimit)
	txBuilder.SetFeeAmount(txFee)

	return txBuilder.GetTx(), nil
}

// CosmosTxArgs contains the params to create a cosmos tx
type CosmosTxArgs struct {
	// Gas to be used on the tx
	Gas uint64
	// GasPrice to use on tx
	GasPrice *sdkmath.Int
	// Fees is the fee to be used on the tx (amount and denom)
	Fees sdk.Coins
	// FeeGranter is the account address of the fee granter
	FeeGranter sdk.AccAddress
	// Msgs slice of messages to include on the tx
	Msgs []sdk.Msg
}

// PrepareCosmosTx creates a cosmos tx and signs it with the provided messages and private key.
// It returns the signed transaction and an error
func (suite *ChainIntegrationTestSuite) PrepareCosmosTx(
	ctx sdk.Context,
	account *itutiltypes.TestAccount,
	args CosmosTxArgs,
) (authsigning.Tx, error) {
	suite.Require().NotNil(account)

	txBuilder := suite.EncodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(args.Gas)

	var fees sdk.Coins
	if args.GasPrice != nil {
		fees = sdk.Coins{
			{
				Denom:  suite.ChainConstantsConfig.GetMinDenom(),
				Amount: args.GasPrice.MulRaw(int64(args.Gas)),
			},
		}
	} else {
		fees = sdk.Coins{
			{
				Denom:  suite.ChainConstantsConfig.GetMinDenom(),
				Amount: suite.TestConfig.DefaultFeeAmount,
			},
		}
	}

	txBuilder.SetFeeAmount(fees)
	if err := txBuilder.SetMsgs(args.Msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetFeeGranter(args.FeeGranter)

	return suite.signCosmosTx(
		ctx,
		account,
		txBuilder,
	)
}

// signCosmosTx signs the cosmos transaction on the txBuilder provided using
// the provided private key
func (suite *ChainIntegrationTestSuite) signCosmosTx(
	ctx sdk.Context,
	account *itutiltypes.TestAccount,
	txBuilder client.TxBuilder,
) (authsigning.Tx, error) {
	suite.Require().NotNil(account)

	txCfg := suite.EncodingConfig.TxConfig

	seq, err := suite.ChainApp.AccountKeeper().GetSequence(ctx, account.GetCosmosAddress())
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: account.GetPubKey(),
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
	accNumber := suite.ChainApp.AccountKeeper().GetAccount(ctx, account.GetCosmosAddress()).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       suite.ChainConstantsConfig.GetCosmosChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = clienttx.SignWithPrivKey(
		txCfg.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder, account.PrivateKey, txCfg,
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

// QueryTxResponse returns the TxResponse for the given tx
func (suite *ChainIntegrationTestSuite) QueryTxResponse(tx authsigning.Tx) *cosmostxtypes.GetTxResponse {
	var bz []byte
	bz, err := suite.EncodingConfig.TxConfig.TxEncoder()(tx)
	suite.Require().NoError(err)
	txHash := hex.EncodeToString(tmtypes.Tx(bz).Hash())

	txResponse, err := suite.QueryClients.ServiceClient.GetTx(context.Background(), &cosmostxtypes.GetTxRequest{
		Hash: txHash,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(txResponse)
	return txResponse
}
