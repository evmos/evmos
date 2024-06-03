// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Keeper is a wrapper around the Cosmos SDK staking keeper.
type Keeper struct {
	*stakingkeeper.Keeper
	ak types.AccountKeeper
	bk types.BankKeeper
}

// NewKeeper creates a new staking Keeper wrapper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		stakingkeeper.NewKeeper(cdc, key, ak, bk, authority),
		ak,
		bk,
	}
}
