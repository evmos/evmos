// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"errors"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var (
	//go:embed solidity/ERC20Burnable.json
	erc20BurnableJSON []byte
)

func LoadBurnableContract() (evmtypes.CompiledContract, error) {
	// ERC20BurnableHardhatContract is the compiled ERC20Burnable contract
	// generated with hardhat
	var ERC20BurnableHardhatContract evmtypes.HardhatCompiledContract

	err := json.Unmarshal(erc20BurnableJSON, &ERC20BurnableHardhatContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	ERC20BurnableContract, err := ERC20BurnableHardhatContract.ToCompiledContract()
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	print("ERC20BurnableContract.Bin: ", len(ERC20BurnableContract.Bin))
	if len(ERC20BurnableContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("load contract failed")
	}

	return ERC20BurnableContract, nil
}
