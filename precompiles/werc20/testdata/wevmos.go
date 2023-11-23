// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

var (
	//go:embed WEVMOS.json
	WevmosJSON []byte

	// WEVMOSContract is the compiled contract of WEVMOS
	WEVMOSContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(WevmosJSON, &WEVMOSContract)
	if err != nil {
		panic(err)
	}

	if len(WEVMOSContract.Bin) == 0 {
		panic("failed to load WEVMOS smart contract")
	}
}
