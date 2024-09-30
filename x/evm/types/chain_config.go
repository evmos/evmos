// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// This config provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	geth "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v20/types"
)

// testChainID represents the ChainID used for the purpose of testing.
const testChainID string = "evmos_9002-1"

// chainConfig is the chain configuration used in the EVM to defined which
// opcodes are active based on Ethereum upgrades.
var chainConfig *ChainConfig

// EthereumConfig returns an Ethereum ChainConfig for EVM state transitions.
// All the negative or nil values are converted to nil
func (cc ChainConfig) EthereumConfig(chainID *big.Int) *params.ChainConfig {
	cId := big.NewInt(int64(cc.ChainId))
	if chainID != nil {
		cId = chainID
	}
	return &params.ChainConfig{
		ChainID:                 cId,
		HomesteadBlock:          getBlockValue(cc.HomesteadBlock),
		DAOForkBlock:            getBlockValue(cc.DAOForkBlock),
		DAOForkSupport:          cc.DAOForkSupport,
		EIP150Block:             getBlockValue(cc.EIP150Block),
		EIP150Hash:              common.HexToHash(cc.EIP150Hash),
		EIP155Block:             getBlockValue(cc.EIP155Block),
		EIP158Block:             getBlockValue(cc.EIP158Block),
		ByzantiumBlock:          getBlockValue(cc.ByzantiumBlock),
		ConstantinopleBlock:     getBlockValue(cc.ConstantinopleBlock),
		PetersburgBlock:         getBlockValue(cc.PetersburgBlock),
		IstanbulBlock:           getBlockValue(cc.IstanbulBlock),
		MuirGlacierBlock:        getBlockValue(cc.MuirGlacierBlock),
		BerlinBlock:             getBlockValue(cc.BerlinBlock),
		LondonBlock:             getBlockValue(cc.LondonBlock),
		ArrowGlacierBlock:       getBlockValue(cc.ArrowGlacierBlock),
		GrayGlacierBlock:        getBlockValue(cc.GrayGlacierBlock),
		MergeNetsplitBlock:      getBlockValue(cc.MergeNetsplitBlock),
		ShanghaiBlock:           getBlockValue(cc.ShanghaiBlock),
		CancunBlock:             getBlockValue(cc.CancunBlock),
		TerminalTotalDifficulty: nil,
		Ethash:                  nil,
		Clique:                  nil,
	}
}

func DefaultChainConfig(chainID string) *ChainConfig {
	if chainID == "" {
		chainID = testChainID
	}

	eip155ChainID, err := types.ParseChainID(chainID)
	if err != nil {
		panic(err)
	}

	homesteadBlock := sdkmath.ZeroInt()
	daoForkBlock := sdkmath.ZeroInt()
	eip150Block := sdkmath.ZeroInt()
	eip155Block := sdkmath.ZeroInt()
	eip158Block := sdkmath.ZeroInt()
	byzantiumBlock := sdkmath.ZeroInt()
	constantinopleBlock := sdkmath.ZeroInt()
	petersburgBlock := sdkmath.ZeroInt()
	istanbulBlock := sdkmath.ZeroInt()
	muirGlacierBlock := sdkmath.ZeroInt()
	berlinBlock := sdkmath.ZeroInt()
	londonBlock := sdkmath.ZeroInt()
	arrowGlacierBlock := sdkmath.ZeroInt()
	grayGlacierBlock := sdkmath.ZeroInt()
	mergeNetsplitBlock := sdkmath.ZeroInt()
	shanghaiBlock := sdkmath.ZeroInt()
	cancunBlock := sdkmath.ZeroInt()
	cfg := &ChainConfig{
		ChainId:             eip155ChainID.Uint64(),
		HomesteadBlock:      &homesteadBlock,
		DAOForkBlock:        &daoForkBlock,
		DAOForkSupport:      true,
		EIP150Block:         &eip150Block,
		EIP150Hash:          common.Hash{}.String(),
		EIP155Block:         &eip155Block,
		EIP158Block:         &eip158Block,
		ByzantiumBlock:      &byzantiumBlock,
		ConstantinopleBlock: &constantinopleBlock,
		PetersburgBlock:     &petersburgBlock,
		IstanbulBlock:       &istanbulBlock,
		MuirGlacierBlock:    &muirGlacierBlock,
		BerlinBlock:         &berlinBlock,
		LondonBlock:         &londonBlock,
		ArrowGlacierBlock:   &arrowGlacierBlock,
		GrayGlacierBlock:    &grayGlacierBlock,
		MergeNetsplitBlock:  &mergeNetsplitBlock,
		ShanghaiBlock:       &shanghaiBlock,
		CancunBlock:         &cancunBlock,
	}
	return cfg
}

// SetChainConfig allows to set the `chainConfig` variable modifying the
// default values. The method is private because it should only be called once
// in the EVMConfigurator.
func SetChainConfig(cc *ChainConfig) {
	if cc != nil {
		chainConfig = cc
		return
	}

	chainConfig = DefaultChainConfig("")
}

// GetChainConfig returns the `chainConfig` used in the EVM.
func GetChainConfig() *geth.ChainConfig {
	return chainConfig.EthereumConfig(nil)
}

func getBlockValue(block *sdkmath.Int) *big.Int {
	if block == nil || block.IsNegative() {
		return nil
	}

	return block.BigInt()
}
