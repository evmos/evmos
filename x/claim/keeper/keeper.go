package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/tharsis/evmos/x/claim/types"
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

// GetModuleAccountBalances gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

// GetModuleAccountBalances gets the balances of module account that escrows the
// airdrop tokens
func (k Keeper) GetModuleAccountBalances(ctx sdk.Context) sdk.Coins {
	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	return k.bankKeeper.GetAllBalances(ctx, moduleAccAddr)
}

// GetClaimRecord returns the claim record for a specific address
func (k Keeper) GetClaimRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimRecords)

	bz := store.Get(addr)
	if len(bz) == 0 {
		return types.ClaimRecord{}, false
	}

	var claimRecord types.ClaimRecord
	k.cdc.MustUnmarshal(bz, &claimRecord)

	return claimRecord, true
}

// HasClaimRecord returns true if the address already has a claim record.
func (k Keeper) HasClaimRecord(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimRecords)
	return store.Has(addr)
}

// SetClaimRecord sets a claim record for an address in store
func (k Keeper) SetClaimRecord(ctx sdk.Context, addr sdk.AccAddress, claimRecord types.ClaimRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimRecords)
	bz := k.cdc.MustMarshal(&claimRecord)
	store.Set(addr, bz)
}

// DeleteClaimRecord deletes a claim record from the store
func (k Keeper) DeleteClaimRecord(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimRecords)
	store.Delete(addr)
}

func (k Keeper) IterateClaimRecords(ctx sdk.Context, handlerFn func(addr sdk.AccAddress, cr types.ClaimRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixClaimRecords)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var claimRecord types.ClaimRecord
		k.cdc.MustUnmarshal(iterator.Value(), &claimRecord)

		addr := sdk.AccAddress(iterator.Key()[1:])
		cr := types.ClaimRecord{
			InitialClaimableAmount: claimRecord.InitialClaimableAmount,
			ActionsCompleted:       claimRecord.ActionsCompleted,
		}

		if handlerFn(addr, cr) {
			break
		}
	}
}

// GetClaimables get claimables for genesis export
func (k Keeper) GetClaimRecords(ctx sdk.Context) []types.ClaimRecordAddress {
	claimRecords := []types.ClaimRecordAddress{}
	k.IterateClaimRecords(ctx, func(addr sdk.AccAddress, cr types.ClaimRecord) (stop bool) {
		cra := types.ClaimRecordAddress{
			Address:                addr.String(),
			InitialClaimableAmount: cr.InitialClaimableAmount,
			ActionsCompleted:       cr.ActionsCompleted,
		}

		claimRecords = append(claimRecords, cra)
		return false
	})

	return claimRecords
}
