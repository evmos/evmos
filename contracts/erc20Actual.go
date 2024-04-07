// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var (
	//go:embed compiled_contracts/ActualERC20MinterBurnerDecimals.json
	ActualERC20MinterBurnerDecimalsJSON []byte //nolint: golint

	// ActualERC20MinterBurnerDecimalsContract is the compiled erc20 contract derived from Hardhat JSON
	ActualERC20MinterBurnerDecimalsHardhatContract evmtypes.HardhatCompiledContract

	// ActualERC20MinterBurnerDecimalsContract is the compiled erc20 contract
	ActualERC20MinterBurnerDecimalsContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(ActualERC20MinterBurnerDecimalsJSON, &ActualERC20MinterBurnerDecimalsHardhatContract)
	if err != nil {
		panic(err)
	}

	// TODO: remove this whole file when done and replace the known ERC-20 minter files
	ActualERC20MinterBurnerDecimalsContract, err := ActualERC20MinterBurnerDecimalsHardhatContract.ToCompiledContract()
	if err != nil {
		panic(err)
	}

	if len(ActualERC20MinterBurnerDecimalsContract.Bin) == 0 {
		panic("load contract failed")
	}
}
