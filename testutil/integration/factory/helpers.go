package factory

import (
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/testutil/tx"
	evmostypes "github.com/evmos/evmos/v14/types"
	"github.com/gogo/protobuf/proto"
)

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

func buildAndEncodeEthTx(msg evmtypes.MsgEthereumTx, evmDenom string) ([]byte, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
	txBuilder := txConfig.NewTxBuilder()
	signingTx, err := msg.BuildTx(txBuilder, evmDenom)
	if err != nil {
		return nil, err
	}

	txBytes, err := txConfig.TxEncoder()(signingTx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func buildMsgEthereumTx(txArgs evmtypes.EvmTxArgs, fromAddr common.Address) (evmtypes.MsgEthereumTx, error) {
	msgEthereumTx := evmtypes.NewTx(&txArgs)
	msgEthereumTx.From = fromAddr.String()

	// Validate the transaction to avoid unrealistic behavior
	err := msgEthereumTx.ValidateBasic()
	if err != nil {
		return evmtypes.MsgEthereumTx{}, err
	}
	return *msgEthereumTx, nil
}

// signMsgEthereumTx signs a MsgEthereumTx with the provided private key and chainID.
func signMsgEthereumTx(msgEthereumTx evmtypes.MsgEthereumTx, privKey cryptotypes.PrivKey, chainID string) (evmtypes.MsgEthereumTx, error) {
	ethChainID, err := evmostypes.ParseChainID(chainID)
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
