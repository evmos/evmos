package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var (
	//go:embed ERC20Burnable.json
	erc20BurnableJSON []byte

	// ERC20BurnableContract is the compiled ERC20Burnable contract
	ERC20BurnableContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(erc20BurnableJSON, &ERC20BurnableContract)
	if err != nil {
		panic(err)
	}
}
