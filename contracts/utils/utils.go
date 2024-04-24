// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// LoadContract is a helper method to convert the embedded bytes from a JSON file,
// that contain compilation information from Hardhat, into a CompiledContract instance.
func LoadContract(compiledBytes []byte) (evmtypes.CompiledContract, error) {
	var contract evmtypes.HardhatCompiledContract
	err := json.Unmarshal(compiledBytes, &contract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	compiledContract, err := contract.ToCompiledContract()
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(compiledContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, err
	}

	return compiledContract, nil
}
