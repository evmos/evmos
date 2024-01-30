// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package factory

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/encoding"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutiltypes "github.com/cosmos/cosmos-sdk/types/module/testutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v16/testutil/tx"
	evmostypes "github.com/evmos/evmos/v16/types"
)

// buildMsgEthereumTx builds an Ethereum transaction from the given arguments and populates the From field.
func buildMsgEthereumTx(txArgs evmtypes.EvmTxArgs, fromAddr common.Address) evmtypes.MsgEthereumTx {
	msgEthereumTx := evmtypes.NewTx(&txArgs)
	msgEthereumTx.From = fromAddr.String()
	return *msgEthereumTx
}

// signMsgEthereumTx signs a MsgEthereumTx with the provided private key and chainID.
func signMsgEthereumTx(msgEthereumTx evmtypes.MsgEthereumTx, privKey cryptotypes.PrivKey, chainID string) (evmtypes.MsgEthereumTx, error) {
	ethChainID, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrapf(err, "failed to parse chainID: %v", chainID)
	}

	signer := ethtypes.LatestSignerForChainID(ethChainID)
	err = msgEthereumTx.Sign(signer, tx.NewSigner(privKey))
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to sign transaction")
	}
	return msgEthereumTx, nil
}

// makeConfig creates an EncodingConfig for testing
func makeConfig(mb module.BasicManager) testutiltypes.TestEncodingConfig {
	return encoding.MakeConfig(mb)
}
