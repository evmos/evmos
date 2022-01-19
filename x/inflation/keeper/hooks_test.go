package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmoapp "github.com/osmosis-labs/osmosis/app"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/mint/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEndOfEpochMintedCoinDistribution(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	setupGaugeForLPIncentives(t, app, ctx)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	// set developer rewards address
	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{
			Address: sdk.AccAddress([]byte("addr1---------------")).String(),
			Weight:  sdk.NewDec(1),
		},
	}
	app.MintKeeper.SetParams(ctx, mintParams)

	// setup developer rewards account
	app.MintKeeper.CreateDeveloperVestingModuleAccount(
		ctx, sdk.NewCoin("stake", sdk.NewInt(156*500000*2)))

	height := int64(1)
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	// correct rewards
	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams = app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, height)

		// test that the dev rewards module account balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}

	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

	lastHalvenPeriod = app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs)

	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)

		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams = app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, height)

		// test that the balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}
}

func TestMintedCoinDistributionWhenDevRewardsAddressEmpty(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	setupGaugeForLPIncentives(t, app, ctx)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	// setup developer rewards account
	app.MintKeeper.CreateDeveloperVestingModuleAccount(
		ctx, sdk.NewCoin("stake", sdk.NewInt(156*500000*2)))

	height := int64(1)
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	// correct rewards
	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams := app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking.Add(mintParams.DistributionProportions.DeveloperRewards)).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, height)

		// test that the dev rewards module account balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}

	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

	lastHalvenPeriod = app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs)

	for ; height < lastHalvenPeriod+app.MintKeeper.GetParams(ctx).ReductionPeriodInEpochs; height++ {
		devRewardsModuleAcc := app.AccountKeeper.GetModuleAccount(ctx, types.DeveloperVestingModuleAcctName)
		devRewardsModuleOrigin := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)

		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		mintParams := app.MintKeeper.GetParams(ctx)
		mintedCoin := app.MintKeeper.GetMinter(ctx).EpochProvision(mintParams)
		expectedRewardsAmount := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.Staking.Add(mintParams.DistributionProportions.DeveloperRewards)).Amount
		expectedRewards := sdk.NewDecCoin("stake", expectedRewardsAmount)

		// check community pool balance increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool.Add(expectedRewards), feePoolNew.CommunityPool, expectedRewards.String())

		// test that the dev rewards module account balance decreased by the correct amount
		devRewardsModuleAfter := app.BankKeeper.GetAllBalances(ctx, devRewardsModuleAcc.GetAddress())
		expectedDevRewards := app.MintKeeper.GetProportions(ctx, mintedCoin, mintParams.DistributionProportions.DeveloperRewards)
		require.Equal(t, devRewardsModuleAfter.Add(expectedDevRewards), devRewardsModuleOrigin, expectedRewards.String())
	}
}

func TestEndOfEpochNoDistributionWhenIsNotYetStartTime(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.MintingRewardsDistributionStartEpoch = 4
	app.MintKeeper.SetParams(ctx, mintParams)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	setupGaugeForLPIncentives(t, app, ctx)

	params := app.IncentivesKeeper.GetParams(ctx)
	futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))

	height := int64(1)
	// Run through epochs 0 through mintParams.MintingRewardsDistributionStartEpoch - 1
	// ensure no rewards sent out
	for ; height < mintParams.MintingRewardsDistributionStartEpoch; height++ {
		feePoolOrigin := app.DistrKeeper.GetFeePool(ctx)
		app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
		app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)

		// check community pool balance not increase
		feePoolNew := app.DistrKeeper.GetFeePool(ctx)
		require.Equal(t, feePoolOrigin.CommunityPool, feePoolNew.CommunityPool, "height = %v", height)
	}
	// Run through epochs mintParams.MintingRewardsDistributionStartEpoch
	// ensure tokens distributed
	app.EpochsKeeper.BeforeEpochStart(futureCtx, params.DistrEpochIdentifier, height)
	app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, height)
	require.NotEqual(t, sdk.DecCoins{}, app.DistrKeeper.GetFeePool(ctx).CommunityPool,
		"Tokens to community pool at start distribution epoch")

	// halven period should be set to mintParams.MintingRewardsDistributionStartEpoch
	lastHalvenPeriod := app.MintKeeper.GetLastHalvenEpochNum(ctx)
	require.Equal(t, lastHalvenPeriod, mintParams.MintingRewardsDistributionStartEpoch)
}

func setupGaugeForLPIncentives(t *testing.T, app *osmoapp.OsmosisApp, ctx sdk.Context) {
	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := simapp.FundAccount(app.BankKeeper, ctx, addr, coins)
	require.NoError(t, err)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	_, err = app.IncentivesKeeper.CreateGauge(ctx, true, addr, coins, distrTo, time.Now(), 1)
	require.NoError(t, err)
}
