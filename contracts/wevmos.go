// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

var (
	//go:embed compiled_contracts/WEVMOS.json
	WEVMOSJSON []byte

	// WEVMOSContract is the compiled contract of WEVMOS
	WEVMOSContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(WEVMOSJSON, &WEVMOSContract)
	if err != nil {
		panic(err)
	}

	if len(WEVMOSContract.Bin) == 0 {
		panic("failed to load WEVMOS smart contract")
	}
}
