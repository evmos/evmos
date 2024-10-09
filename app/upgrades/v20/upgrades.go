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
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v20/utils"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	"github.com/evmos/evmos/v20/x/evm/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for Evmos v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	gk govkeeper.Keeper,
	sk *stakingkeeper.Keeper,
	bk bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		logger.Debug("Enabling gov precompile...")
		if err := EnableGovPrecompile(ctx, ek); err != nil {
			logger.Error("error while enabling gov precompile", "error", err.Error())
		}

		// run the sdk v0.50 migrations
		logger.Debug("Running module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		logger.Debug("Updating expedited prop params...")
		if err := UpdateExpeditedPropsParams(ctx, gk); err != nil {
			logger.Error("error while updating gov params", "error", err.Error())
		}

		if err := AddSuperPowerValidator(ctx, logger, sk, bk); err != nil {
			return nil, err
		}

		return vm, nil
	}
}

func EnableGovPrecompile(ctx sdk.Context, ek *evmkeeper.Keeper) error {
	// Enable gov precompile
	params := ek.GetParams(ctx)
	params.ActiveStaticPrecompiles = append(params.ActiveStaticPrecompiles, types.GovPrecompileAddress)
	if err := params.Validate(); err != nil {
		return err
	}
	return ek.SetParams(ctx, params)
}

func UpdateExpeditedPropsParams(ctx sdk.Context, gk govkeeper.Keeper) error {
	params, err := gk.Params.Get(ctx)
	if err != nil {
		return err
	}

	// use the same denom as the min deposit denom
	// also amount must be greater than MinDeposit amount
	denom := params.MinDeposit[0].Denom
	expDepAmt := params.ExpeditedMinDeposit[0].Amount
	if expDepAmt.LTE(params.MinDeposit[0].Amount) {
		expDepAmt = params.MinDeposit[0].Amount.MulRaw(govv1.DefaultMinExpeditedDepositTokensRatio)
	}
	params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(denom, expDepAmt))

	// if expedited voting period > voting period
	// set expedited voting period to be half the voting period
	if params.ExpeditedVotingPeriod != nil && params.VotingPeriod != nil && *params.ExpeditedVotingPeriod > *params.VotingPeriod {
		expPeriod := *params.VotingPeriod / 2
		params.ExpeditedVotingPeriod = &expPeriod
	}

	if err := params.ValidateBasic(); err != nil {
		return err
	}
	return gk.Params.Set(ctx, params)
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
	_, err = srv.CreateValidator(ctx, &stakingtypes.MsgCreateValidator{
		Description:       stakingtypes.NewDescription(moniker, "new super powerful val", "none", "none", "none"),
		Commission:        stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2), math.LegacyNewDecWithPrec(5, 2)),
		MinSelfDelegation: currentSupply,
		DelegatorAddress:  valOperAccAddr.String(),
		ValidatorAddress:  valAddr,
		Pubkey:            pubkey,
		Value:             sdk.NewCoin(utils.BaseDenom, currentSupply.MulRaw(3)),
	})

	return err
}
