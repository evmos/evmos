package fees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v5/x/fees/keeper"
	"github.com/evmos/evmos/v5/x/fees/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	for _, fee := range data.Fees {
		// prevent storing the same address for deployer and withdrawer
		if fee.DeployerAddress == fee.WithdrawAddress {
			fee.WithdrawAddress = ""
		}

		contract := common.HexToAddress(fee.ContractAddress)
		deployer := sdk.MustAccAddressFromBech32(fee.DeployerAddress)
		withdrawal := sdk.MustAccAddressFromBech32(fee.WithdrawAddress)

		// Set initial contracts receiving transaction fees
		fee := types.NewFee(contract, deployer, withdrawal)
		k.SetFee(ctx, fee)
		k.SetDeployerMap(ctx, deployer, contract)
		if withdrawal != nil {
			k.SetWithdrawMap(ctx, withdrawal, contract)
		}
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Fees:   k.GetFees(ctx),
	}
}
