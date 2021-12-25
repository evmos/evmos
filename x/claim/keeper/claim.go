package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/tharsis/evmos/x/claim/types"
)

// GetModuleAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress {
	return k.accountKeeper.GetModuleAddress(types.ModuleName)
}

// GetModuleAccountBalance gets the airdrop coin balance of module account
func (k Keeper) GetModuleAccountBalance(ctx sdk.Context) sdk.Coin {
	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	params := k.GetParams(ctx)
	return k.bankKeeper.GetBalance(ctx, moduleAccAddr, params.ClaimDenom)
}

// SetModuleAccountBalance set balance of airdrop module
func (k Keeper) CreateModuleAccount(ctx sdk.Context, amount sdk.Coin) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter)
	k.accountKeeper.SetModuleAccount(ctx, moduleAcc)

	mintCoins := sdk.NewCoins(amount)

	existingModuleAcctBalance := k.bankKeeper.GetBalance(ctx,
		k.accountKeeper.GetModuleAddress(types.ModuleName), amount.Denom)
	if existingModuleAcctBalance.IsPositive() {
		actual := existingModuleAcctBalance.Add(amount)
		ctx.Logger().Error(fmt.Sprintf("WARNING! There is a bug in claims on InitGenesis, that you are subject to."+
			" You likely expect the claims module account balance to be %d %s, but it will actually be %d %s due to this bug.",
			amount.Amount.Int64(), amount.Denom, actual.Amount.Int64(), actual.Denom))
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoins); err != nil {
		panic(err)
	}
}

func (k Keeper) EndAirdrop(ctx sdk.Context) error {
	ctx.Logger().Info("Beginning EndAirdrop logic")
	err := k.clawbackUnclaimedCoins(ctx)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Clearing claims module state entries")
	k.clearInitialClaimables(ctx)

	err = k.ClawbackAirdrop(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ClawbackAirdrop implements prop 32 by clawing back all the OSMO and IONs from airdrop
// recipient accounts with a sequence number of 0
func (k Keeper) ClawbackAirdrop(ctx sdk.Context) error {
	totalClawback := sdk.NewCoins()
	for _, bechAddr := range types.AirdropAddrs {
		addr, err := sdk.AccAddressFromBech32(bechAddr)
		if err != nil {
			return err
		}

		acc := k.accountKeeper.GetAccount(ctx, addr)
		// if account is nil, do nothing.
		if acc == nil {
			continue
		}
		seq, err := k.accountKeeper.GetSequence(ctx, addr)
		if err != nil {
			return err
		}
		// When sequence number is 0, _and_ from an airdrop account,
		// clawback all the uosmo and uion to community pool.
		// There is an edge case here, where if the address has otherwise been sent uosmo or uion
		// but never made any tx, that excess sent would also get sent to the community pool.
		// This is viewed as an edge case, that the text of Osmosis proposal 32 indicates should
		// be done via sending these excess funds to the community pool.
		//
		// Proposal text to go off of: https://www.mintscan.io/osmosis/proposals/32
		// ***Reminder***
		// 'Unclaimed' tokens are defined as being in wallets which have a Sequence Number = 0,
		// which means the address has NOT performed a single action during the 6 month airdrop claim window.
		// ******CLAWBACK PROPOSED FRAMEWORK******
		// TLDR -- Send ALL unclaimed ION & OSMO back to the community pool
		// and prune those inactive wallets from current state.
		if seq == 0 {
			osmoBal := k.bankKeeper.GetBalance(ctx, addr, "uosmo")
			ionBal := k.bankKeeper.GetBalance(ctx, addr, "uion")
			clawbackCoins := sdk.NewCoins(osmoBal, ionBal)
			totalClawback = totalClawback.Add(clawbackCoins...)
			err = k.distrKeeper.FundCommunityPool(ctx, clawbackCoins, addr)
			if err != nil {
				return err
			}
		}
	}
	ctx.Logger().Info(fmt.Sprintf("clawed back %d uion into community pool", totalClawback.AmountOf("uion").Int64()))
	ctx.Logger().Info(fmt.Sprintf("clawed back %d uosmo into community pool", totalClawback.AmountOf("uosmo").Int64()))
	return nil
}

// ClearClaimables clear claimable amounts
func (k Keeper) clearInitialClaimables(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.ClaimRecordsStorePrefix))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

// SetClaimables set claimable amount from balances object
func (k Keeper) SetClaimRecords(ctx sdk.Context, claimRecords []types.ClaimRecord) error {
	for _, claimRecord := range claimRecords {
		addr, _ := sdk.AccAddressFromBech32(claimRecord.Address)
		k.SetClaimRecord(ctx, addr, claimRecord)
	}
}

// GetClaimables get claimables for genesis export
func (k Keeper) GetClaimRecords(ctx sdk.Context) []types.ClaimRecord {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ClaimRecordsStorePrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	claimRecords := []types.ClaimRecord{}
	for ; iterator.Valid(); iterator.Next() {

		var claimRecord types.ClaimRecord
		k.cdc.MustUnmarshal(iterator.Value(), &claimRecord)
		claimRecords = append(claimRecords, claimRecord)
	}

	return claimRecords
}

// GetClaimRecord returns the claim record for a specific address
func (k Keeper) GetClaimRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ClaimRecordsStorePrefix)

	bz := store.Get(addr)
	if len(bz) == 0 {
		return types.ClaimRecord{}, false
	}

	var claimRecord types.ClaimRecord
	k.cdc.MustUnmarshal(bz, &claimRecord)

	return claimRecord, true
}

// SetClaimRecord sets a claim record for an address in store
func (k Keeper) SetClaimRecord(ctx sdk.Context, addr sdk.AccAddress, claimRecord types.ClaimRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ClaimRecordsStorePrefix)
	bz := k.cdc.MustMarshal(&claimRecord)
	store.Set(addr, bz)
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetClaimableAmountForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error) {
	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrUnexpectedEndOfGroupClaim, addr.String())
	}

	// if action already completed, nothing is claimable
	if claimRecord.ActionsCompleted[action] {
		return sdk.Coins{}, nil
	}

	params := k.GetParams(ctx)

	// If we are before the start time, do nothing.
	// This case _shouldn't_ occur on chain, since the
	// start time ought to be chain start time.
	if ctx.BlockTime().Before(params.AirdropStartTime) {
		return sdk.Coins{}, nil
	}

	initialClaimablePerAction := sdk.Coins{}

	// FIXME: update this and explicitely define the %
	for _, coin := range claimRecord.InitialClaimableAmount {
		initialClaimablePerAction = initialClaimablePerAction.Add(
			sdk.NewCoin(coin.Denom,
				coin.Amount.QuoRaw(int64(len(types.Action_name))),
			),
		)
	}

	elapsedAirdropTime := ctx.BlockTime().Sub(params.AirdropStartTime)
	// Are we early enough in the airdrop s.t. theres no decay?
	if elapsedAirdropTime <= params.DurationUntilDecay {
		return initialClaimablePerAction, nil
	}

	// The entire airdrop has completed
	if elapsedAirdropTime > params.DurationUntilDecay+params.DurationOfDecay {
		return sdk.Coins{}, nil
	}

	// Positive, since goneTime > params.DurationUntilDecay
	decayTime := elapsedAirdropTime - params.DurationUntilDecay
	decayPercent := sdk.NewDec(decayTime.Nanoseconds()).QuoInt64(params.DurationOfDecay.Nanoseconds())
	claimablePercent := sdk.OneDec().Sub(decayPercent)

	claimableCoins := sdk.Coins{}
	// TODO: define claimable percent
	for _, coin := range initialClaimablePerAction {
		claimableCoins = claimableCoins.Add(sdk.NewCoin(coin.Denom, coin.Amount.ToDec().Mul(claimablePercent).RoundInt()))
	}

	return claimableCoins, nil
}

// GetClaimable returns claimable amount for a specific action done by an address
func (k Keeper) GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	totalClaimable := sdk.Coins{}

	// FIXME: don't iterate over maps!
	for action := range types.Action_name {
		claimableForAction, err := k.GetClaimableAmountForAction(ctx, addr, types.Action(action))
		if err != nil {
			return sdk.Coins{}, err
		}
		totalClaimable = totalClaimable.Add(claimableForAction...)
	}
	return totalClaimable, nil
}

// ClaimCoins remove claimable amount entry and transfer it to user's account
func (k Keeper) ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error) {
	claimableAmount, err := k.GetClaimableAmountForAction(ctx, addr, action)
	if err != nil {
		return nil, err
	}

	if claimableAmount.Empty() {
		return claimableAmount, nil
	}

	claimRecord, found := k.GetClaimRecord(ctx, addr)
	if !found {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, claimableAmount)
	if err != nil {
		return nil, err
	}

	claimRecord.ActionCompleted[action] = true

	k.SetClaimRecord(ctx, addr, claimRecord)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(sdk.AttributeKeySender, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, claimableAmount.String()),
		),
	})

	return claimableAmount, nil
}

// ClawbackUnclaimedCoins transfer remaining coins to the community pool treasury when airdrop period ends
func (k Keeper) clawbackUnclaimedCoins(ctx sdk.Context) error {
	moduleAccAddr := k.GetModuleAccountAddress(ctx)
	amt := k.GetModuleAccountBalance(ctx)
	k.Logger(ctx).Info(
		"clawback of funds to community pool treasury",
		"total", amt.Amount.String(),
	)
	return k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(amt), moduleAccAddr)
}
