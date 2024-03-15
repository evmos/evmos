// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var (
	//go:embed tokenTransfer.json
	tokenTransferJSON []byte

	// tokenTransferContract is the compiled tokenTransfer contract
	tokenTransferContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(tokenTransferJSON, &tokenTransferContract)
	if err != nil {
		panic(err)
	}
}
