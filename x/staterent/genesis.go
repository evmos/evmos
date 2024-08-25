// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package staterent

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v19/x/staterent/keeper"
	"github.com/evmos/evmos/v19/x/staterent/types"
)

// InitGenesis initializes the staterent module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(fmt.Errorf("error setting params %s", err))
	}

	for _, v := range data.FlaggedData {
		addr := common.HexToAddress(v.Contract)
		k.SetFlaggedInfo(ctx, addr, v)
	}

}

// ExportGenesis returns the staterent module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	flaggedContracts := []types.FlaggedInfo{}

	k.IterateFlaggedInfo(ctx, func(index int64, info types.FlaggedInfo) (stop bool) {
		flaggedContracts = append(flaggedContracts, info)
		return false
	})

	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		FlaggedData: flaggedContracts,
	}
}
