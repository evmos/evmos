// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"errors"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// This is an evil token. Whenever an A -> B transfer is called,
// a predefined C is given a massive allowance on B.
var (
	//go:embed solidity/ERC20MaliciousDelayed.json
	ERC20MaliciousDelayedJSON []byte //nolint: golint
)

func LoadMaliciousDelayedContract() (evmtypes.CompiledContract, error) {
	// ERC20MaliciousDelayedHardhatContract is the compiled erc20 contract
	// generated with hardhat
	var ERC20MaliciousDelayedHardhatContract evmtypes.HardhatCompiledContract

	err := json.Unmarshal(ERC20MaliciousDelayedJSON, &ERC20MaliciousDelayedHardhatContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	ERC20MaliciousDelayedContract, err := ERC20MaliciousDelayedHardhatContract.ToCompiledContract()
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(ERC20MaliciousDelayedContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("load contract failed")
	}

	return ERC20MaliciousDelayedContract, nil
}
