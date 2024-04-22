// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"errors"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

//go:embed FlashLoan.json
var FlashLoanJSON []byte //nolint:golint // used to embed the compiled contract

func LoadFlashLoanContract() (evmtypes.CompiledContract, error) {
	// FlashLoanHardhatContract is the compiled flash loan contract
	// generated by hardhat
	var FlashLoanHardhatContract evmtypes.HardhatCompiledContract

	err := json.Unmarshal(FlashLoanJSON, &FlashLoanHardhatContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	FlashLoanContract, err := FlashLoanHardhatContract.ToCompiledContract()
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(FlashLoanContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("load contract failed")
	}

	return FlashLoanContract, nil
}
