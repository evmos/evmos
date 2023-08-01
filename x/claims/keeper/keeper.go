// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/evmos/evmos/v14/x/claims/types"
)

// Keeper struct
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
	// the address capable of executing a MsgUpdateParams message. Typically, this should be the x/gov module account.
	authority     sdk.AccAddress
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	distrKeeper   types.DistrKeeper
	channelKeeper types.ChannelKeeper
	ics4Wrapper   porttypes.ICS4Wrapper
}

// NewKeeper returns keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	authority sdk.AccAddress,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	dk types.DistrKeeper,
	ck types.ChannelKeeper,
) *Keeper {
	// ensure gov module account is set and is not nil
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(err)
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		authority:     authority,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
		distrKeeper:   dk,
		channelKeeper: ck,
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
