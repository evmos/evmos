// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"errors"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// This is an evil token. Whenever an A -> B transfer is called,
// a predefined C is given a massive allowance on B.
var (
	//go:embed ERC20DirectBalanceManipulation.json
	ERC20DirectBalanceManipulationJSON []byte //nolint: golint
)

// LoadBalanceManipulationContract loads the ERC20DirectBalanceManipulation contract
// from the compiled JSON data.
func LoadBalanceManipulationContract() (evmtypes.CompiledContract, error) {
	// ERC20DirectBalanceManipulationHardhatContract is the compiled erc20 contract
	// generated with hardhat
	var ERC20DirectBalanceManipulationHardhatContract evmtypes.HardhatCompiledContract

	err := json.Unmarshal(ERC20DirectBalanceManipulationJSON, &ERC20DirectBalanceManipulationHardhatContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	ERC20DirectBalanceManipulationContract, err := ERC20DirectBalanceManipulationHardhatContract.ToCompiledContract()
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(ERC20DirectBalanceManipulationContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("load contract failed")
	}

	return ERC20DirectBalanceManipulationContract, nil
}
