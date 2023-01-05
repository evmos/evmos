package v11

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	"github.com/evmos/evmos/v10/types"
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

		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("distributing incentivized testnet rewards...")
			HandleRewardDistribution(ctx, bk, sk, logger)
		}

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

// HandleMainnetUpgrade handles the logic for the reward distribution, it only commits to the db if successful
func HandleRewardDistribution(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper, logger log.Logger) {
	// use a cache context as a rollback mechanism in case
	// the distrbution fails
	cacheCtx, writeFn := ctx.CacheContext()
	err := DistributeRewards(cacheCtx, bk, sk)
	if err != nil {
		// log error instead of aborting the upgrade
		logger.Error("failed to distribute rewards", "error", err.Error())
	} else {
		writeFn()
	}
}

// DistributeRewards distributes the token allocations from the Olympus Mons incentivized testnet
func DistributeRewards(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper) error {
	// TODO check the remaining rewards on each iteration to avoid sending more/less than supposed to (similar to v9.1 upgrade)
	for _, currentDistribute := range Accounts {

		// move rewards to the receiving account
		receivingAccount := sdk.MustAccAddressFromBech32(currentDistribute[0])
		receivingAmount, ok := sdk.NewIntFromString(currentDistribute[1])
		if !ok {
			return fmt.Errorf(
				"reward distribution to address %s failed due to invalid parsing",
				currentDistribute[0],
			)
		}
		currentRewards := sdk.Coins{
			sdk.NewCoin(types.BaseDenom, receivingAmount),
		}
		err := bk.SendCoins(ctx, sdk.MustAccAddressFromBech32(FundingAccount), receivingAccount, currentRewards)
		if err != nil {
			return fmt.Errorf(
				"unable to send coins from fund account to participant account",
			)
		}

		// stake from the receiving account to all validators equally
		numValidators := len(Validators)
		currentStakeAmount := (currentRewards.QuoInt(sdk.NewInt(int64(numValidators)))).AmountOf(types.BaseDenom)
		for _, validatorBech32 := range Validators {
			validatorAddress, err := sdk.ValAddressFromBech32(validatorBech32)
			if err != nil {
				return fmt.Errorf(
					"unable to convert validator address %s",
					validatorBech32,
				)
			}
			validator, found := sk.GetValidator(ctx, validatorAddress)
			if !found {
				return fmt.Errorf(
					"unable to find validator corresponding to address %s",
					validatorBech32,
				)
			}
			// 1 signifies unbonded tokens, subtractAccount being true means delegation, not redelegation
			_, err = sk.Delegate(ctx, receivingAccount, currentStakeAmount, 1, validator, true)
			if err != nil {
				return fmt.Errorf(
					"unable to delegate to validator with address %s",
					validatorBech32,
				)
			}
		}
	}

	// transfer all remaining tokens after distribution to the community pool
	remainingFunds := bk.GetAllBalances(ctx, sdk.AccAddress(FundingAccount))
	err := bk.SendCoins(ctx, sdk.AccAddress(FundingAccount), sdk.AccAddress(CommunityPoolAccount), remainingFunds)
	if err != nil {
		return fmt.Errorf(
			"unable to send coins from fund account to community pool",
		)
	}

	return nil
}
