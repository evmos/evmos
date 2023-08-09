// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14rc2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v13/utils"
	vestingkeeper "github.com/evmos/evmos/v13/x/vesting/keeper"
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
