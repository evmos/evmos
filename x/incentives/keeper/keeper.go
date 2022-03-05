package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/tharsis/evmos/v2/x/incentives/types"
)

// Keeper of this module maintains collections of incentives.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	inflationKeeper types.InflationKeeper

	// Currently not used, but added to prevent breaking change s in case we want
	// to allocate incentives to staking instead of transferring the deferred
	// rewards to the user's wallet
	stakeKeeper types.StakeKeeper
	evmKeeper   types.EVMKeeper
}

// NewKeeper creates new instances of the incentives Keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ik types.InflationKeeper,
	sk types.StakeKeeper,
	evmKeeper types.EVMKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:        storeKey,
		cdc:             cdc,
		paramstore:      ps,
		accountKeeper:   ak,
		bankKeeper:      bk,
		inflationKeeper: ik,
		stakeKeeper:     sk,
		evmKeeper:       evmKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
