// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evmv1

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	protov2 "google.golang.org/protobuf/proto"
)

// supportedTxs holds the Ethereum transaction types
// supported by Evmos.
// Use a function to return a new pointer and avoid
// possible reuse or racing conditions when using the same pointer
var supportedTxs = map[string]func() TxDataV2{
	"/ethermint.evm.v1.DynamicFeeTx": func() TxDataV2 { return &DynamicFeeTx{} },
	"/ethermint.evm.v1.AccessListTx": func() TxDataV2 { return &AccessListTx{} },
	"/ethermint.evm.v1.LegacyTx":     func() TxDataV2 { return &LegacyTx{} },
}

// getSender extracts the sender address from the signature values using the latest signer for the given chainID.
func getSender(txData TxDataV2) (common.Address, error) {
	signer := ethtypes.LatestSignerForChainID(txData.GetChainID())
	from, err := signer.Sender(ethtypes.NewTx(txData.AsEthereumData()))
	if err != nil {
		return common.Address{}, err
	}
	return from, nil
}

// GetSigners is the custom function to get signers on Ethereum transactions
// Gets the signer's address from the Ethereum tx signature
func GetSigners(msg protov2.Message) ([][]byte, error) {
	msgEthTx, ok := msg.(*MsgEthereumTx)
	if !ok {
		return nil, fmt.Errorf("invalid type, expected MsgEthereumTx and got %T", msg)
	}

	txDataFn, found := supportedTxs[msgEthTx.Data.TypeUrl]
	if !found {
		return nil, fmt.Errorf("invalid TypeUrl %s", msgEthTx.Data.TypeUrl)
	}
	txData := txDataFn()

	// msgEthTx.Data is a message (DynamicFeeTx, LegacyTx or AccessListTx)
	if err := msgEthTx.Data.UnmarshalTo(txData); err != nil {
		return nil, err
	}

	sender, err := getSender(txData)
	if err != nil {
		return nil, err
	}

	return [][]byte{sender.Bytes()}, nil
}
