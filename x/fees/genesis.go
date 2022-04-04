package fees

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v3/x/fees/keeper"
	"github.com/tharsis/evmos/v3/x/fees/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	// Ensure fees module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the fees module account has not been set")
	}

	for _, fee := range data.DevFeeInfos {
		var withdrawal *sdk.AccAddress
		contract := common.HexToAddress(fee.ContractAddress)
		deployer, _ := sdk.AccAddressFromBech32(fee.DeployerAddress)
		_withdrawal, err := sdk.AccAddressFromBech32(fee.DeployerAddress)
		if err == nil {
			withdrawal = &_withdrawal
		}

		// Set initial contracts receiving transaction fees
		k.SetFee(ctx, contract, deployer, withdrawal)
		k.SetFeeInverse(ctx, deployer, contract)
	}
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		DevFeeInfos: k.GetAllFees(ctx),
	}
}
