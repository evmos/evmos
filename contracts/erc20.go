// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract

	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var (
	//go:embed solidity/ERC20MinterBurnerDecimals.json
	ERC20MinterBurnerDecimalsJSON []byte //nolint: golint

	// ERC20MinterBurnerDecimalsContract is the compiled erc20 contract
	ERC20MinterBurnerDecimalsContract evmtypes.CompiledContract
)

func init() {
	var err error
	ERC20MinterBurnerDecimalsContract, err = contractutils.LoadContract(ERC20MinterBurnerDecimalsJSON)
	if err != nil {
		panic(err)
	}
}
