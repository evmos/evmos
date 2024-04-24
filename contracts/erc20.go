// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

// ERC20MinterBurnerDecimalsContract is the compiled erc20 contract
var ERC20MinterBurnerDecimalsContract evmtypes.CompiledContract

func init() {
	var err error
	ERC20MinterBurnerDecimalsContract, err = contractutils.LoadContractFromJSONFile(
		"solidity/ERC20MinterBurnerDecimals.json",
	)
	if err != nil {
		panic(err)
	}
}
