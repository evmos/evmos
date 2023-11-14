// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

var (
	//go:embed ERC20Caller.json
	ERC20Caller []byte

	// ERC20CallerContract is the compiled contract calling the staking precompile
	ERC20CallerContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(ERC20Caller, &ERC20CallerContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20CallerContract.Bin) == 0 {
		panic("failed to load smart contract that calls erc20 precompile")
	}
}
