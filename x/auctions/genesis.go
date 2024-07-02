// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/auctions/keeper"
	"github.com/evmos/evmos/v18/x/auctions/types"
)

func InitGenesis(
	ctx sdk.Context, k keeper.Keeper, data types.GenesisState,
) {
}

func ExportGenesis(
	ctx sdk.Context, k keeper.Keeper,
) *types.GenesisState {
	return &types.GenesisState{}
}
