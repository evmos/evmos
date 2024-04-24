// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	"encoding/json"
	"errors"
	"os"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func LoadWEVMOSContract() (evmtypes.CompiledContract, error) {
	wevmosJSON, err := os.ReadFile("WEVMOS.json")
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	var wevmosContract evmtypes.CompiledContract
	err = json.Unmarshal(wevmosJSON, &wevmosContract)
	if err != nil {
		return evmtypes.CompiledContract{}, err
	}

	if len(wevmosContract.Bin) == 0 {
		return evmtypes.CompiledContract{}, errors.New("empty contract binary")
	}

	return wevmosContract, nil
}
