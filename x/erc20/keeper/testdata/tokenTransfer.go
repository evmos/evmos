// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

var (
	//go:embed tokenTransfer.json
	tokenTransferJSON []byte

	// TokenTransferContract is the compiled tokenTransfer contract
	TokenTransferContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(tokenTransferJSON, &TokenTransferContract)
	if err != nil {
		panic(err)
	}
}
