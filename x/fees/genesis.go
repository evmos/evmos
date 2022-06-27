package fees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v6/x/fees/keeper"
	"github.com/evmos/evmos/v6/x/fees/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	for _, fee := range data.Fees {
		contract := common.HexToAddress(fee.ContractAddress)
		deployer := sdk.MustAccAddressFromBech32(fee.DeployerAddress)
		withdrawal := sdk.MustAccAddressFromBech32(fee.WithdrawAddress)

		// Set initial contracts receiving transaction fees
		k.SetFee(ctx, contract, deployer, withdrawal)
		k.SetDeployerFees(ctx, deployer, contract)
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Fees:   k.GetFees(ctx),
	}
}
