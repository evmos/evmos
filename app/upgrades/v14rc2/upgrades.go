// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14rc2

import (
	"fmt"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v13/utils"
	vestingkeeper "github.com/evmos/evmos/v13/x/vesting/keeper"
	vestingtypes "github.com/evmos/evmos/v13/x/vesting/types"
)

const (
	// newTeamPremintWallet is the new vesting team multisig
	newTeamPremintWallet = "0x83ef4C096F9A9daC61081121CCE30578fe437182"
	// newTeamStrategicReserve is the new strategic reserve multisig
	newTeamStrategicReserve = "0x29fDcB7b64B84fD54D0fB0E04A8f6B062046fc6F"
	// OldFunder1 is one of the old vesting funders to be replaced
	OldFunder1 = "evmos1sgjgup7wz3qyfcqqpr66jlm9qpk3j63ajupc9l"
	// OldFunder2 is the other old vesting funder to be replaced
	OldFunder2 = "evmos1xp38jqcjf2s7wyuyh3fwrjukuj4ny54k2yaq97"
	// oldTeamPremintWallet is the old team premint wallet
	oldTeamPremintWallet = "evmos1sgjgup7wz3qyfcqqpr66jlm9qpk3j63ajupc9l"
	// VestingAddrByFunder1 is the vesting account funded by OldFunder1
	VestingAddrByFunder1 = "evmos1pxjncpsu2rd3hjxgswkqaenrpu3v5yxurzm7jp"
)

var (
	// VestingAddrsByFunder2 is a slice of vesting accounts funded by OldFunder1
	VestingAddrsByFunder2 = []string{
		"evmos12aqyq9d4k7a8hzh5av2xgxp0njan48498dvj2s",
		"evmos1rtj2r4eaz0v68mxjt5jleynm85yjfu2uxm7pxx",
	}

	// OldMultisigs is a list of old vesting multisigs to be replaced
	OldMultisigs = []string{
		"evmos1z8ynrnhdn4l69mu6v6ckjr4wukcacd0e7j0akn", // Strategic Reserve 1
		"evmos1w2rl60wr9sxjv60qsh9v8aratk0x2r3v78utzt", // Strategic Reserve 2
		"evmos1fgg4xaakwmrxdk9my6uc8nxeatf7u35uaal529", // Strategic Reserve 3
		"evmos15xm3h3fgjrkqtkr79t7rj9spq3qlzuheae5vss", // Strategic Reserve 4
		"evmos15l8jnxynhldtydknzla2xpv8uxg00xgmg2enst", // Strategic Reserve 5
	}

	newTeamPremintWalletAddr    = common.HexToAddress(newTeamPremintWallet)
	NewTeamPremintWalletAcc     = sdk.AccAddress(newTeamPremintWalletAddr.Bytes())
	newTeamStrategicReserveAddr = common.HexToAddress(newTeamStrategicReserve)
	NewTeamStrategicReserveAcc  = sdk.AccAddress(newTeamStrategicReserveAddr.Bytes())
)

// CreateUpgradeHandler creates an SDK upgrade handler for v14
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	vk vestingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if utils.IsMainnet(ctx.ChainID()) {
			logger.Debug("updating vesting funders to new team multisig")
			if err := UpdateVestingFunders(ctx, vk); err != nil {
				// log error instead of aborting the upgrade
				logger.Error("error while updating vesting funders", "error", err)
			}

			logger.Debug("migrating strategic reserves")
			if err := MigrateNativeMultisigs(ctx, bk, sk, OldMultisigs, NewTeamStrategicReserveAcc); err != nil {
				logger.Error("error while migrating native multisigs", "error", err)
			}

			logger.Debug("migration team premint wallet")
			if err := MigrateNativeMultisigs(ctx, bk, sk, []string{oldTeamPremintWallet}, NewTeamPremintWalletAcc); err != nil {
				logger.Error("error while migrating team premint wallet", "error", err)
			}
		}

		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// UpdateVestingFunders updates the vesting funders for accounts managed by the team
// to the new dedicated multisig address.
func UpdateVestingFunders(ctx sdk.Context, vk vestingkeeper.Keeper) error {
	if _, err := UpdateVestingFunder(ctx, vk, VestingAddrByFunder1, OldFunder1); err != nil {
		return err
	}
	for _, address := range VestingAddrsByFunder2 {
		if _, err := UpdateVestingFunder(ctx, vk, address, OldFunder2); err != nil {
			return err
		}
	}
	return nil
}

// UpdateVestingFunder updates the vesting funder for a single vesting account when address and the previous funder
// are given as strings.
func UpdateVestingFunder(ctx sdk.Context, k vestingkeeper.Keeper, address, oldFunder string) (*vestingtypes.MsgUpdateVestingFunderResponse, error) {
	vestingAcc := sdk.MustAccAddressFromBech32(address)
	oldFunderAcc := sdk.MustAccAddressFromBech32(oldFunder)
	msgUpdate := vestingtypes.NewMsgUpdateVestingFunder(oldFunderAcc, NewTeamPremintWalletAcc, vestingAcc)

	return k.UpdateVestingFunder(ctx, msgUpdate)
}

// MigratedDelegation holds the relevant information about a delegation to be migrated
type MigratedDelegation struct {
	// validator is the validator address
	validator sdk.ValAddress
	// amount is the amount to be delegated
	amount math.Int
}

// MigrateNativeMultisigs migrates the native multisigs to the new team multisig including all
// staking delegations.
func MigrateNativeMultisigs(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper, oldMultisigs []string, targetAcc sdk.AccAddress) error {
	var (
		// bondDenom is the staking bond denomination used
		bondDenom = sk.BondDenom(ctx)
		// migratedDelegations stores all delegations that must be migrated
		migratedDelegations []MigratedDelegation
	)

	for _, oldMultisig := range oldMultisigs {
		oldMultisigAcc := sdk.MustAccAddressFromBech32(oldMultisig)
		delegations := sk.GetAllDelegatorDelegations(ctx, oldMultisigAcc)
		fmt.Printf("Iterating over %d delegations for %s\n", len(delegations), oldMultisigAcc.String())

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
		err := bk.SendCoins(ctx, oldMultisigAcc, targetAcc, balances)
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
		if _, err := sk.Delegate(ctx, targetAcc, migration.amount, stakingtypes.Unbonded, val, true); err != nil {
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
		return unbondAmount, fmt.Errorf("redelegation(s) found for delegator %s and validator %s", delAddr, valAddr)
	}

	unbondAmount, err = sk.Unbond(ctx, delAddr, valAddr, del.GetShares())
	if err != nil {
		return unbondAmount, err
	}
	unbondCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondAmount))

	// transfer the validator tokens to the not bonded pool if necessary
	validator, found := sk.GetValidator(ctx, valAddr)
	if !found {
		return unbondAmount, fmt.Errorf("validator %s not found", valAddr)
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
		return unbondAmount, err
	}

	return unbondAmount, nil
}
