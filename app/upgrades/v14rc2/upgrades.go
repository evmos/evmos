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
	// oldFunder1 is one of the old vesting funders to be replaced
	oldFunder1 = "evmos1sgjgup7wz3qyfcqqpr66jlm9qpk3j63ajupc9l"
	// oldFunder2 is the other old vesting funder to be replaced
	oldFunder2 = "evmos1xp38jqcjf2s7wyuyh3fwrjukuj4ny54k2yaq97"
	// newTeamMultisig is the new vesting team multisig
	newTeamMultisig = "0x83ef4C096F9A9daC61081121CCE30578fe437182"
)

var (
	// AffectedAddresses is a map of vesting accounts to be updated
	// with their respective funder addresses
	AffectedAddresses = map[string]string{
		"evmos12aqyq9d4k7a8hzh5av2xgxp0njan48498dvj2s": oldFunder2,
		"evmos1pxjncpsu2rd3hjxgswkqaenrpu3v5yxurzm7jp": oldFunder1,
		"evmos1rtj2r4eaz0v68mxjt5jleynm85yjfu2uxm7pxx": oldFunder2,
	}

	// OldMultisigs is a list of old vesting multisigs to be replaced
	OldMultisigs = []string{
		"evmos1z8ynrnhdn4l69mu6v6ckjr4wukcacd0e7j0akn", // Strategic Reserve 1
		"evmos1w2rl60wr9sxjv60qsh9v8aratk0x2r3v78utzt", // Strategic Reserve 2
		"evmos1fgg4xaakwmrxdk9my6uc8nxeatf7u35uaal529", // Strategic Reserve 3
		"evmos15xm3h3fgjrkqtkr79t7rj9spq3qlzuheae5vss", // Strategic Reserve 4
		"evmos15l8jnxynhldtydknzla2xpv8uxg00xgmg2enst", // Strategic Reserve 5
		"evmos1sgjgup7wz3qyfcqqpr66jlm9qpk3j63ajupc9l", // Team Premint Wallet
		"evmos1f7vxxvmd544dkkmyxan76t76d39k7j3gr8d45y", // Consolidation Wallet
	}

	newTeamMultisigAddr = common.HexToAddress(newTeamMultisig)
	NewTeamMultisigAcc  = sdk.AccAddress(newTeamMultisigAddr.Bytes())
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
			if err := MigrateNativeMultisigs(ctx, bk, sk, OldMultisigs); err != nil {
				logger.Error("error while migrating native multisigs", "error", err)
			}
		}

		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// UpdateVestingFunders updates the vesting funders for accounts managed by the team
// to the new dedicated multisig address.
func UpdateVestingFunders(ctx sdk.Context, k vestingkeeper.Keeper) error {
	for address, oldFunder := range AffectedAddresses {
		vestingAcc := sdk.MustAccAddressFromBech32(address)
		oldFunderAcc := sdk.MustAccAddressFromBech32(oldFunder)
		msgUpdate := vestingtypes.NewMsgUpdateVestingFunder(oldFunderAcc, NewTeamMultisigAcc, vestingAcc)

		if _, err := k.UpdateVestingFunder(ctx, msgUpdate); err != nil {
			return err
		}
	}

	return nil
}

// MigrateNativeMultisigs migrates the native multisigs to the new team multisig including all
// staking delegations.
func MigrateNativeMultisigs(ctx sdk.Context, bk bankkeeper.Keeper, sk stakingkeeper.Keeper, oldMultisigs []string) error {
	var (
		// bondDenom is the staking bond denomination used
		bondDenom = sk.BondDenom(ctx)
		// delegationsMap holds the validator addresses and the total amount to be delegated to
		// each of them.
		delegationsMap = make(map[string]math.Int)
	)

	for _, oldMultisig := range oldMultisigs {
		oldMultisigAcc := sdk.MustAccAddressFromBech32(oldMultisig)
		delegations := sk.GetAllDelegatorDelegations(ctx, oldMultisigAcc)
		fmt.Printf("\ncurrent balance for %s: %v\n", oldMultisig, bk.GetAllBalances(ctx, oldMultisigAcc))

		for _, delegation := range delegations {
			unbondAmount, err := InstantUnbonding(ctx, bk, sk, delegation, bondDenom)
			if err != nil {
				return err
			}

			if _, ok := delegationsMap[delegation.ValidatorAddress]; !ok {
				delegationsMap[delegation.ValidatorAddress] = math.ZeroInt()
			}
			delegationsMap[delegation.ValidatorAddress] = delegationsMap[delegation.ValidatorAddress].Add(unbondAmount)
		}

		// Send coins to new team multisig
		balances := bk.GetAllBalances(ctx, oldMultisigAcc)
		fmt.Printf(" --> old balance: %v\n", bk.GetAllBalances(ctx, oldMultisigAcc))
		err := bk.SendCoins(ctx, oldMultisigAcc, NewTeamMultisigAcc, balances)
		if err != nil {
			return err
		}

		fmt.Printf("Sent %s from %q to %q\n", balances, oldMultisigAcc, NewTeamMultisigAcc)
		fmt.Printf(" --> new multisig balance: %v\n", bk.GetAllBalances(ctx, NewTeamMultisigAcc))
	}

	// Delegate from multisig to same validators
	for validator, amount := range delegationsMap {
		validatorAddr, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			return err
		}
		val, ok := sk.GetValidator(ctx, validatorAddr)
		if !ok {
			return fmt.Errorf("validator %s not found", validator)
		}
		if _, err := sk.Delegate(ctx, NewTeamMultisigAcc, amount, stakingtypes.Unbonded, val, true); err != nil {
			return err
		}

		fmt.Printf("Delegated %s from %q to %q\n", amount, NewTeamMultisigAcc, validatorAddr)
		fmt.Printf(" --> new multisig balance: %v\n", bk.GetAllBalances(ctx, NewTeamMultisigAcc))
	}

	return nil
}

// InstantUnbonding will execute an instant unbonding of the given delegation
//
// NOTE: this logic is copied from the staking keepers's undelegate implementation
func InstantUnbonding(
	ctx sdk.Context,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
	del stakingtypes.Delegation,
	bondDenom string,
) (unbondAmount math.Int, err error) {
	delAddr := del.GetDelegatorAddr()
	valAddr := del.GetValidatorAddr()

	unbondAmount, err = sk.Unbond(ctx, delAddr, valAddr, del.GetShares())
	fmt.Printf("unbonded %s from %s\n", unbondAmount, delAddr)
	if err != nil {
		return unbondAmount, err
	}

	// transfer the validator tokens to the not bonded pool if necessary
	validator, found := sk.GetValidator(ctx, valAddr)
	if !found {
		return unbondAmount, fmt.Errorf("validator %s not found", valAddr)
	}
	if validator.IsBonded() {
		bondedTokensToNotBonded(ctx, bk, unbondAmount, bondDenom)
	}

	// Transfer the tokens from the not bonded pool to the delegator
	if err := bk.UndelegateCoinsFromModuleToAccount(
		ctx, stakingtypes.NotBondedPoolName, delAddr, sdk.Coins{sdk.Coin{Denom: bondDenom, Amount: unbondAmount}},
	); err != nil {
		return unbondAmount, err
	}
	fmt.Printf("  --> updated balance for %s: %v\n", delAddr, bk.GetAllBalances(ctx, delAddr))

	return unbondAmount, nil
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func bondedTokensToNotBonded(ctx sdk.Context, bk bankkeeper.Keeper, amount math.Int, bondDenom string) {
	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, amount))
	if err := bk.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins); err != nil {
		panic(err)
	}
}
