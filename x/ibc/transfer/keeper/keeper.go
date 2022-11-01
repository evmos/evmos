package keeper

import (
	"github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"

	"github.com/evmos/evmos/v9/x/ibc/transfer/types"
)

// Keeper defines the modified IBC transfer keeper
type Keeper struct {
	*keeper.Keeper
	bankKeeper  types.BankKeeper
	erc20Keeper types.ERC20Keeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(
	transferKeeper keeper.Keeper,
	bankKeeper types.BankKeeper,
	erc20Keeper types.ERC20Keeper,
) Keeper {
	return Keeper{
		Keeper:      &transferKeeper,
		bankKeeper:  bankKeeper,
		erc20Keeper: erc20Keeper,
	}
}
