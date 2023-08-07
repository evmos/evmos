// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14rc2

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
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
			if err := MigrateNativeMultisigs(ctx, sk, OldMultisigs); err != nil {
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
func MigrateNativeMultisigs(ctx sdk.Context, sk stakingkeeper.Keeper, oldMultisigs []string) error {
	for _, oldMultisig := range oldMultisigs {
		oldMultisigAcc := sdk.MustAccAddressFromBech32(oldMultisig)
		delegations := sk.GetAllDelegatorDelegations(ctx, oldMultisigAcc)

		for _, delegation := range delegations {
			fmt.Printf("delegator: %s, validator: %s, amount: %s\n", delegation.DelegatorAddress, delegation.ValidatorAddress, delegation.GetShares())
		}
	}

	return nil
}
