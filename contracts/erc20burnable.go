// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v17/x/evm/types"
)

var (
	//go:embed compiled_contracts/ERC20Burnable.json
	erc20BurnableJSON []byte

	// ERC20BurnableContract is the compiled ERC20Burnable contract
	ERC20BurnableContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(erc20BurnableJSON, &ERC20BurnableContract)
	if err != nil {
		panic(err)
	}
}
