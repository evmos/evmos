// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evmv1

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethutils "github.com/evmos/evmos/v20/utils/eth"
)

// GetChainID returns the chain id field from the DynamicFeeTx
func (tx *DynamicFeeTx) GetChainID() *big.Int {
	return stringToBigInt(tx.GetChainId())
}

// AsEthereumData returns an DynamicFeeTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *DynamicFeeTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.DynamicFeeTx{
		ChainID:    tx.GetChainID(),
		Nonce:      tx.GetNonce(),
		GasTipCap:  stringToBigInt(tx.GetGasTipCap()),
		GasFeeCap:  stringToBigInt(tx.GetGasFeeCap()),
		Gas:        tx.GetGas(),
		To:         stringToAddress(tx.GetTo()),
		Value:      stringToBigInt(tx.GetValue()),
		Data:       tx.GetData(),
		AccessList: tx.GetAccessList(),
		V:          v,
		R:          r,
		S:          s,
	}
}

// GetAccessList returns the AccessList field.
func (tx *DynamicFeeTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	var ethAccessList ethtypes.AccessList

	for _, tuple := range tx.Accesses {
		storageKeys := make([]common.Hash, len(tuple.StorageKeys))

		for i := range tuple.StorageKeys {
			storageKeys[i] = common.HexToHash(tuple.StorageKeys[i])
		}

		ethAccessList = append(ethAccessList, ethtypes.AccessTuple{
			Address:     common.HexToAddress(tuple.Address),
			StorageKeys: storageKeys,
		})
	}

	return ethAccessList
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *DynamicFeeTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return ethutils.RawSignatureValues(tx.V, tx.R, tx.S)
}
