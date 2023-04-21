// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/evmos/evmos/v12/x/inflation/types"
)

// Keeper of the inflation store
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	// the address capable of executing a MsgUpdateParams message. Typically, this should be the x/gov module account.
	authority        sdk.AccAddress
	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      types.DistrKeeper
	stakingKeeper    types.StakingKeeper
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority sdk.AccAddress,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	dk types.DistrKeeper,
	sk types.StakingKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// ensure gov module account is set and is not nil
	if err := sdk.VerifyAddressFormat(authority); err != nil {
		panic(err)
	}

	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		authority:        authority,
		accountKeeper:    ak,
		bankKeeper:       bk,
		distrKeeper:      dk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
