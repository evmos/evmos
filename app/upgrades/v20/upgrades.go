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
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v19/utils"
)

// CreateUpgradeHandler creates an SDK upgrade handler for Evmos v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	sk *stakingkeeper.Keeper,
	bk bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// run module migrations first.
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		if err := AddSuperPowerValidator(ctx, logger, sk, bk); err != nil {
			return nil, err
		}

		return vm, err
	}
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
	_, err = srv.CreateValidator(ctx, &types.MsgCreateValidator{
		Description:       types.NewDescription(moniker, "new super powerful val", "none", "none", "none"),
		Commission:        types.NewCommissionRates(math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2)),
		MinSelfDelegation: currentSupply,
		DelegatorAddress:  valOperAccAddr.String(),
		ValidatorAddress:  valAddr,
		Pubkey:            pubkey,
		Value:             sdk.NewCoin(utils.BaseDenom, currentSupply.MulRaw(3)),
	})

	return err
}
