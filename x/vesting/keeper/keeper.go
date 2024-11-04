// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/evmos/evmos/v20/x/vesting/types"
)

// Keeper of this module maintains collections of vesting.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	evmKeeper          types.EVMKeeper
	stakingKeeper      types.StakingKeeper
	distributionKeeper types.DistributionKeeper
	govKeeper          govkeeper.Keeper

	// The x/gov module account used for executing transaction by governance.
	authority sdk.AccAddress
}

// NewKeeper creates new instances of the vesting Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	authority sdk.AccAddress,
	cdc codec.BinaryCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	dk types.DistributionKeeper,
	ek types.EVMKeeper,
	sk types.StakingKeeper,
	gk govkeeper.Keeper,
) Keeper {
	// ensure gov module account is set and is not nil
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(err)
	}

	return Keeper{
		storeKey:           storeKey,
		authority:          authority,
		cdc:                cdc,
		distributionKeeper: dk,
		accountKeeper:      ak,
		bankKeeper:         bk,
		evmKeeper:          ek,
		stakingKeeper:      sk,
		govKeeper:          gk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
