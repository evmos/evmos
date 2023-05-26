// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

var addresses = []string{
	"evmos19mqtl7pyvtazl85jlre9jltpuff9enjdn9m7hz",
}

// MigrateStore migrates the x/vesting module state from the consensus version 1 to
// version 2. Specifically, it adds a new store key to enable team vesting accounts subject to
// clawback from governance.
// See Evmos Token Model blog post for details: https://medium.com/evmos/the-evmos-token-model-edc07014978b
func MigrateStore(
	ctx sdk.Context,
	k VestingKeeper,
) error {
	logger := k.Logger(ctx)

	for _, addr := range addresses {
		accAddres := sdk.MustAccAddressFromBech32(addr)
		k.SetGovClawbackEnabled(ctx, accAddres)
		logger.Debug("enabled clawback via governance", "address", addr)
	}

	return nil
}

// VestingKeeper defines the expected keeper for vesting
type VestingKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetGovClawbackEnabled(ctx sdk.Context, address sdk.AccAddress)
}
