// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// This files contains handler for the testing suite that has to be run to
// modify the chain configuration depending on the chainID

package network

import (
	"github.com/evmos/evmos/v20/app"
	"github.com/evmos/evmos/v20/utils"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
)

// handleDefaultErc20GenesisState modify the default genesis state for the
// testing suite depending on the chainID.
func handleDefaultErc20GenesisState(evmosApp *app.Evmos, erc20GenesisState erc20types.GenesisState) erc20types.GenesisState {
	chainID := evmosApp.ChainID()
	if !utils.IsTestnet(chainID) {
		return erc20GenesisState
	}

	erc20GenesisState.Params = updateErc20Params(chainID, erc20GenesisState.Params)
	erc20GenesisState.TokenPairs = updateErc20TokenPairs(chainID, erc20GenesisState.TokenPairs)

	return erc20GenesisState
}

func updateErc20Params(chainID string, params erc20types.Params) erc20types.Params {
	mainnetAddress := erc20types.GetWEVMOSContractHex(utils.MainnetChainID)
	testnetAddress := erc20types.GetWEVMOSContractHex(chainID)

	for i, nativePrecompile := range params.NativePrecompiles {
		if nativePrecompile == mainnetAddress {
			params.NativePrecompiles[i] = testnetAddress
		}
	}
	return params
}

func updateErc20TokenPairs(chainID string, tokenPairs []erc20types.TokenPair) []erc20types.TokenPair {
	mainnetAddress := erc20types.GetWEVMOSContractHex(utils.MainnetChainID)
	testnetAddress := erc20types.GetWEVMOSContractHex(chainID)

	updatedTokenPairs := make([]erc20types.TokenPair, len(tokenPairs))
	for i, tokerPair := range tokenPairs {
		if tokerPair.Erc20Address == mainnetAddress {
			tp := tokerPair
			tp.Erc20Address = testnetAddress
			updatedTokenPairs[i] = tp
		} else {
			updatedTokenPairs[i] = tokerPair
		}
	}
	return updatedTokenPairs
}
