// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"hash"

	"github.com/cometbft/cometbft/crypto/txhash"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	_ txhash.Hash = &Hash{}
	_ hash.Hash   = &ethHash{}
)

// Hash defines a constructor for Ethereum transaction hasher.
type Hash struct{}

func (h *Hash) New() hash.Hash {
	return &ethHash{}
}

type ethHash struct {
	tx *ethtypes.Transaction
}

// Write decodes the raw transaction bytes into an Ethereum transaction.
func (h *ethHash) Write(p []byte) (n int, err error) {
	if err := h.tx.UnmarshalBinary(p); err != nil {
		panic(err)
	}

	return len(p), nil
}

// Sum computes the hash of the ethereum tx using RLP encoding.
func (h *ethHash) Sum(b []byte) []byte {
	if h.tx == nil {
		return nil
	}

	return h.tx.Hash().Bytes()
}

// Resets the hash to its initial state. No-op
func (h *ethHash) Reset() {}

// Size returns the size of the hash in bytes.
func (h *ethHash) Size() int {
	return common.HashLength
}

// FIXME: Should be the same as KeccakState but not sure
func (h *ethHash) BlockSize() int {
	return crypto.NewKeccakState().BlockSize()
}

// HashFmtFunc returns an 0x prefixed representation of the transaction hash
var HashFmtFunc = func(bz []byte) string {
	return common.Bytes2Hex(bz)
}
