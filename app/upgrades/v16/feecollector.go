// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MigrateFeeCollector migrates the fee collector module account to include the `Burner` permission.
func MigrateFeeCollector(ak authkeeper.AccountKeeper, ctx sdk.Context) error {
	feeCollectorModuleAccount := ak.GetModuleAccount(ctx, types.FeeCollectorName)
	if feeCollectorModuleAccount == nil {
		return fmt.Errorf("fee collector module account not found")
	}

	modAcc, ok := feeCollectorModuleAccount.(*types.ModuleAccount)
	if !ok {
		return fmt.Errorf("fee collector module account is not a module account")
	}

	// Create a new FeeCollector module account with the same address and balance as the old one.
	newFeeCollectorModuleAccount := types.NewModuleAccount(modAcc.BaseAccount, types.FeeCollectorName, types.Burner)

	// Override the FeeCollector module account in the auth keeper.
	ak.SetModuleAccount(ctx, newFeeCollectorModuleAccount)

	return nil
}
