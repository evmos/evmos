// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	"os"

	contractutils "github.com/evmos/evmos/v16/contracts/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func LoadERC20MinterV5Contract() (evmtypes.CompiledContract, error) {
	erc20MinterV5JSON, err := os.ReadFile("ERC20Minter_OpenZeppelinV5.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	return contractutils.LoadContract(erc20MinterV5JSON)
}
