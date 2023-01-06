// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package v11

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/evmos/evmos/v10/types"
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

		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("distributing incentivized testnet rewards...")
			HandleRewardDistribution(ctx, bk, sk, dk, logger)
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

// HandleRewardDistribution handles the logic for the reward distribution, it only commits to the db if successful
func HandleRewardDistribution(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper, dk distributionkeeper.Keeper, logger log.Logger) {
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

// DistributeRewards distributes the token allocations from the Olympus Mons incentivized testnet
func DistributeRewards(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper, dk distributionkeeper.Keeper) error {
	funder := sdk.MustAccAddressFromBech32(FundingAccount)
	numValidators := sdk.NewInt(int64(len(Validators)))

	for _, allocation := range Allocations {

		// send rewards to receivers
		receiver := sdk.MustAccAddressFromBech32(allocation[0])
		receivingAmount, ok := sdk.NewIntFromString(allocation[1])
		if !ok {
			return errorsmod.Wrapf(errortypes.ErrInvalidType,
				"cannot retrieve allocation from string for address %s",
				allocation[0])
		}
		reward := sdk.Coins{
			sdk.NewCoin(types.BaseDenom, receivingAmount),
		}
		err := bk.SendCoins(ctx, funder, receiver, reward)
		if err != nil {
			return err
		}

		// delegate receiver's rewards to validators selected validators equally
		delegationAmt := reward.QuoInt(numValidators)[0].Amount
		remainderAmount := (reward[0].Amount).Mod(numValidators)
		for i, validatorBech32 := range Validators {
			validatorAddress, err := sdk.ValAddressFromBech32(validatorBech32)
			if err != nil {
				return err
			}
			validator, found := sk.GetValidator(ctx, validatorAddress)
			if !found {
				return errorsmod.Wrapf(slashingtypes.ErrBadValidatorAddr,
					"validator address %s cannot be found",
					validatorAddress)
			}
			// we delegate the remainder to the first validator, for the sake of testing consistency
			// this remainder is in the order of 10^-18 evmos, and at most 10^-15 evmos after all rewards are allocated
			if remainderAmount.IsPositive() && i == 0 {
				_, err = sk.Delegate(ctx, receiver, delegationAmt.Add(remainderAmount), 1, validator, true)
			} else {
				// 1 signifies unbonded tokens, subtractAccount being true means delegation, not redelegation
				_, err = sk.Delegate(ctx, receiver, delegationAmt, 1, validator, true)
			}
			if err != nil {
				return err
			}
		}
	}

	// transfer all remaining tokens (1.775M = 7.4M - 5.625M) after rewards distribution to the community pool
	remainingFunds := bk.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(FundingAccount))
	err := dk.FundCommunityPool(ctx, remainingFunds, funder)
	if err != nil {
		return err
	}

	return nil
}
