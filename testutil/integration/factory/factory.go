// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	cosmostx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/gogoproto/proto"

	sdkmath "cosmossdk.io/math"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v15/testutil/integration/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/network"
	"github.com/evmos/evmos/v15/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/server/config"
)

const (
	GasAdjustment = float64(1.7)
)

type TxFactory interface {
	// DeployContract deploys a contract with the provided private key,
	// compiled contract data and constructor arguments
	DeployContract(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, deploymentData ContractDeploymentData) (common.Address, error)
	// ExecuteContractCall executes a contract call with the provided private key
	ExecuteContractCall(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, callArgs CallArgs) (abcitypes.ResponseDeliverTx, error)
	// ExecuteEthTx builds, signs and broadcasts an Ethereum tx with the provided private key and txArgs.
	// If the txArgs are not provided, they will be populated with default values or gas estimations.
	ExecuteEthTx(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs) (abcitypes.ResponseDeliverTx, error)
	// ExecuteEthTx builds, signs and broadcasts a Cosmos tx with the provided private key and txArgs
	ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error)
	// EstimateGasLimit estimates the gas limit for a tx with the provided address and txArgs
	EstimateGasLimit(from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error)
}

var _ TxFactory = (*IntegrationTxFactory)(nil)

// IntegrationTxFactory is a helper struct to build and broadcast transactions
// to the network on integration tests. This is to simulate the behavior of a real user.
type IntegrationTxFactory struct {
	grpcHandler grpc.Handler
	network     network.Network
	ec          *testutiltypes.TestEncodingConfig
}

// New creates a new IntegrationTxFactory instance
func New(
	network network.Network,
	grpcHandler grpc.Handler,
) TxFactory {
	ec := makeConfig(app.ModuleBasics)
	return &IntegrationTxFactory{
		grpcHandler: grpcHandler,
		network:     network,
		ec:          &ec,
	}
}

// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments.
// TxArgs Input and Nonce fields are overwritten.
func (tf *IntegrationTxFactory) DeployContract(
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
	deploymentData ContractDeploymentData,
) (common.Address, error) {
	// Get account's nonce to create contract hash
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	account, err := tf.grpcHandler.GetEvmAccount(from)
	if err != nil {
		return common.Address{}, errorsmod.Wrapf(err, "failed to get evm account: %s", from.String())
	}
	nonce := account.GetNonce()

	ctorArgs, err := deploymentData.Contract.ABI.Pack("", deploymentData.ConstructorArgs...)
	if err != nil {
		return common.Address{}, errorsmod.Wrap(err, "failed to pack constructor arguments")
	}
	data := deploymentData.Contract.Bin
	data = append(data, ctorArgs...)

	txArgs.Input = data
	txArgs.Nonce = nonce
	res, err := tf.ExecuteEthTx(priv, txArgs)
	if err != nil || !res.IsOK() {
		return common.Address{}, errorsmod.Wrap(err, "failed to execute eth tx")
	}
	return crypto.CreateAddress(from, nonce), nil
}

// ExecuteContractCall executes a contract call with the provided private key
func (tf *IntegrationTxFactory) ExecuteContractCall(privKey cryptotypes.PrivKey, txArgs evmtypes.EvmTxArgs, callArgs CallArgs) (abcitypes.ResponseDeliverTx, error) {
	// Create MsgEthereumTx that calls the contract
	input, err := callArgs.ContractABI.Pack(callArgs.MethodName, callArgs.Args...)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to pack contract arguments")
	}
	txArgs.Input = input

	return tf.ExecuteEthTx(privKey, txArgs)
}

// ExecuteEthTx executes an Ethereum transaction - contract call with the provided private key and txArgs
// It first builds a MsgEthereumTx and then broadcasts it to the network.
func (tf *IntegrationTxFactory) ExecuteEthTx(
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (abcitypes.ResponseDeliverTx, error) {
	msgEthereumTx, err := tf.createMsgEthereumTx(priv, txArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to create ethereum tx")
	}

	signedMsg, err := signMsgEthereumTx(msgEthereumTx, priv, tf.network.GetChainID())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to sign ethereum tx")
	}

	txBytes, err := tf.buildAndEncodeEthTx(signedMsg)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to build and encode ethereum tx")
	}

	res, err := tf.network.BroadcastTxSync(txBytes)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to broadcast ethereum tx")
	}

	if err := tf.checkEthTxResponse(&res); err != nil {
		return res, errorsmod.Wrap(err, "failed ETH tx")
	}
	return res, nil
}

// ExecuteCosmosTx creates, signs and broadcasts a Cosmos transaction
func (tf *IntegrationTxFactory) ExecuteCosmosTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error) {
	txBuilder, err := tf.buildTx(privKey, txArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to build tx")
	}

	txBytes, err := tf.encodeTx(txBuilder)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, errorsmod.Wrap(err, "failed to encode tx")
	}

	return tf.network.BroadcastTxSync(txBytes)
}

// EstimateGasLimit estimates the gas limit for a tx with the provided address and txArgs
func (tf *IntegrationTxFactory) EstimateGasLimit(from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error) {
	args, err := json.Marshal(evmtypes.TransactionArgs{
		Data:       (*hexutil.Bytes)(&txArgs.Input),
		From:       from,
		To:         txArgs.To,
		AccessList: txArgs.Accesses,
	})
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to marshal tx args")
	}

	res, err := tf.grpcHandler.EstimateGas(args, config.DefaultGasCap)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to estimate gas")
	}
	gas := res.Gas
	return gas, nil
}

// buildTx builds a tx with the provided private key and txArgs
func (tf *IntegrationTxFactory) buildTx(privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (client.TxBuilder, error) {
	txConfig := tf.ec.TxConfig
	txBuilder := txConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(txArgs.Msgs...); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set tx msgs")
	}

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	senderAddress := sdktypes.AccAddress(privKey.PubKey().Address().Bytes())
	account, err := tf.grpcHandler.GetAccount(senderAddress.String())
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to get account: %s", senderAddress.String())
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
		return nil, errorsmod.Wrap(err, "failed to set tx signatures")
	}

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	txBuilder.SetFeePayer(senderAddress)

	gasLimit, err := tf.estimateGas(txArgs, txBuilder)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to estimate gas")
	}
	txBuilder.SetGasLimit(gasLimit)

	fees, err := tf.calculateFees(txArgs.GasPrice, gasLimit)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to calculate fees")
	}
	txBuilder.SetFeeAmount(fees)

	signature, err := cosmostx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, txConfig, sequence)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to sign tx")
	}

	err = txBuilder.SetSignatures(signature)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to set tx signatures")
	}

	return txBuilder, nil
}

// encodeTx encodes the tx using the txConfig's encoder.
func (tf *IntegrationTxFactory) encodeTx(txBuilder client.TxBuilder) ([]byte, error) {
	txConfig := tf.ec.TxConfig
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to encode tx")
	}
	return txBytes, nil
}

// calculateFees calculates the fees for the transaction.
func (tf *IntegrationTxFactory) calculateFees(gasPrice *sdkmath.Int, gasLimit uint64) (sdktypes.Coins, error) {
	denom := tf.network.GetDenom()
	var fees sdktypes.Coins
	if gasPrice != nil {
		fees = sdktypes.Coins{{Denom: denom, Amount: gasPrice.MulRaw(int64(gasLimit))}}
	} else {
		baseFee, err := tf.grpcHandler.GetBaseFee()
		if err != nil {
			return sdktypes.Coins{}, errorsmod.Wrap(err, "failed to get base fee")
		}
		price := baseFee.BaseFee
		fees = sdktypes.Coins{{Denom: denom, Amount: price.MulRaw(int64(gasLimit))}}
	}
	return fees, nil
}

// estimateGas estimates the gas needed for the transaction.
func (tf *IntegrationTxFactory) estimateGas(txArgs CosmosTxArgs, txBuilder client.TxBuilder) (uint64, error) {
	txConfig := tf.ec.TxConfig
	simulateBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to encode tx")
	}

	var gasLimit uint64
	if txArgs.Gas == 0 {
		simulateRes, err := tf.network.Simulate(simulateBytes)
		if err != nil {
			return 0, errorsmod.Wrap(err, "failed to simulate tx")
		}

		gasAdj := new(big.Float).SetFloat64(GasAdjustment)
		gasUsed := new(big.Float).SetUint64(simulateRes.GasInfo.GasUsed)
		gasLimit, _ = gasAdj.Mul(gasAdj, gasUsed).Uint64()
	} else {
		gasLimit = txArgs.Gas
	}
	return gasLimit, nil
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
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to populate tx args")
	}
	msg := buildMsgEthereumTx(txArgs, fromAddr)

	return msg, nil
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
			return evmtypes.EvmTxArgs{}, errorsmod.Wrapf(err, "failed to parse chain id: %v", tf.network.GetChainID())
		}
		txArgs.ChainID = ethChainID
	}

	if txArgs.Nonce == 0 {
		accountResp, err := tf.grpcHandler.GetEvmAccount(fromAddr)
		if err != nil {
			return evmtypes.EvmTxArgs{}, errorsmod.Wrapf(err, "failed to get evm account: %s", fromAddr.String())
		}
		txArgs.Nonce = accountResp.GetNonce()
	}

	// If there is no GasPrice it is assume this is a DynamicFeeTx.
	// If fields are empty they are populated with current dynamic values.
	if txArgs.GasPrice == nil {
		if txArgs.GasTipCap == nil {
			txArgs.GasTipCap = big.NewInt(1)
		}
		if txArgs.GasFeeCap == nil {
			baseFeeResp, err := tf.grpcHandler.GetBaseFee()
			if err != nil {
				return evmtypes.EvmTxArgs{}, errorsmod.Wrap(err, "failed to get base fee")
			}
			txArgs.GasFeeCap = baseFeeResp.BaseFee.BigInt()
		}
	}

	// If the gas limit is not set, estimate it
	// through the /simulate endpoint.
	if txArgs.GasLimit == 0 {
		gasLimit, err := tf.EstimateGasLimit(&fromAddr, &txArgs)
		if err != nil {
			return evmtypes.EvmTxArgs{}, errorsmod.Wrap(err, "failed to estimate gas limit")
		}
		txArgs.GasLimit = gasLimit
	}

	return txArgs, nil
}

func (tf *IntegrationTxFactory) buildAndEncodeEthTx(msg evmtypes.MsgEthereumTx) ([]byte, error) {
	txConfig := tf.ec.TxConfig
	txBuilder := txConfig.NewTxBuilder()
	signingTx, err := msg.BuildTx(txBuilder, tf.network.GetDenom())
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to build tx")
	}

	txBytes, err := txConfig.TxEncoder()(signingTx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to encode tx")
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
		return errorsmod.Wrap(err, "failed to unmarshal tx data")
	}

	if len(txData.MsgResponses) != 1 {
		return fmt.Errorf("expected 1 message response, got %d", len(txData.MsgResponses))
	}

	var evmRes evmtypes.MsgEthereumTxResponse
	if err := proto.Unmarshal(txData.MsgResponses[0].Value, &evmRes); err != nil {
		return errorsmod.Wrap(err, "failed to unmarshal evm tx response")
	}

	if evmRes.Failed() {
		return fmt.Errorf("tx failed. VmError: %v, Logs: %s", evmRes.VmError, res.GetLog())
	}
	return nil
}
