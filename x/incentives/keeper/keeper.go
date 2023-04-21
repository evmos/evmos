// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/evmos/evmos/v12/x/incentives/types"
)

// Keeper of this module maintains collections of incentives.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	// the address capable of executing a MsgUpdateParams message. Typically, this should be the x/gov module account.
	authority sdk.AccAddress

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
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority sdk.AccAddress,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ik types.InflationKeeper,
	sk types.StakeKeeper,
	evmKeeper types.EVMKeeper,
) Keeper {
	// ensure gov module account is set and is not nil
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(err)
	}

	return Keeper{
		storeKey:        storeKey,
		cdc:             cdc,
		authority:       authority,
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
