// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convenient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v20/types"
)

// testChainID represents the ChainID used for the purpose of testing.
const testChainID string = "evmos_9002-1"

// chainConfig is the chain configuration used in the EVM to define which
// opcodes are active based on Ethereum upgrades.
var chainConfig *geth.ChainConfig

// DefaultChainConfig returns a reference to the default `ChainConfig` in which
// all Ethereum upgrades happen at block 0. This means that all opcodes are
// active.
func DefaultChainConfig(chainID string) *geth.ChainConfig {
	if chainID == "" {
		chainID = testChainID
	}

	eip155ChainID, err := types.ParseChainID(chainID)
	if err != nil {
		panic(err)
	}
	blockZero := big.NewInt(0)
	cfg := &geth.ChainConfig{
		ChainID:                 eip155ChainID,
		HomesteadBlock:          blockZero,
		DAOForkBlock:            blockZero,
		DAOForkSupport:          true,
		EIP150Block:             blockZero,
		EIP150Hash:              common.Hash{},
		EIP155Block:             blockZero,
		EIP158Block:             blockZero,
		ByzantiumBlock:          blockZero,
		ConstantinopleBlock:     blockZero,
		PetersburgBlock:         blockZero,
		IstanbulBlock:           blockZero,
		MuirGlacierBlock:        blockZero,
		BerlinBlock:             blockZero,
		LondonBlock:             blockZero,
		ArrowGlacierBlock:       blockZero,
		GrayGlacierBlock:        blockZero,
		MergeNetsplitBlock:      blockZero,
		ShanghaiBlock:           blockZero,
		CancunBlock:             blockZero,
		TerminalTotalDifficulty: nil,
		Ethash:                  nil,
		Clique:                  nil,
	}
	return cfg
}

// setChainConfig allows to set the `chainConfig` variable modifying the
// default values. The method is private because it should only be called once
// in the EVMConfigurator.
func setChainConfig(cc *geth.ChainConfig) {
	if cc != nil {
		chainConfig = cc
		return
	}

	// If no reference to a ChainConfig is passed, fallback to a ChainConfig for
	// a test chain.
	chainConfig = DefaultChainConfig("")
}

// GetChainConfig returns the `chainConfig` used in the EVM.
func GetChainConfig() *geth.ChainConfig {
	return chainConfig
}
