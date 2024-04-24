// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	"os"

	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func LoadDistributionCallerContract() (evmtypes.CompiledContract, error) {
	distCallerJSON, err := os.ReadFile("DistributionCaller.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return contractutils.LoadContract(distCallerJSON)
}
