package fees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/x/fees/keeper"
	"github.com/tharsis/evmos/x/fees/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	for _, cwa := range data.WithdrawAddresses {
		contractAddr := common.HexToAddress(cwa.ContractAddress)
		withdrawAddr := common.HexToAddress(cwa.WithdrawAddress)
		k.SetContractWithdrawAddress(ctx, contractAddr, withdrawAddr)
		k.SetContractWithdrawAddressInverse(ctx, contractAddr, withdrawAddr)
	}
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		WithdrawAddresses: k.GetContractWithdrawAddresses(ctx),
	}
}
