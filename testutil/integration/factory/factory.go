package factory

import (
	"encoding/json"
	"fmt"
	"math/big"

	cosmostx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/gogoproto/proto"

	sdkmath "cosmossdk.io/math"
	simappparams "cosmossdk.io/simapp/params"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v14/testutil/integration/grpc"
	"github.com/evmos/evmos/v14/testutil/integration/network"
	"github.com/evmos/evmos/v14/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"

	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/server/config"
)

const (
	gasAdjustment = float64(1.7)
)

type TxFactory interface {
	// DeployContract deploys a contract with the provided private key,
	// compiled contract data and constructor arguments
	DeployContract(privKey cryptotypes.PrivKey, contract evmtypes.CompiledContract, constructorArgs ...interface{}) (common.Address, error)
	// ExecuteEthTx builds, signs and broadcast an Ethereum tx with the provided private key and txArgs
	ExecuteEthTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (abcitypes.ResponseDeliverTx, error)
	// DoCosmosTx builds, signs and broadcast a Cosmos tx with the provided private key and txArgs
	ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error)
	// EstimateGasLimit estimates the gas limit for a tx with the provided private key and txArgs
	EstimateGasLimit(from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error)
}

var _ TxFactory = (*IntegrationTxFactory)(nil)

// IntegrationTxFactory is a helper struct to build and broadcast transactions
// to the network on integration tests. This is to simulate the behavior of a real user.
type IntegrationTxFactory struct {
	grpcHandler grpc.Handler
	network     network.Network
	ec          *simappparams.EncodingConfig
}

// New creates a new IntegrationTxFactory instance
func New(
	grpcHandler grpc.Handler,
	network network.Network,
) *IntegrationTxFactory {
	ec := encoding.MakeConfig(app.ModuleBasics)
	return &IntegrationTxFactory{
		grpcHandler: grpcHandler,
		network:     network,
		ec:          &ec,
	}
}

// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments
func (tf *IntegrationTxFactory) DeployContract(
	priv cryptotypes.PrivKey,
	contract evmtypes.CompiledContract,
	constructorArgs ...interface{},
) (common.Address, error) {
	// Get account's nonce to create contract hash
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	account, err := tf.grpcHandler.GetEvmAccount(from)
	if err != nil {
		return common.Address{}, err
	}
	nonce := account.GetNonce()

	ctorArgs, err := contract.ABI.Pack("", constructorArgs...)
	if err != nil {
		return common.Address{}, err
	}
	data := contract.Bin
	data = append(data, ctorArgs...)

	args := evmtypes.EvmTxArgs{
		Input: data,
		Nonce: nonce,
	}
	_, err = tf.ExecuteEthTx(priv, args)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.CreateAddress(from, nonce), nil
}

// ExecuteEthTx executes an Ethereum transaction - contract call with the provided private key and txArgs
// It first builds a MsgEthereumTx and then broadcast it to the network.
func (tf *IntegrationTxFactory) ExecuteEthTx(
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (abcitypes.ResponseDeliverTx, error) {
	msgEthereumTx, err := tf.createMsgEthereumTx(priv, txArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	signedMsg, err := signMsgEthereumTx(msgEthereumTx, priv, tf.network.GetChainID())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	txBytes, err := tf.buildAndEncodeEthTx(signedMsg)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	res, err := tf.network.BroadcastTxSync(txBytes)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	if err := tf.checkEthTxResponse(&res); err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}
	return res, nil
}

// CosmosTxArgs contains the params to create a cosmos tx
type CosmosTxArgs struct {
	// ChainID is the chain's id on cosmos format, e.g. 'evmos_9000-1'
	ChainID string
	// Gas to be used on the tx
	Gas uint64
	// GasPrice to use on tx
	GasPrice *sdkmath.Int
	// Fees is the fee to be used on the tx (amount and denom)
	Fees sdktypes.Coins
	// FeeGranter is the account address of the fee granter
	FeeGranter sdktypes.AccAddress
	// Msgs slice of messages to include on the tx
	Msgs []sdktypes.Msg
}

// ExecuteCosmosTx creates, signs and broadcasts a Cosmos transaction
func (tf *IntegrationTxFactory) ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error) {
	txConfig := tf.ec.TxConfig
	txBuilder := txConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(txArgs.Msgs...); err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	senderAddress := sdktypes.AccAddress(privKey.PubKey().Address().Bytes())
	account, err := tf.grpcHandler.GetAccount(senderAddress.String())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	sequence := account.GetSequence()
	signMode := txConfig.SignModeHandler().DefaultMode()
	signerData := xauthsigning.SignerData{
		ChainID:       tf.network.GetChainID(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      sequence,
		Address:       senderAddress.String(),
	}

	// sign tx
	sigsV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		},
		Sequence: sequence,
	}

	err = txBuilder.SetSignatures(sigsV2)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	txBuilder.SetFeePayer(senderAddress)
	simulateBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	var gasLimit uint64
	if txArgs.Gas == 0 {
		simulateRes, err := tf.network.Simulate(simulateBytes)
		if err != nil {
			return abcitypes.ResponseDeliverTx{}, err
		}
		gasLimit = uint64(gasAdjustment * float64(simulateRes.GasInfo.GasUsed))
	} else {
		gasLimit = txArgs.Gas
	}
	txBuilder.SetGasLimit(gasLimit)

	denom := tf.network.GetDenom()
	var fees sdktypes.Coins
	if txArgs.GasPrice != nil {
		fees = sdktypes.Coins{{Denom: denom, Amount: txArgs.GasPrice.MulRaw(int64(gasLimit))}}
	} else {
		baseFee, err := tf.grpcHandler.GetBaseFee()
		if err != nil {
			return abcitypes.ResponseDeliverTx{}, err
		}
		price := baseFee.BaseFee
		fees = sdktypes.Coins{{Denom: denom, Amount: price.MulRaw(int64(gasLimit))}}
	}
	txBuilder.SetFeeAmount(fees)

	signature, err := cosmostx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, txConfig, sequence)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	err = txBuilder.SetSignatures(signature)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}
	return tf.network.BroadcastTxSync(txBytes)
}

// EstimateGasLimit estimates the gas limit for a tx with the provided address and txArgs
func (tf *IntegrationTxFactory) EstimateGasLimit(from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error) {
	args, err := json.Marshal(evmtypes.TransactionArgs{
		Data: (*hexutil.Bytes)(&txArgs.Input),
		From: from,
	})
	if err != nil {
		return 0, err
	}

	res, err := tf.grpcHandler.EstimateGas(args, config.DefaultGasCap)
	if err != nil {
		return 0, err
	}
	gas := res.Gas
	return gas, nil
}

// createMsgEthereumTx creates a new MsgEthereumTx with the provided arguments.
// If any of the arguments are not provided, they will be populated with default values.
func (tf *IntegrationTxFactory) createMsgEthereumTx(
	privKey cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (evmtypes.MsgEthereumTx, error) {
	fromAddr := common.BytesToAddress(privKey.PubKey().Address().Bytes())
	// Fill TxArgs with default values
	txArgs, err := tf.populateEvmTxArgs(fromAddr, txArgs)
	if err != nil {
		return evmtypes.MsgEthereumTx{}, err
	}
	return buildMsgEthereumTx(txArgs, fromAddr)
}

// populateEvmTxArgs populates the missing fields in the provided EvmTxArgs with default values.
// If no GasLimit is present it will estimate the gas needed for the transaction.
func (tf *IntegrationTxFactory) populateEvmTxArgs(
	fromAddr common.Address,
	txArgs evmtypes.EvmTxArgs,
) (evmtypes.EvmTxArgs, error) {
	if txArgs.ChainID == nil {
		ethChainID, err := types.ParseChainID(tf.network.GetChainID())
		if err != nil {
			return evmtypes.EvmTxArgs{}, err
		}
		txArgs.ChainID = ethChainID
	}

	if txArgs.Nonce == 0 {
		accountResp, err := tf.grpcHandler.GetEvmAccount(fromAddr)
		if err != nil {
			return evmtypes.EvmTxArgs{}, err
		}
		txArgs.Nonce = accountResp.GetNonce()
	}

	if txArgs.GasPrice == nil {
		if txArgs.GasTipCap == nil {
			txArgs.GasTipCap = big.NewInt(1)
		}
		if txArgs.GasFeeCap == nil {
			baseFeeResp, err := tf.grpcHandler.GetBaseFee()
			if err != nil {
				return evmtypes.EvmTxArgs{}, err
			}
			txArgs.GasFeeCap = baseFeeResp.BaseFee.BigInt()
		}
	}

	// If the gas limit is not set, estimate it
	// through the /simulate endpoint.
	if txArgs.GasLimit == 0 {
		gasLimit, err := tf.EstimateGasLimit(&fromAddr, &txArgs)
		if err != nil {
			return evmtypes.EvmTxArgs{}, err
		}
		txArgs.GasLimit = gasLimit
	}

	if txArgs.Accesses == nil {
		txArgs.Accesses = &ethtypes.AccessList{}
	}
	return txArgs, nil
}

func (tf *IntegrationTxFactory) buildAndEncodeEthTx(msg evmtypes.MsgEthereumTx) ([]byte, error) {
	txConfig := tf.ec.TxConfig
	txBuilder := txConfig.NewTxBuilder()
	signingTx, err := msg.BuildTx(txBuilder, tf.network.GetDenom())
	if err != nil {
		return nil, err
	}

	txBytes, err := txConfig.TxEncoder()(signingTx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

// checkEthTxResponse checks if the response is valid and returns the MsgEthereumTxResponse
func (tf *IntegrationTxFactory) checkEthTxResponse(res *abcitypes.ResponseDeliverTx) error {
	var txData sdktypes.TxMsgData
	if !res.IsOK() {
		return fmt.Errorf("tx failed. Code: %d, Logs: %s", res.Code, res.Log)
	}

	cdc := tf.ec.Codec
	if err := cdc.Unmarshal(res.Data, &txData); err != nil {
		return err
	}

	if len(txData.MsgResponses) != 1 {
		return fmt.Errorf("expected 1 message response, got %d", len(txData.MsgResponses))
	}

	var evmRes evmtypes.MsgEthereumTxResponse
	if err := proto.Unmarshal(txData.MsgResponses[0].Value, &evmRes); err != nil {
		return err
	}

	if evmRes.Failed() {
		return fmt.Errorf("tx failed. VmError: %v, Logs: %s", evmRes.VmError, res.GetLog())
	}
	return nil
}
