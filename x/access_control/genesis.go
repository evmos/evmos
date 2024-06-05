// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/x/access_control/keeper"
	"github.com/evmos/evmos/v18/x/access_control/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	// Set contract owners
	for _, owner := range data.ContractOwner {
		ownerAddr := common.HexToAddress(owner.Account)
		contractAddr := common.HexToAddress(owner.Contract)
		k.SetOwner(ctx, contractAddr, ownerAddr)
	}

	// Pause all contracts in the genesis
	for _, contract := range data.PausedContracts {
		contractAddr := common.HexToAddress(contract)
		k.Pause(ctx, contractAddr)
	}

}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	pausedContracts := k.GetPausedContracts(ctx)
	pausedContractsStr := make([]string, len(pausedContracts))
	for i, address := range pausedContracts {
		pausedContractsStr[i] = address.Hex()
	}
	return &types.GenesisState{
		ContractOwner:   k.GetOwners(ctx),
		PausedContracts: pausedContractsStr,
	}
}
