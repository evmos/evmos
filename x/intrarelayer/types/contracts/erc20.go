package contracts

import (
	_ "embed"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var (
	//go:embed ERC20PresentMinterPauser.json
	ERC20BurnableAndMintableJSON []byte

	// ERC20BurnableAndMintableContract is the compiled erc20 contract
	ERC20BurnableAndMintableContract CompiledContract

	// ERC20BurnableAndMintableAddress is the irm module address
	ERC20BurnableAndMintableAddress common.Address
)

func init() {
	ERC20BurnableAndMintableAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20BurnableAndMintableJSON, &ERC20BurnableAndMintableContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20BurnableAndMintableContract.Bin) == 0 {
		panic("load contract failed")
	}
}
