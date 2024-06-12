// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

var (
	//go:embed compiled_contracts/TokenFactoryCoin.json
	TOKENFACTORYCOINJSON []byte

	// TokenFactoryCoinContract is the compiled contract of Token Factory Coin
	TokenFactoryCoinContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(TOKENFACTORYCOINJSON, &TokenFactoryCoinContract)
	if err != nil {
		panic(err)
	}

	if len(TokenFactoryCoinContract.Bin) == 0 {
		panic("failed to load TokenFactoryCoin smart contract")
	}
}
