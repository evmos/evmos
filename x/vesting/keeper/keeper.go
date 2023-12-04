// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v16/x/vesting/types"
)

// Keeper of this module maintains collections of vesting.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	stakingKeeper      types.StakingKeeper
	distributionKeeper types.DistributionKeeper
	govKeeper          types.GovKeeper

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
	sk types.StakingKeeper,
	gk types.GovKeeper,
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
		stakingKeeper:      sk,
		govKeeper:          gk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
