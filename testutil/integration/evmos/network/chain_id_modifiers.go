// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// This files contains handler for the testing suite that has to be run to
// modify the chain configuration depending on the chainID

package network

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v20/utils"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

// updateErc20GenesisStateForChainID modify the default genesis state for the
// bank module of the testing suite depending on the chainID.
func updateBankGenesisStateForChainID(chainID string, bankGenesisState banktypes.GenesisState) banktypes.GenesisState {
	metadata := generateBankGenesisMetadata(chainID)
	bankGenesisState.DenomMetadata = []banktypes.Metadata{metadata}

	return bankGenesisState
}

// generateBankGenesisMetadata generates the metadata
// for the Evm coin depending on the chainID.
func generateBankGenesisMetadata(chainID string) banktypes.Metadata {
	if utils.IsTestnet(chainID) {
		return banktypes.Metadata{
			Description: "The native EVM, governance and staking token of the Evmos testnet",
			Base:        "atevmos",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "atevmos",
					Exponent: 0,
				},
				{
					Denom:    "tevmos",
					Exponent: 18,
				},
			},
			Name:    "tEvmos",
			Symbol:  "tEVMOS",
			Display: "tevmos",
		}
	}

	return banktypes.Metadata{
		Description: "The native EVM, governance and staking token of the Evmos mainnet",
		Base:        "aevmos",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "aevmos",
				Exponent: 0,
			},
			{
				Denom:    "evmos",
				Exponent: 18,
			},
		},
		Name:    "Evmos",
		Symbol:  "EVMOS",
		Display: "evmos",
	}
}

// updateErc20GenesisStateForChainID modify the default genesis state for the
// erc20 module on the testing suite depending on the chainID.
func updateErc20GenesisStateForChainID(chainID string, erc20GenesisState erc20types.GenesisState) erc20types.GenesisState {
	if !utils.IsTestnet(chainID) {
		return erc20GenesisState
	}

	erc20GenesisState.Params = updateErc20Params(chainID, erc20GenesisState.Params)
	erc20GenesisState.TokenPairs = updateErc20TokenPairs(chainID, erc20GenesisState.TokenPairs)

	return erc20GenesisState
}

// updateErc20Params modifies the erc20 module params to use the correct
// WEVMOS contract depending on ChainID
func updateErc20Params(chainID string, params erc20types.Params) erc20types.Params {
	mainnetAddress := erc20types.GetWEVMOSContractHex(utils.MainnetChainID)
	testnetAddress := erc20types.GetWEVMOSContractHex(chainID)

	nativePrecompiles := make([]string, len(params.NativePrecompiles))
	for i, nativePrecompile := range params.NativePrecompiles {
		if nativePrecompile == mainnetAddress {
			nativePrecompiles[i] = testnetAddress
		} else {
			nativePrecompiles[i] = nativePrecompile
		}
	}
	params.NativePrecompiles = nativePrecompiles
	return params
}

// updateErc20TokenPairs modifies the erc20 token pairs to use the correct
// WEVMOS depending on ChainID
func updateErc20TokenPairs(chainID string, tokenPairs []erc20types.TokenPair) []erc20types.TokenPair {
	testnetAddress := erc20types.GetWEVMOSContractHex(chainID)
	coinInfo := evmtypes.ChainsCoinInfo[utils.MainnetChainID]

	mainnetAddress := erc20types.GetWEVMOSContractHex(utils.MainnetChainID)

	updatedTokenPairs := make([]erc20types.TokenPair, len(tokenPairs))
	for i, tokenPair := range tokenPairs {
		if tokenPair.Erc20Address == mainnetAddress {
			updatedTokenPairs[i] = erc20types.TokenPair{
				Erc20Address:  testnetAddress,
				Denom:         coinInfo.Denom,
				Enabled:       tokenPair.Enabled,
				ContractOwner: tokenPair.ContractOwner,
			}
		} else {
			updatedTokenPairs[i] = tokenPair
		}
	}
	return updatedTokenPairs
}
