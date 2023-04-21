// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"

	"github.com/evmos/evmos/v12/x/ibc/transfer/types"
)

// Keeper defines the modified IBC transfer keeper that embeds the original one.
// It also contains the bank keeper and the erc20 keeper to support ERC20 tokens
// to be sent via IBC.
type Keeper struct {
	*keeper.Keeper
	bankKeeper    types.BankKeeper
	erc20Keeper   types.ERC20Keeper
	accountKeeper types.AccountKeeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,

	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper transfertypes.ChannelKeeper,
	portKeeper transfertypes.PortKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	erc20Keeper types.ERC20Keeper,
) Keeper {
	// create the original IBC transfer keeper for embedding
	transferKeeper := keeper.NewKeeper(
		cdc, storeKey, paramSpace,
		ics4Wrapper, channelKeeper, portKeeper,
		accountKeeper, bankKeeper, scopedKeeper,
	)

	return Keeper{
		Keeper:        &transferKeeper,
		bankKeeper:    bankKeeper,
		erc20Keeper:   erc20Keeper,
		accountKeeper: accountKeeper,
	}
}
