package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"

	"github.com/evmos/evmos/v10/x/claims/types"
)

// Keeper struct
type Keeper struct {
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
	paramstore    paramtypes.Subspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	distrKeeper   types.DistrKeeper
	ics4Wrapper   porttypes.ICS4Wrapper
}

// NewKeeper returns keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	dk types.DistrKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramstore:    ps,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
		distrKeeper:   dk,
	}
}

// SetICS4Wrapper sets the ICS4 wrapper to the keeper.
// It panics if already set
func (k *Keeper) SetICS4Wrapper(ics4Wrapper porttypes.ICS4Wrapper) {
	if k.ics4Wrapper != nil {
		panic("ICS4 wrapper already set")
	}

	k.ics4Wrapper = ics4Wrapper
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetModuleAccount returns the module account for the claim module
func (k Keeper) GetModuleAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetModuleAccountAddress gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountAddress() sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

// GetModuleAccountBalances gets the balances of module account that escrows the
// airdrop tokens
func (k Keeper) GetModuleAccountBalances(ctx sdk.Context) sdk.Coins {
	moduleAccAddr := k.GetModuleAccountAddress()
	return k.bankKeeper.GetAllBalances(ctx, moduleAccAddr)
}
