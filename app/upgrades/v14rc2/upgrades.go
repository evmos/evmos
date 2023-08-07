// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14rc2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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

	newTeamMultisigAddr = common.HexToAddress(newTeamMultisig)
	NewTeamMultisigAcc  = sdk.AccAddress(newTeamMultisigAddr.Bytes())
)

// CreateUpgradeHandler creates an SDK upgrade handler for v14
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	vk vestingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if utils.IsMainnet(ctx.ChainID()) {
			logger.Debug("updating vesting funders to new team multisig")
			if err := UpdateVestingFunders(ctx, vk); err != nil {
				// log error instead of aborting the upgrade
				logger.Error("error while updating vesting funders", "error", err)
				// TODO: remove panic after debugging
				panic(err)
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
