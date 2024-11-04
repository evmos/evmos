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
// solely to define the custom logic for getting signers on Ethereum transactions.
type TxDataV2 interface {
	GetChainID() *big.Int
	AsEthereumData() ethtypes.TxData

	ProtoReflect() protoreflect.Message
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

func stringToAddress(toStr string) *common.Address {
	if toStr == "" {
		return nil
	}
	addr := common.HexToAddress(toStr)
	return &addr
}
