// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	"os"

	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func LoadERC20AllowanceCaller() (evmtypes.CompiledContract, error) {
	ERC20AllowanceCallerJSON, err := os.ReadFile("ERC20AllowanceCaller.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return contractutils.LoadContract(ERC20AllowanceCallerJSON)
}
