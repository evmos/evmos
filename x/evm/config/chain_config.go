// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v20/types"
)

// chainConfig is the chain configuration used in the EVM to defined which
// opcodes are active based on Ethereum upgrades.
var chainConfig *geth.ChainConfig = nil

func DefaultChainConfig(chainID *big.Int) *geth.ChainConfig {
	qq := &geth.ChainConfig{
		ChainID:                 chainID,
		HomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            big.NewInt(0),
		DAOForkSupport:          true,
		EIP150Block:             big.NewInt(0),
		EIP150Hash:              common.Hash{},
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		GrayGlacierBlock:        big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		ShanghaiBlock:           big.NewInt(0),
		CancunBlock:             big.NewInt(0),
		TerminalTotalDifficulty: nil,
		Ethash:                  nil,
		Clique:                  nil,
	}
	return qq
}

// setChainConfig allows to set the `chainConfig` variable modifying the
// default values. The method is private because it should only be called once
// in the EVMConfigurator.
func setChainConfig(cc *geth.ChainConfig, chainID string) error {
	if cc != nil {
		chainConfig = cc
		return nil
	}

	eip155ChainID, err := types.ParseChainID(chainID)
	if err != nil {
		return err
	}

	chainConfig = DefaultChainConfig(eip155ChainID)
	return nil
}

// GetChainConfig returns the `chainConfig` used in the EVM.
func GetChainConfig() *geth.ChainConfig {
	return chainConfig
}
