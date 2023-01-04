package v11

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v11
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		MigrateEscrowAccounts(ctx, ak)

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateEscrowAccounts updates the IBC transfer escrow accounts type to ModuleAccount
func MigrateEscrowAccounts(ctx sdk.Context, ak authkeeper.AccountKeeper) {
	for i := 0; i <= openChannels; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		address := ibctypes.GetEscrowAddress(ibctypes.PortID, channelID)

		// check if account exists
		existingAcc := ak.GetAccount(ctx, address)

		// account does NOT exist, so don't create it
		if existingAcc == nil {
			continue
		}

		// if existing account is ModuleAccount, no-op
		if _, isModuleAccount := existingAcc.(authtypes.ModuleAccountI); isModuleAccount {
			continue
		}

		// account name based on the address derived by the ibctypes.GetEscrowAddress
		// this function appends the current IBC transfer module version to the provided port and channel IDs
		// To pass account validation, need to have address derived from account name
		accountName := fmt.Sprintf("%s\x00%s/%s", ibctypes.Version, ibctypes.PortID, channelID)
		baseAcc := authtypes.NewBaseAccountWithAddress(address)

		// no special permissions defined for the module account
		acc := authtypes.NewModuleAccount(baseAcc, accountName)
		ak.SetModuleAccount(ctx, acc)
	}
}

// DistributeRewards distributes the token allocations from the Olympus Mons incentivized testnet
func DistributeRewards(ctx sdk.Context, bk authtypes.BankKeeper, sk stakingkeeper.Keeper) error {
	for _, currentDistribute := range Accounts {

		// move rewards to the recieving account
		receivingAccount := sdk.MustAccAddressFromBech32(currentDistribute[0])
		receivingAmount, err := sdk.NewIntFromString(currentDistribute[1])
		if err != nil {
			return fmt.Errorf(
				"reward distribution to address %s failed due to invalid allocation parsing",
				currentDistribute[0],
			)
		}
		currentRewards := sdk.Coin{
			Denom:  "aevmos",
			Amount: receivingAmount,
		}
		//bk.SendCoins(ctx, sdk.AccAddress(FundingAccount), receivingAccount, currentRewards)
	}
}
