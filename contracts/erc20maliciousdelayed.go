package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/v4/x/erc20/types"
)

// This is an evil token. Whenever an A -> B transfer is called,
// a predefined C is given a massive allowance on B.
var (
	//go:embed compiled_contracts/ERC20MaliciousDelayed.json
	ERC20MaliciousDelayedJSON []byte // nolint: golint

	// ERC20MaliciousDelayedContract is the compiled erc20 contract
	ERC20MaliciousDelayedContract evmtypes.CompiledContract

	// ERC20MaliciousDelayedAddress is the erc20 module address
	ERC20MaliciousDelayedAddress common.Address
)

func init() {
	ERC20MaliciousDelayedAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20MaliciousDelayedJSON, &ERC20MaliciousDelayedContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20MaliciousDelayedContract.Bin) == 0 {
		panic("load contract failed")
	}
}
