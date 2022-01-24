package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO turn into epoch provision stored as sdk.Dec
func (k Keeper) GetEpochProvision(ctx sdk.Context) {}

func (k Keeper) SetEpochProvision(ctx sdk.Context) {}

// TODO turn into epoch provision stored as int64
func (k Keeper) GetPeriod(ctx sdk.Context) {}

func (k Keeper) SetPeriod(ctx sdk.Context) {}
