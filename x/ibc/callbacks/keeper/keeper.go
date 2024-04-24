package keeper

import (
	"github.com/cosmos/ibc-go/modules/apps/callbacks/types"
)

var _ types.ContractKeeper = Keeper{}

// Keeper defines the modified IBC transfer keeper that embeds the original one.
// It also contains the bank keeper and the erc20 keeper to support ERC20 tokens
// to be sent via IBC.
type Keeper struct {
}
