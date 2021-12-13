package contracts

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/erc20/types"
)

var (
	//go:embed ERC20MinterBurner.json
	ERC20BurnableAndMintableJSON []byte // nolint: golint

	// ERC20BurnableAndMintableContract is the compiled erc20 contract
	ERC20BurnableAndMintableContract evmtypes.CompiledContract

	// ERC20BurnableAndMintableAddress is the erc20 module address
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
