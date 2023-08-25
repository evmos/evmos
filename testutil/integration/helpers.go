package testnetwork

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	cosmostx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/gogo/protobuf/proto"

	sdkmath "cosmossdk.io/math"

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
	"github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"

	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/server/config"
	evm "github.com/evmos/evmos/v14/x/evm/types"
)

// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments
func DeployContract(
	queryClientHelper *grpc.GrpcQueryHelper,
	priv cryptotypes.PrivKey,
	contract evm.CompiledContract,
	constructorArgs ...interface{},
) (common.Address, error) {
	// Get account's nonce to create contract hash
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	account, err := queryClientHelper.GetEvmAccount(from)
	if err != nil {
		return common.Address{}, err
	}
	nonce := account.GetNonce()

	ctorArgs, err := contract.ABI.Pack("", constructorArgs...)
	if err != nil {
		return common.Address{}, err
	}
	data := append(contract.Bin, ctorArgs...)

	args := evm.EvmTxArgs{
		Input: data,
		Nonce: nonce,
	}
	res, err := DoEthTx(queryClientHelper, priv, args)
	if err != nil {
		return common.Address{}, err
	}

	if err := checkEthTxResponse(&res); err != nil {
		return common.Address{}, err
	}
	return crypto.CreateAddress(from, nonce), nil
}

// DoContractCall executes a contract call with the provided private key and txArgs
// It first builds a MsgEthereumTx and then broadcast it to the network.
func DoEthTx(
	queryClientHelper *grpc.GrpcQueryHelper,
	priv cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (abcitypes.ResponseDeliverTx, error) {
	msgEthereumTx, err := CreateMsgEthereumTx(queryClientHelper, priv, txArgs)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	signedMsg, err := SignMsgEthereumTx(msgEthereumTx, priv, queryClientHelper.GetChainID())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	txBytes, err := encodeTx(queryClientHelper, signedMsg)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	_, err = queryClientHelper.Simulate(txBytes)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	res, err := queryClientHelper.BroadcastTxSync(txBytes)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	if err := checkEthTxResponse(&res); err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	return res, nil
}

// checkEthTxResponse checks if the response is valid and returns the MsgEthereumTxResponse
func checkEthTxResponse(res *abcitypes.ResponseDeliverTx) error {
	var txData sdktypes.TxMsgData
	if !res.IsOK() {
		return fmt.Errorf("tx failed. Code: %d, Logs: %s", res.Code, res.Log)
	}

	cdc := encoding.MakeConfig(app.ModuleBasics).Codec
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

// CreateMsgEthereumTx creates a new MsgEthereumTx with the provided arguments.
// If any of the arguments are not provided, they will be populated with default values.
func CreateMsgEthereumTx(
	queryClientHelper *grpc.GrpcQueryHelper,
	privKey cryptotypes.PrivKey,
	txArgs evmtypes.EvmTxArgs,
) (evmtypes.MsgEthereumTx, error) {
	fromAddr := common.BytesToAddress(privKey.PubKey().Address().Bytes())
	// Fill TxArgs with default values
	txArgs, err := populateEvmTxArgs(queryClientHelper, fromAddr, txArgs)
	if err != nil {
		return evmtypes.MsgEthereumTx{}, err
	}

	return buildMsgEthereumTx(txArgs, fromAddr)
}

// populateEvmTxArgs populates the missing fields in the provided EvmTxArgs with default values.
// If no GasLimit is present it will estimate the gas needed for the transaction.
func populateEvmTxArgs(
	queryClientHelper *grpc.GrpcQueryHelper,
	fromAddr common.Address,
	txArgs evmtypes.EvmTxArgs,
) (evmtypes.EvmTxArgs, error) {
	if txArgs.ChainID == nil {
		chainID := queryClientHelper.GetChainID()
		ethChainID, err := types.ParseChainID(chainID)
		if err != nil {
			return evmtypes.EvmTxArgs{}, err
		}
		txArgs.ChainID = ethChainID
	}

	if txArgs.Nonce == 0 {
		accountResp, err := queryClientHelper.GetEvmAccount(fromAddr)
		if err != nil {
			return evmtypes.EvmTxArgs{}, err
		}
		txArgs.Nonce = accountResp.GetNonce()
	}

	// TODO - Check if this should be the case
	if txArgs.GasPrice == nil {
		if txArgs.GasTipCap == nil {
			txArgs.GasTipCap = big.NewInt(1)
		}
		if txArgs.GasFeeCap == nil {
			baseFeeResp, err := queryClientHelper.GetBaseFee()
			if err != nil {
				return evmtypes.EvmTxArgs{}, err
			}
			fmt.Println("baseFeeResp", baseFeeResp)
			txArgs.GasFeeCap = baseFeeResp.BaseFee.BigInt()
		}
	}

	// If the gas limit is not set, estimate it
	// through the /simulate endpoint.
	if txArgs.GasLimit == 0 {
		gasLimit, err := GasLimit(queryClientHelper, &fromAddr, &txArgs)
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

func buildMsgEthereumTx(txArgs evmtypes.EvmTxArgs, fromAddr common.Address) (evmtypes.MsgEthereumTx, error) {
	msgEthereumTx := evmtypes.NewTx(&txArgs)
	msgEthereumTx.From = fromAddr.String()

	// Validate the transaction to avoid unrealistic behaviour
	err := msgEthereumTx.ValidateBasic()
	if err != nil {
		return evmtypes.MsgEthereumTx{}, err
	}

	return *msgEthereumTx, nil
}

// SignMsgEthereumTx signs a MsgEthereumTx with the provided private key and chainID.
func SignMsgEthereumTx(msgEthereumTx evmtypes.MsgEthereumTx, privKey cryptotypes.PrivKey, chainID string) (evmtypes.MsgEthereumTx, error) {
	ethChainID, err := types.ParseChainID(chainID)
	if err != nil {
		return evmtypes.MsgEthereumTx{}, err
	}

	signer := ethtypes.LatestSignerForChainID(ethChainID)
	err = msgEthereumTx.Sign(signer, tx.NewSigner(privKey))
	if err != nil {
		return evmtypes.MsgEthereumTx{}, err
	}
	return msgEthereumTx, nil
}

// GasLimit estimates the gas limit for the provided parameters.
func GasLimit(queryClientHelper *grpc.GrpcQueryHelper, from *common.Address, txArgs *evmtypes.EvmTxArgs) (uint64, error) {
	args, err := json.Marshal(evmtypes.TransactionArgs{
		Data: (*hexutil.Bytes)(&txArgs.Input),
		From: from,
	})
	if err != nil {
		return 0, err
	}

	res, err := queryClientHelper.EstimateGas(args, config.DefaultGasCap)
	if err != nil {
		return 0, err
	}

	gas := res.Gas
	return gas, nil
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

var (
	defaultPrice  = sdkmath.NewIntFromUint64(uint64(math.Pow10(7)))
	gasAdjustment = float64(1.7)
)

func BuildAndBroadcastMsg(queryClient *grpc.GrpcQueryHelper, privKey cryptotypes.PrivKey, txArgs CosmosTxArgs) (abcitypes.ResponseDeliverTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
	txBuilder := txConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(txArgs.Msgs...); err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	if txArgs.FeeGranter != nil {
		txBuilder.SetFeeGranter(txArgs.FeeGranter)
	}

	senderAddress := sdktypes.AccAddress(privKey.PubKey().Address().Bytes())
	account, err := queryClient.GetAccount(senderAddress.String())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	chainID := queryClient.GetChainID()
	sequence := account.GetSequence()
	signMode := txConfig.SignModeHandler().DefaultMode()
	signerData := xauthsigning.SignerData{
		ChainID:       chainID,
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
	// Generated Protobuf-encoded bytes.
	simulateBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	var gasLimit uint64
	if txArgs.Gas == 0 {
		simulateRes, err := queryClient.Simulate(simulateBytes)
		if err != nil {
			return abcitypes.ResponseDeliverTx{}, err
		}
		gasLimit = uint64(gasAdjustment * float64(simulateRes.GasInfo.GasUsed))
	} else {
		gasLimit = txArgs.Gas
	}
	txBuilder.SetGasLimit(gasLimit)

	denom := queryClient.GetDenom()
	var fees sdktypes.Coins
	if txArgs.GasPrice != nil {
		fees = sdktypes.Coins{{Denom: denom, Amount: txArgs.GasPrice.MulRaw(int64(gasLimit))}}
	} else {
		baseFee, err := queryClient.GetBaseFee()
		if err != nil {
			return abcitypes.ResponseDeliverTx{}, err
		}
		price := baseFee.BaseFee
		fees = sdktypes.Coins{{Denom: denom, Amount: price.MulRaw(int64(gasLimit))}}
	}
	txBuilder.SetFeeAmount(fees)

	// txBuilder.SetFeeAmount(fees)
	signature, err := cosmostx.SignWithPrivKey(signMode, signerData, txBuilder, privKey, txConfig, sequence)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	txBuilder.SetSignatures(signature)

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	return queryClient.BroadcastTxSync(txBytes)
}

func SimulateGas(queryClient *grpc.GrpcQueryHelper, txBytes []byte) (uint64, error) {
	res, err := queryClient.Simulate(txBytes)
	if err != nil {
		return 0, err
	}
	return res.GasInfo.GasUsed, nil
}

func encodeTx(queryClientHelper *grpc.GrpcQueryHelper, msg evmtypes.MsgEthereumTx) ([]byte, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
	txBuilder := txConfig.NewTxBuilder()
	signingTx, err := msg.BuildTx(txBuilder, queryClientHelper.GetDenom())
	if err != nil {
		return nil, err
	}

	txBytes, err := txConfig.TxEncoder()(signingTx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}
