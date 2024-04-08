// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

var (
	//go:embed compiled_contracts/FlashLoan.json
	FlashLoanJSON []byte //nolint:golint // used to embed the compiled contract

	// FlashLoanContract is the compiled flash loan contract
	FlashLoanContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(FlashLoanJSON, &FlashLoanContract)
	if err != nil {
		panic(err)
	}

	if len(FlashLoanContract.Bin) == 0 {
		panic("load contract failed")
	}
}
