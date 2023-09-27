// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
	"github.com/ethereum/go-ethereum/common"
	vestingprecompile "github.com/evmos/evmos/v14/precompiles/vesting"
	evmkeeper "github.com/evmos/evmos/v14/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
	vestingkeeper "github.com/evmos/evmos/v14/x/vesting/keeper"
)

const (
	// newTeamPremintWallet is the new vesting team multisig
	newTeamPremintWallet = "0x83ef4C096F9A9daC61081121CCE30578fe437182"
	// newTeamStrategicReserve is the new strategic reserve multisig
	newTeamStrategicReserve = "0x29fDcB7b64B84fD54D0fB0E04A8f6B062046fc6F"
	// OldFunder1 is one of the old vesting funders to be replaced
	OldFunder1 = "evmos1jcltmuhplrdcwp7stlr4hlhlhgd4htqh3a79sq"
	// OldFunder2 is the other old vesting funder to be replaced
	OldFunder2 = "evmos1cml96vmptgw99syqrrz8az79xer2pcgp84pdun"
	// oldTeamPremintWallet is the old team premint wallet
	oldTeamPremintWallet = "evmos1jcltmuhplrdcwp7stlr4hlhlhgd4htqh3a79sq"
	// VestingAddrByFunder1 is the vesting account funded by OldFunder1
	VestingAddrByFunder1 = "evmos1pxjncpsu2rd3hjxgswkqaenrpu3v5yxurzm7jp"
)

var (
	// VestingAddrsByFunder2 is a slice of vesting accounts funded by OldFunder1
	VestingAddrsByFunder2 = []string{
		"evmos12aqyq9d4k7a8hzh5av2xgxp0njan48498dvj2s",
		"evmos1rtj2r4eaz0v68mxjt5jleynm85yjfu2uxm7pxx",
	}

	// OldStrategicReserves is a list of old multisigs to be replaced
	OldStrategicReserves = []string{
		"evmos1gzsvk8rruqn2sx64acfsskrwy8hvrmafqkaze8", // Strategic Reserve 1
		"evmos1fx944mzagwdhx0wz7k9tfztc8g3lkfk6rrgv6l", // Strategic Reserve 2
	}

	newTeamPremintWalletAddr    = common.HexToAddress(newTeamPremintWallet)
	NewTeamPremintWalletAcc     = sdk.AccAddress(newTeamPremintWalletAddr.Bytes())
	newTeamStrategicReserveAddr = common.HexToAddress(newTeamStrategicReserve)
	NewTeamStrategicReserveAcc  = sdk.AccAddress(newTeamStrategicReserveAddr.Bytes())
)

// CreateUpgradeHandler creates an SDK upgrade handler for v13
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	ek *evmkeeper.Keeper,
	sk stakingkeeper.Keeper,
	vk vestingkeeper.Keeper,
	ck consensuskeeper.Keeper,
	clientKeeper ibctmmigrations.ClientKeeper,
	pk paramskeeper.Keeper,
	cdc codec.BinaryCodec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// !! ATTENTION !!
		// v14 upgrade handler
		// !! WHEN UPGRADING TO SDK v0.47 MAKE SURE TO INCLUDE THIS
		// source: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xconsensus
		// !! If not migrating to v0.47 in this upgrade,
		// !! make sure to move it to the corresponding upgrade
		// Migrate Tendermint consensus parameters from x/params module to a
		// dedicated x/consensus module.

		// Set param key table for params module migration
		for _, subspace := range pk.GetSubspaces() {
			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck,nolintlint
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable()
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck,nolintlint
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
			case ibctransfertypes.ModuleName:
				keyTable = ibctransfertypes.ParamKeyTable()
			case evmtypes.ModuleName:
				keyTable = evmtypes.ParamKeyTable() //nolint:staticcheck
			case feemarkettypes.ModuleName:
				keyTable = feemarkettypes.ParamKeyTable()
			default:
				continue
			}
			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		baseAppLegacySS := pk.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

		baseapp.MigrateParams(ctx, baseAppLegacySS, &ck)

		// Include this when migrating to ibc-go v7 (optional)
		// source: https://github.com/cosmos/ibc-go/blob/v7.2.0/docs/migrations/v6-to-v7.md
		// prune expired tendermint consensus states to save storage space
		if _, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, cdc, clientKeeper); err != nil {
			return nil, err
		}
		// !! ATTENTION !!

		// Leave modules are as-is to avoid running InitGenesis.
		// NOTE: we are running the module migrations BEFORE the migration of mainnet data, to ensure
		// that params are correctly found when e.g. delegating!
		logger.Info("running module migrations ...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		logger.Info("adding vesting EVM extension to active precompiles")
		if err := EnableVestingExtension(ctx, ek); err != nil {
			// log error instead of aborting the upgrade
			logger.Error("error while enabling vesting extension", "error", err)
		}

		logger.Info("updating vesting funders to new team multisig")
		if err := UpdateVestingFunders(ctx, vk, NewTeamPremintWalletAcc); err != nil {
			logger.Error("error while updating vesting funders", "error", err)
		}

		logger.Info("migrating strategic reserves")
		if err := MigrateNativeMultisigs(ctx, bk, sk, NewTeamStrategicReserveAcc, OldStrategicReserves...); err != nil {
			logger.Error("error while migrating native multisigs", "error", err)
		}

		logger.Info("migrating team premint wallet")
		if err := MigrateNativeMultisigs(ctx, bk, sk, NewTeamPremintWalletAcc, oldTeamPremintWallet); err != nil {
			logger.Error("error while migrating team premint wallet", "error", err)
		}

		return vm, nil
	}
}

// EnableVestingExtension appends the address of the vesting EVM extension
// to the list of active precompiles.
func EnableVestingExtension(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	// Get the list of active precompiles from the genesis state
	params := evmKeeper.GetParams(ctx)
	activePrecompiles := params.ActivePrecompiles
	activePrecompiles = append(activePrecompiles, vestingprecompile.Precompile{}.Address().String())
	params.ActivePrecompiles = activePrecompiles

	return evmKeeper.SetParams(ctx, params)
}
