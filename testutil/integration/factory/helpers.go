package factory

import (
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v14/testutil/tx"
	evmostypes "github.com/evmos/evmos/v14/types"
)

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
