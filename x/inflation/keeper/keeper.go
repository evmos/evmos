package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/tharsis/evmos/x/inflation/types"
)

// Keeper of the inflation store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      types.DistrKeeper
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	dk types.DistrKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		paramstore:       ps,
		accountKeeper:    ak,
		bankKeeper:       bk,
		distrKeeper:      dk,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// CreateDeveloperVestingModuleAccount mints and sends coins to the unvested
// team account.
func (k Keeper) MintGenesisTeamVestingCoins(
	ctx sdk.Context,
	amount sdk.Coins,
) error {
	// Mint coins to inflation account
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		return err
	}

	// Send coins from inflation module mto unvested team module account
	return k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		types.UnvestedTeamAccount,
		amount,
	)

}
