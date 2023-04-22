// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v11

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ica "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts"
	genesistypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/evmos/evmos/v12/utils"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v11
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	dk distributionkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if utils.IsMainnet(ctx.ChainID()) {
			logger.Debug("distributing incentivized testnet rewards...")
			HandleRewardDistribution(ctx, logger, bk, sk, dk)
		}

		MigrateEscrowAccounts(ctx, logger, ak)

		// create ICS27 Controller submodule params, with the controller module NOT enabled
		gs := &genesistypes.GenesisState{
			ControllerGenesisState: genesistypes.ControllerGenesisState{},
			HostGenesisState: genesistypes.HostGenesisState{
				Port: icatypes.HostPortID,
				Params: icahosttypes.Params{
					HostEnabled: true,
					AllowMessages: []string{
						sdk.MsgTypeURL(&banktypes.MsgSend{}),
						sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
						sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
						sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
						sdk.MsgTypeURL(&govtypes.MsgVote{}),
						sdk.MsgTypeURL(&govtypes.MsgVoteWeighted{}),
						sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
						sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
						sdk.MsgTypeURL(&stakingtypes.MsgCancelUnbondingDelegation{}),
						sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
						sdk.MsgTypeURL(&transfertypes.MsgTransfer{}),
					},
				},
			},
		}

		bz, err := icatypes.ModuleCdc.MarshalJSON(gs)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to marshal %s genesis state", icatypes.ModuleName)
		}

		// Register the consensus version in the version map to avoid the SDK from triggering the default
		// InitGenesis function.
		vm[icatypes.ModuleName] = ica.AppModule{}.ConsensusVersion()

		_ = mm.Modules[icatypes.ModuleName].InitGenesis(ctx, icatypes.ModuleCdc, bz)

		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateEscrowAccounts updates the IBC transfer escrow accounts type to ModuleAccount
func MigrateEscrowAccounts(ctx sdk.Context, logger log.Logger, ak authkeeper.AccountKeeper) {
	for i := 0; i <= OpenChannels; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		address := transfertypes.GetEscrowAddress(transfertypes.PortID, channelID)

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

		// account name based on the address derived by the transfertypes.GetEscrowAddress
		// this function appends the current IBC transfer module version to the provided port and channel IDs
		// To pass account validation, need to have address derived from account name
		accountName := fmt.Sprintf("%s\x00%s/%s", transfertypes.Version, transfertypes.PortID, channelID)
		baseAcc := authtypes.NewBaseAccountWithAddress(address)

		// Set same account number and sequence as the existing account
		if err := baseAcc.SetAccountNumber(existingAcc.GetAccountNumber()); err != nil {
			// log error instead of aborting the upgrade
			logger.Error("failed to set escrow account number for account", accountName, "error", err.Error())
		}
		if err := baseAcc.SetSequence(existingAcc.GetSequence()); err != nil {
			// log error instead of aborting the upgrade
			logger.Error("failed to set escrow account sequence for account", accountName, "error", err.Error())
		}

		// no special permissions defined for the module account
		acc := authtypes.NewModuleAccount(baseAcc, accountName)
		ak.SetModuleAccount(ctx, acc)
	}
}

// HandleRewardDistribution handles the logic for the reward distribution,
// it only commits to the db if successful
func HandleRewardDistribution(
	ctx sdk.Context,
	logger log.Logger,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	dk distributionkeeper.Keeper,
) {
	// use a cache context as a rollback mechanism in case
	// the distrbution fails
	cacheCtx, writeFn := ctx.CacheContext()
	err := DistributeRewards(cacheCtx, bk, sk, dk)
	if err != nil {
		// log error instead of aborting the upgrade
		logger.Error("failed to distribute rewards", "error", err.Error())
	} else {
		writeFn()
	}
}

// DistributeRewards distributes the token allocations from the Olympus Mons
// incentivized testnet for completing the Mars Meteor Missions
func DistributeRewards(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper,
	dk distributionkeeper.Keeper,
) error {
	funder := sdk.MustAccAddressFromBech32(FundingAccount)

	for _, allocation := range Allocations {
		// send reward to receiver
		receiver := sdk.MustAccAddressFromBech32(allocation[0])

		amount, ok := sdk.NewIntFromString(allocation[1])
		if !ok {
			return errorsmod.Wrapf(
				errortypes.ErrInvalidType,
				"cannot retrieve allocation amount from string for address %s",
				allocation[0],
			)
		}

		if !amount.IsPositive() {
			return errorsmod.Wrapf(
				errortypes.ErrInvalidCoins,
				"amount cannot be zero negative for address %s, got %s",
				allocation[0], allocation[1],
			)
		}

		// delegate receiver's rewards to selected validator
		validatorAddress, err := sdk.ValAddressFromBech32(allocation[2])
		if err != nil {
			return err
		}

		reward := sdk.Coins{{Denom: utils.BaseDenom, Amount: amount}}

		if err := bk.SendCoins(ctx, funder, receiver, reward); err != nil {
			return err
		}

		validator, found := sk.GetValidator(ctx, validatorAddress)
		if !found {
			return errorsmod.Wrap(stakingtypes.ErrNoValidatorFound, allocation[2])
		}

		_, err = sk.Delegate(ctx, receiver, amount, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return err
		}
	}

	// transfer all remaining tokens (1.775M = 7.4M - 5.625M) after rewards distribution
	// to the community pool
	remainingFunds := bk.GetBalance(ctx, funder, utils.BaseDenom)
	if !remainingFunds.Amount.IsPositive() {
		return nil
	}

	return dk.FundCommunityPool(ctx, sdk.Coins{remainingFunds}, funder)
}
