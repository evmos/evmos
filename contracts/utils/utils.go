// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"

	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// LoadContractFromJSONFile is a helper method to convert the embedded bytes from a JSON file,
// that contain compilation information from Hardhat, into a CompiledContract instance.
func LoadContractFromJSONFile(jsonFile string) (evmtypes.CompiledContract, error) {
	compiledBytes, err := loadCompiledBytesFromJSONFile(jsonFile)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return ConvertHardhatBytesToCompiledContract(compiledBytes)
}

// LegacyLoadContractFromJSONFile is a helper method to convert the embedded bytes from a JSON file,
// that contain compilation information, into a CompiledContract instance.
//
// NOTE: This is used for contracts that were compiled manually and not using the current Hardhat setup.
func LegacyLoadContractFromJSONFile(jsonFile string) (evmtypes.CompiledContract, error) {
	compiledBytes, err := loadCompiledBytesFromJSONFile(jsonFile)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	var contract evmtypes.CompiledContract
	err = json.Unmarshal(compiledBytes, &contract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(contract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("got empty binary data for contract")
	}

	return contract, nil
}

// loadCompiledBytesFromJSONFile is a helper method to load the embedded bytes from a JSON file.
// It takes in a file path that's relative to where this function is called,
// similar to how go:embed would be used.
func loadCompiledBytesFromJSONFile(jsonFile string) ([]byte, error) {
	// We need to get the directory of the caller to load
	// the JSON file relative to where the function is called.
	//
	// The caller of interest is 2 levels up the stack as this
	// method is being called in the functions above.
	_, caller, _, ok := runtime.Caller(2)
	if !ok {
		return nil, errors.New("could not get the caller")
	}

	callerDir := filepath.Dir(caller)
	compiledBytes, err := os.ReadFile(filepath.Join(callerDir, jsonFile))
	if err != nil {
		return nil, err
	}

	return compiledBytes, nil
}

// ConvertHardhatBytesToCompiledContract is a helper method to convert the embedded bytes from a
// Hardhat JSON file into an instance of the CompiledContract type.
func ConvertHardhatBytesToCompiledContract(bz []byte) (evmtypes.CompiledContract, error) {
	var hardhatContract evmtypes.HardhatCompiledContract
	err := json.Unmarshal(bz, &hardhatContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	compiledContract, err := hardhatContract.ToCompiledContract()
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(compiledContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("got empty binary data for contract")
	}

	return compiledContract, nil
}
