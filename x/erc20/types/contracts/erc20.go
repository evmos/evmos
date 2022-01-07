package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/x/erc20/types"
)

var (
	//go:embed ERC20PresetMinterPauserDecimal.json
	ERC20PresetMinterPauserDecimalJSON []byte // nolint: golint

	// ERC20PresetMinterPauserDecimalContract is the compiled erc20 contract
	ERC20PresetMinterPauserDecimalContract evmtypes.CompiledContract

	// ERC20PresetMinterPauserDecimalAddress is the erc20 module address
	ERC20PresetMinterPauserDecimalAddress common.Address
)

func init() {
	ERC20PresetMinterPauserDecimalAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20PresetMinterPauserDecimalJSON, &ERC20PresetMinterPauserDecimalContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20PresetMinterPauserDecimalContract.Bin) == 0 {
		panic("load contract failed")
	}
}
