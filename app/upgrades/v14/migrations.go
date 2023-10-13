// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigratedDelegation holds the relevant information about a delegation to be migrated
type MigratedDelegation struct {
	// validator is the validator address
	validator sdk.ValAddress
	// amount is the amount to be delegated
	amount math.Int
}

// MigrateNativeMultisigs migrates the native multisigs to the new team multisig including all
// staking delegations.
func MigrateNativeMultisigs(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper, newMultisig sdk.AccAddress, oldMultisigs ...string) error {
	var (
		// bondDenom is the staking bond denomination used
		bondDenom = sk.BondDenom(ctx)
		// migratedDelegations stores all delegations that must be migrated
		migratedDelegations []MigratedDelegation
	)

	// NOTE: We are checking the bond denomination here because this is what caused the panic
	// during the v14.0.0 upgrade.
	if bondDenom == "" {
		return fmt.Errorf("invalid bond denom received during migration: %s", bondDenom)
	}

	logger := ctx.Logger().With("module", "v14-migrations")

	for _, oldMultisig := range oldMultisigs {
		oldMultisigAcc := sdk.MustAccAddressFromBech32(oldMultisig)
		delegations := sk.GetAllDelegatorDelegations(ctx, oldMultisigAcc)

		for _, delegation := range delegations {
			unbondAmount, err := InstantUnbonding(ctx, bk, sk, delegation, bondDenom)
			if err != nil {
				// NOTE: log error instead of aborting the whole migration
				logger.Error(fmt.Sprintf("failed to unbond delegation %s from validator %s: %s", delegation.GetDelegatorAddr(), delegation.GetValidatorAddr(), err.Error()))
				continue
			}

			// NOTE: if the unbonded amount is zero we are not adding it
			// to the migrated delegations, because there is nothing to be delegated.
			if unbondAmount.IsZero() {
				continue
			}

			migratedDelegations = append(migratedDelegations, MigratedDelegation{
				validator: delegation.GetValidatorAddr(),
				amount:    unbondAmount,
			})
		}

		// Send coins to new team multisig
		balances := bk.GetAllBalances(ctx, oldMultisigAcc)
		err := bk.SendCoins(ctx, oldMultisigAcc, newMultisig, balances)
		if err != nil {
			// NOTE: log error instead of aborting the whole migration
			logger.Error(fmt.Sprintf("failed to send coins from %s to %s: %s", oldMultisig, newMultisig.String(), err.Error()))
			continue
		}
	}

	// Delegate from multisig to same validators
	for _, migration := range migratedDelegations {
		val, ok := sk.GetValidator(ctx, migration.validator)
		if !ok {
			// NOTE: log error instead of aborting the whole migration
			logger.Error(fmt.Sprintf("validator %s not found", migration.validator.String()))
			continue
		}
		if _, err := sk.Delegate(ctx, newMultisig, migration.amount, stakingtypes.Unbonded, val, true); err != nil {
			// NOTE: log error instead of aborting the whole migration
			logger.Error(fmt.Sprintf("failed to delegate %s from %s to %s",
				migration.amount.String(), newMultisig.String(), migration.validator.String(),
			))
			continue
		}
	}

	return nil
}

// InstantUnbonding will execute an instant unbonding of the given delegation
//
// NOTE: this logic contains code copied from different functions of the
// staking keepers's undelegate implementation.
func InstantUnbonding(
	ctx sdk.Context,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	del stakingtypes.Delegation,
	bondDenom string,
) (unbondAmount math.Int, err error) {
	delAddr := del.GetDelegatorAddr()
	valAddr := del.GetValidatorAddr()

	// Check if there are any outstanding redelegations for delegator - validator pair
	// - this would require additional handling
	if sk.HasReceivingRedelegation(ctx, delAddr, valAddr) {
		return math.Int{}, fmt.Errorf("redelegation(s) found for delegator %s and validator %s", delAddr, valAddr)
	}

	unbondAmount, err = sk.Unbond(ctx, delAddr, valAddr, del.GetShares())
	if err != nil {
		return math.Int{}, err
	}

	// NOTE: if the unbonded amount is zero there are no tokens to be transferred between the staking pools
	// and neither to be undelegated from the module to the account
	if unbondAmount.IsZero() {
		return unbondAmount, nil
	}

	// NOTE: We avoid using sdk.NewCoins here because it panics on an invalid denom,
	// which was the problem in the v14.0.0 release.
	unbondCoins := sdk.Coins{sdk.Coin{Denom: bondDenom, Amount: unbondAmount}}
	if err := unbondCoins.Validate(); err != nil {
		return math.Int{}, fmt.Errorf("invalid unbonding coins: %v", err)
	}

	// transfer the validator tokens to the not bonded pool if necessary
	validator, found := sk.GetValidator(ctx, valAddr)
	if !found {
		return math.Int{}, fmt.Errorf("validator %s not found", valAddr)
	}
	if validator.IsBonded() {
		if err := bk.SendCoinsFromModuleToModule(
			ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, unbondCoins,
		); err != nil {
			panic(err)
		}
	}

	// Transfer the tokens from the not bonded pool to the delegator
	if err := bk.UndelegateCoinsFromModuleToAccount(
		ctx, stakingtypes.NotBondedPoolName, delAddr, unbondCoins,
	); err != nil {
		return math.Int{}, err
	}

	return unbondAmount, nil
}
