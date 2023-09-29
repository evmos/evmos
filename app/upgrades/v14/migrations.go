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

	for _, oldMultisig := range oldMultisigs {
		oldMultisigAcc := sdk.MustAccAddressFromBech32(oldMultisig)
		delegations := sk.GetAllDelegatorDelegations(ctx, oldMultisigAcc)

		for _, delegation := range delegations {
			unbondAmount, err := InstantUnbonding(ctx, bk, sk, delegation, bondDenom)
			if err != nil {
				return err
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
			return err
		}
	}

	// Delegate from multisig to same validators
	for _, migration := range migratedDelegations {
		val, ok := sk.GetValidator(ctx, migration.validator)
		if !ok {
			return fmt.Errorf("validator %s not found", migration.validator.String())
		}
		if _, err := sk.Delegate(ctx, newMultisig, migration.amount, stakingtypes.Unbonded, val, true); err != nil {
			return err
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
	unbondCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondAmount))

	// transfer the validator tokens to the not bonded pool if necessary
	validator, found := sk.GetValidator(ctx, valAddr)
	if !found {
		return math.Int{}, fmt.Errorf("validator %s not found", valAddr)
	}
	if validator.IsBonded() {
		if err := bk.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, unbondCoins); err != nil {
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
