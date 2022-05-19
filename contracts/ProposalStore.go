package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var (
	//go:embed compiled_contracts/ProposalStore.json
	ProposalStoreJSON []byte

	// ERC20BurnableContract is the compiled ERC20Burnable contract
  ProposalStoreContract evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(ProposalStoreJSON, &ProposalStoreContract)
	if err != nil {
		panic(err)
	}
}
