// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evmv1

import (
	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethutils "github.com/evmos/evmos/v20/utils/eth"
)

// GetChainID returns the chain id field from the derived signature values
func (tx *LegacyTx) GetChainID() *big.Int {
	v, _, _ := tx.GetRawSignatureValues()
	return ethutils.DeriveChainID(v)
}

// AsEthereumData returns an LegacyTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *LegacyTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.LegacyTx{
		Nonce:    tx.GetNonce(),
		GasPrice: stringToBigInt(tx.GetGasPrice()),
		Gas:      tx.GetGas(),
		To:       stringToAddress(tx.GetTo()),
		Value:    stringToBigInt(tx.GetValue()),
		Data:     tx.GetData(),
		V:        v,
		R:        r,
		S:        s,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *LegacyTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return ethutils.RawSignatureValues(tx.V, tx.R, tx.S)
}

// GetAccessList returns nil
func (tx *LegacyTx) GetAccessList() ethtypes.AccessList {
	return nil
}
