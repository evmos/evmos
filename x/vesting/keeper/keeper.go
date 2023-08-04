// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/evmos/evmos/v14/x/vesting/types"
)

// Keeper of this module maintains collections of vesting.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	stakingKeeper      types.StakingKeeper
	distributionKeeper types.DistributionKeeper

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
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HasGovClawbackEnabled checks if the given account has governance clawback enabled.
func (k Keeper) HasGovClawbackEnabled(ctx sdk.Context, addr sdk.AccAddress) bool {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackEnabledKey, addr.Bytes()...)
	return ctx.KVStore(k.storeKey).Has(key)
}

// SetGovClawbackEnabled enables the given vesting account address to be clawed back
// via governance.
func (k Keeper) SetGovClawbackEnabled(ctx sdk.Context, addr sdk.AccAddress) {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackEnabledKey, addr.Bytes()...)
	ctx.KVStore(k.storeKey).Set(key, []byte{0x01})
}

// DeleteGovClawbackEnabled disables the given vesting account address to be clawed back
// via governance.
func (k Keeper) DeleteGovClawbackEnabled(ctx sdk.Context, addr sdk.AccAddress) {
	//nolint:gocritic
	key := append(types.KeyPrefixGovClawbackEnabledKey, addr.Bytes()...)
	ctx.KVStore(k.storeKey).Delete(key)
}
