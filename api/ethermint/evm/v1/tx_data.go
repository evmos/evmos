// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evmv1

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	_ TxDataV2 = &LegacyTx{}
	_ TxDataV2 = &AccessListTx{}
	_ TxDataV2 = &DynamicFeeTx{}
)

// TxDataV2 implements the Ethereum transaction tx structure. It is used
// solely as intended in Ethereum abiding by the protocol.
type TxDataV2 interface {
	GetChainID() *big.Int
	GetAccessList() ethtypes.AccessList
	GetData() []byte
	GetNonce() uint64
	GetGas() uint64
	GetToAddress() *common.Address

	GetRawSignatureValues() (v, r, s *big.Int)
	AsEthereumData() ethtypes.TxData

	ProtoReflect() protoreflect.Message
}

func rawSignatureValues(vBz, rBz, sBz []byte) (v, r, s *big.Int) {
	if len(vBz) > 0 {
		v = new(big.Int).SetBytes(vBz)
	}
	if len(rBz) > 0 {
		r = new(big.Int).SetBytes(rBz)
	}
	if len(sBz) > 0 {
		s = new(big.Int).SetBytes(sBz)
	}
	return v, r, s
}

// helper function to parse string to bigInt
func stringToBigInt(str string) *big.Int {
	if str == "" {
		return nil
	}
	res, ok := sdkmath.NewIntFromString(str)
	if !ok {
		return nil
	}
	return res.BigInt()
}
