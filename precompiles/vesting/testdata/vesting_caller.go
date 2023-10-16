// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

var (
	//go:embed VestingCaller.json
	VestingCallerJSON []byte

	// VestingCallerContract is the compiled contract calling the Vesting precompile
	VestingCallerContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(VestingCallerJSON, &VestingCallerContract)
	if err != nil {
		panic(err)
	}

	if len(VestingCallerContract.Bin) == 0 {
		panic("failed to load smart contract that calls the vesting precompile")
	}
}
