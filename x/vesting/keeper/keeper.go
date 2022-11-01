package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/evmos/evmos/v10/x/vesting/types"
)

// Keeper of this module maintains collections of vesting.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates new instances of the vesting Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
