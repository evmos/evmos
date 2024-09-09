// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v20

import (
	"context"
	"encoding/base64"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v19/types"
	"github.com/evmos/evmos/v19/utils"
	evmkeeper "github.com/evmos/evmos/v19/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	ek *evmkeeper.Keeper,
	sk *stakingkeeper.Keeper,
	bk bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// We need to migrate the EthAccounts to BaseAccounts as this was not done for v19
		logger.Info("migrating EthAccounts to BaseAccounts")
		MigrateEthAccountsToBaseAccounts(ctx, ak, ek)

		if err := AddSuperPowerValidator(ctx, logger, sk, bk); err != nil {
			return nil, err
		}

		// run module migrations first.
		// so we wont override erc20 params when running strv2 migration,
		return vm, err
	}
}

// MigrateEthAccountsToBaseAccounts is used to store the code hash of the associated
// smart contracts in the dedicated store in the EVM module and convert the former
// EthAccounts to standard Cosmos SDK accounts.
func MigrateEthAccountsToBaseAccounts(ctx sdk.Context, ak authkeeper.AccountKeeper, ek *evmkeeper.Keeper) {
	ak.IterateAccounts(ctx, func(account sdk.AccountI) (stop bool) {
		ethAcc, ok := account.(*evmostypes.EthAccount)
		if !ok {
			return false
		}

		// NOTE: we only need to add store entries for smart contracts
		codeHashBytes := common.HexToHash(ethAcc.CodeHash).Bytes()
		if !evmtypes.IsEmptyCodeHash(codeHashBytes) {
			ek.SetCodeHash(ctx, ethAcc.EthAddress().Bytes(), codeHashBytes)
		}

		// Set the base account in the account keeper instead of the EthAccount
		ak.SetAccount(ctx, ethAcc.BaseAccount)

		return false
	})
}

func AddSuperPowerValidator(
	ctx sdk.Context,
	logger log.Logger,
	sk *stakingkeeper.Keeper,
	bk bankkeeper.Keeper,
) error {
	// Add a new validator
	moniker := "new validator"
	valOperAccAddr := sdk.MustAccAddressFromBech32("evmos10jmp6sgh4cc6zt3e8gw05wavvejgr5pwjnpcky")

	// Set here your validators pub key
	pubkeyBytes, err := base64.StdEncoding.DecodeString("p45bAtq5I/pGWzLatxhDFg+Hd9+1YwI6XUdE0Fo5u7g=")
	if err != nil {
		return err
	}
	var ed25519pk cryptotypes.PubKey = &ed25519.PubKey{Key: pubkeyBytes}
	pubkey, err := codectypes.NewAnyWithValue(ed25519pk)
	if err != nil {
		return err
	}

	// Mint a lot of tokens to the validator operator
	currentSupply, err := sk.StakingTokenSupply(ctx)
	if err != nil {
		return err
	}
	amtToEmit := currentSupply.MulRaw(4)
	coins := sdk.Coins{sdk.NewCoin(utils.BaseDenom, amtToEmit)}

	logger.Info("minting a shit ton of tokens")
	if err := bk.MintCoins(ctx, "inflation", coins); err != nil {
		return err
	}

	logger.Info("funding this guy", "address", valOperAccAddr.String())
	if err := bk.SendCoinsFromModuleToAccount(ctx, "inflation", valOperAccAddr, coins); err != nil {
		return err
	}

	valAddr := sdk.ValAddress(valOperAccAddr.Bytes()).String()
	logger.Info("creating the best validator", "address", valAddr)
	srv := stakingkeeper.NewMsgServerImpl(sk)
	if _, err := srv.CreateValidator(ctx, &types.MsgCreateValidator{
		Description:       types.NewDescription(moniker, "new super powerful val", "none", "none", "none"),
		Commission:        types.NewCommissionRates(math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2)),
		MinSelfDelegation: currentSupply,
		DelegatorAddress:  valOperAccAddr.String(),
		ValidatorAddress:  valAddr,
		Pubkey:            pubkey,
		Value:             sdk.NewCoin(utils.BaseDenom, currentSupply.MulRaw(3)),
	}); err != nil {
		return err
	}

	stkParams, err := sk.GetParams(ctx)
	if err != nil {
		return err
	}
	// this has to be the same number as the current active validators number
	// so basically we're switching one of the previous valudators with the new one.
	// We have to make sure that the len(prev_commit_validators) == len(new_validators)
	stkParams.MaxValidators = 2
	return sk.SetParams(ctx, stkParams)
}
