// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	"github.com/evmos/evmos/v16/x/erc20/types"
)

var (
	//go:embed solidity/ERC20MinterBurnerDecimals.json
	ERC20MinterBurnerDecimalsJSON []byte //nolint: golint

	// ERC20MinterBurnerDecimalsHardhatContract is the compiled erc20 contract
	// generated with hardhat
	ERC20MinterBurnerDecimalsHardhatContract evmtypes.HardhatCompiledContract

	// ERC20MinterBurnerDecimalsContract is the compiled erc20 contract
	ERC20MinterBurnerDecimalsContract evmtypes.CompiledContract

	// ERC20MinterBurnerDecimalsAddress is the erc20 module address
	ERC20MinterBurnerDecimalsAddress common.Address
)

func init() {
	ERC20MinterBurnerDecimalsAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20MinterBurnerDecimalsJSON, &ERC20MinterBurnerDecimalsHardhatContract)
	if err != nil {
		panic(err)
	}

	ERC20MinterBurnerDecimalsContract, err = ERC20MinterBurnerDecimalsHardhatContract.ToCompiledContract()
	if err != nil {
		panic(err)
	}

	if len(ERC20MinterBurnerDecimalsContract.Bin) == 0 {
		panic("load contract failed")
	}
}
