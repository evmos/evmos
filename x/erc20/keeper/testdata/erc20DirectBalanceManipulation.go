// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	"os"

	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// LoadBalanceManipulationContract loads the ERC20DirectBalanceManipulation contract
// from the compiled JSON data.
//
// This is an evil token. Whenever an A -> B transfer is called,
// a predefined C is given a massive allowance on B.
func LoadBalanceManipulationContract() (evmtypes.CompiledContract, error) {
	contractJSON, err := os.ReadFile("ERC20DirectBalanceManipulation.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return contractutils.LoadContract(contractJSON)
}
