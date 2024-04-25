package keeper

import (
	"bytes"
	"embed"
	"github.com/cosmos/ibc-go/modules/apps/callbacks/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	evmkeeper "github.com/evmos/evmos/v17/x/evm/keeper"
)

var _ types.ContractKeeper = Keeper{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Keeper defines the modified IBC transfer keeper that embeds the original one.
// It also contains the bank keeper and the erc20 keeper to support ERC20 tokens
// to be sent via IBC.
type Keeper struct {
	evmKeeper *evmkeeper.Keeper
	abi.ABI
}

func NewKeeper(
	evmKeeper *evmkeeper.Keeper,
) Keeper {
	abiBz, err := f.ReadFile("abi.json")
	if err != nil {
		panic(err)
	}

	newAbi, err := abi.JSON(bytes.NewReader(abiBz))
	if err != nil {
		panic(err)
	}

	return Keeper{
		ABI:       newAbi,
		evmKeeper: evmKeeper,
	}
}
