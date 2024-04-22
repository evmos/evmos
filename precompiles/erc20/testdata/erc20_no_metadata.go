// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v17/x/evm/types"

	"github.com/evmos/evmos/v17/x/erc20/types"
)

var (
	//go:embed ERC20NoMetadata.json
	ERC20NoMetadataJSON []byte //nolint: golint

	// ERC20NoMetadataContract is the compiled erc20 contract
	ERC20NoMetadataContract evmtypes.CompiledContract

	// ERC20NoMetadataAddress is the erc20 module address
	ERC20NoMetadataAddress common.Address
)

func init() {
	ERC20NoMetadataAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20NoMetadataJSON, &ERC20NoMetadataContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20NoMetadataContract.Bin) == 0 {
		panic("load contract failed")
	}
}
