package keeper

import (
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

// Keeper defines the IBC fungible transfer hooks keeper
type Keeper struct {
	storeKey   sdk.StoreKey
	paramstore paramtypes.Subspace

	transferKeeper types.TransferKeeper
	hooks          types.TransferHooks
}

// NewKeeper creates a new IBC transfer hooks Keeper instance
func NewKeeper(
	key sdk.StoreKey,
	ps paramtypes.Subspace,
	tk types.TransferKeeper) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:       key,
		paramstore:     ps,
		transferKeeper: tk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"-"+types.ModuleName)
}

func (k *Keeper) SetHooks(th types.TransferHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set hooks twice")
	}

	k.hooks = th

	return k
}

// IsTransferHooks returns whether transfer hooks logic should be run.
func (k Keeper) IsTransferHooksEnabled(ctx sdk.Context) bool {
	if k.hooks == nil {
		return false
	}

	return k.GetParams(ctx).EnableTransferHook
}

func (k Keeper) DenomPathFromHash(ctx sdk.Context, denom string) (string, error) {
	return k.transferKeeper.DenomPathFromHash(ctx, denom)
}
