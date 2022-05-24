package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var (
	//go:embed compiled_contracts/NoArgs.json
	NoArgsJSON []byte

	// ERC20BurnableContract is the compiled ERC20Burnable contract
	NoArgsContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(NoArgsJSON, &NoArgsContract)
	if err != nil {
		panic(err)
	}
}
