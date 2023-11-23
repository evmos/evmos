// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1keeper "github.com/evmos/evmos/v15/x/revenue/v1/keeper"
	v1types "github.com/evmos/evmos/v15/x/revenue/v1/types"
)

// Keeper of this module maintains collections of revenues for contracts
// registered to receive transaction fees.
type Keeper struct {
	*v1keeper.Keeper
}

// NewKeeper creates new instances of the fees Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority sdk.AccAddress,
	bk v1types.BankKeeper,
	dk v1types.DistributionKeeper,
	ak v1types.AccountKeeper,
	evmKeeper v1types.EVMKeeper,
	feeCollector string,
) Keeper {
	k := v1keeper.NewKeeper(storeKey, cdc, authority, bk, dk, ak, evmKeeper, feeCollector)
	return Keeper{
		Keeper: &k,
	}
}
