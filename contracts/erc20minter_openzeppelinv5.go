// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	"github.com/evmos/evmos/v15/x/erc20/types"
)

var (
	//go:embed compiled_contracts/ERC20Minter_OpenZeppelinV5.json
	ERC20MinterJSON []byte //nolint: golint

	// ERC20MinterContract is the compiled erc20 contract
	ERC20MinterContract evmtypes.CompiledContract

	// ERC20MinterAddress is the erc20 module address
	ERC20MinterAddress common.Address
)

func init() {
	ERC20MinterAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20MinterJSON, &ERC20MinterContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20MinterContract.Bin) == 0 {
		panic("load contract failed")
	}
}
