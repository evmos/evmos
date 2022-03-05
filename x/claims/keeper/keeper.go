package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

// Keeper struct
type Keeper struct {
	cdc           codec.Codec
	storeKey      sdk.StoreKey
	paramstore    paramtypes.Subspace
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	distrKeeper   types.DistrKeeper
	ics4Wrapper   transfertypes.ICS4Wrapper
}

// NewKeeper returns keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	dk types.DistrKeeper,
	ics4Wrapper transfertypes.ICS4Wrapper,
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
		ics4Wrapper:   ics4Wrapper,
	}
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetModuleAccountAccount returns the module account for the claim module
func (k Keeper) GetModuleAccountAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetModuleAccountAddress gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

// GetModuleAccountBalances gets the balances of module account that escrows the
// airdrop tokens
func (k Keeper) GetModuleAccountBalances(ctx sdk.Context) sdk.Coins {
	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	return k.bankKeeper.GetAllBalances(ctx, moduleAccAddr)
}

// GetClaimsRecord returns the claim record for a specific address
func (k Keeper) GetClaimsRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimsRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)

	bz := store.Get(addr)
	if len(bz) == 0 {
		return types.ClaimsRecord{}, false
	}

	var claimsRecord types.ClaimsRecord
	k.cdc.MustUnmarshal(bz, &claimsRecord)

	return claimsRecord, true
}

// SetClaimsRecord sets a claim record for an address in store
func (k Keeper) SetClaimsRecord(ctx sdk.Context, addr sdk.AccAddress, claimsRecord types.ClaimsRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)
	bz := k.cdc.MustMarshal(&claimsRecord)
	store.Set(addr, bz)
}

// DeleteClaimsRecord deletes a claim record from the store
func (k Keeper) DeleteClaimsRecord(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)
	store.Delete(addr)
}

func (k Keeper) IterateClaimsRecords(ctx sdk.Context, handlerFn func(addr sdk.AccAddress, cr types.ClaimsRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixClaimsRecords)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var claimsRecord types.ClaimsRecord
		k.cdc.MustUnmarshal(iterator.Value(), &claimsRecord)

		addr := sdk.AccAddress(iterator.Key()[1:])
		cr := types.ClaimsRecord{
			InitialClaimableAmount: claimsRecord.InitialClaimableAmount,
			ActionsCompleted:       claimsRecord.ActionsCompleted,
		}

		if handlerFn(addr, cr) {
			break
		}
	}
}

// GetClaimsRecords get claimables for genesis export
func (k Keeper) GetClaimsRecords(ctx sdk.Context) []types.ClaimsRecordAddress {
	claimsRecords := []types.ClaimsRecordAddress{}
	k.IterateClaimsRecords(ctx, func(addr sdk.AccAddress, cr types.ClaimsRecord) (stop bool) {
		cra := types.ClaimsRecordAddress{
			Address:                addr.String(),
			InitialClaimableAmount: cr.InitialClaimableAmount,
			ActionsCompleted:       cr.ActionsCompleted,
		}

		claimsRecords = append(claimsRecords, cra)
		return false
	})

	return claimsRecords
}
