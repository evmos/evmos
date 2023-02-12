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

package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/evmos/evmos/v11/x/vesting/keeper"
	"github.com/evmos/evmos/v11/x/vesting/types"
)

// MigrateStore migrates the x/inflation module state from the consensus version 1 to
// version 2. Specifically, adds all current vesting accounts to the store.
func MigrateStore(ctx sdk.Context, vk keeper.Keeper, ak types.AccountKeeper) error {
	ak.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		if va, ok := account.(exported.VestingAccount); ok {
			vk.AddVestingAccount(ctx, va.GetAddress())
		}
		return false
	})

	return nil
}
