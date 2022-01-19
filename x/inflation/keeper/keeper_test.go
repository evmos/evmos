package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"

	"github.com/osmosis-labs/osmosis/app"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/osmosis-labs/osmosis/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.OsmosisApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, Time: time.Now().UTC()})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestMintCoinsToFeeCollectorAndGetProportions() {
	mintKeeper := suite.app.MintKeeper

	// When coin is minted to the fee collector
	fee := sdk.NewCoin("stake", sdk.NewInt(0))
	fees := sdk.NewCoins(fee)
	coin := mintKeeper.GetProportions(suite.ctx, fee, sdk.NewDecWithPrec(2, 1))
	suite.Equal("0stake", coin.String())

	// When mint the 100K stake coin to the fee collector
	fee = sdk.NewCoin("stake", sdk.NewInt(100000))
	fees = sdk.NewCoins(fee)

	err := simapp.FundModuleAccount(suite.app.BankKeeper,
		suite.ctx,
		authtypes.FeeCollectorName,
		fees)
	suite.NoError(err)

	// check proportion for 20%
	coin = mintKeeper.GetProportions(suite.ctx, fee, sdk.NewDecWithPrec(2, 1))
	suite.Equal(fees[0].Amount.Quo(sdk.NewInt(5)), coin.Amount)
}

func (suite *KeeperTestSuite) TestDistrAssetToDeveloperRewardsAddrWhenNotEmpty() {
	mintKeeper := suite.app.MintKeeper
	params := suite.app.MintKeeper.GetParams(suite.ctx)
	devRewardsReceiver := sdk.AccAddress([]byte("addr1---------------"))
	gaugeCreator := sdk.AccAddress([]byte("addr2---------------"))
	devRewardsReceiver2 := sdk.AccAddress([]byte("addr3---------------"))
	devRewardsReceiver3 := sdk.AccAddress([]byte("addr4---------------"))
	params.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{
			Address: devRewardsReceiver.String(),
			Weight:  sdk.NewDec(1),
		},
	}
	suite.app.MintKeeper.SetParams(suite.ctx, params)

	// Create record
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, gaugeCreator, coins)
	suite.NoError(err)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second,
	}
	gaugeId, err := suite.app.IncentivesKeeper.CreateGauge(suite.ctx, true, gaugeCreator, coins, distrTo, time.Now(), 1)
	suite.NoError(err)
	err = suite.app.PoolIncentivesKeeper.UpdateDistrRecords(suite.ctx, poolincentivestypes.DistrRecord{
		GaugeId: gaugeId,
		Weight:  sdk.NewInt(100),
	})
	suite.NoError(err)

	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin)
	suite.NoError(err)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	suite.Equal(
		mintCoin.Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt(),
		suite.app.BankKeeper.GetAllBalances(suite.ctx, feeCollector).AmountOf("stake"))
	suite.Equal(
		mintCoin.Amount.ToDec().Mul(params.DistributionProportions.CommunityPool),
		feePool.CommunityPool.AmountOf("stake"))
	suite.Equal(
		mintCoin.Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).TruncateInt(),
		suite.app.BankKeeper.GetBalance(suite.ctx, devRewardsReceiver, "stake").Amount)

	// Test for multiple dev reward addresses
	params.WeightedDeveloperRewardsReceivers = []types.WeightedAddress{
		{
			Address: devRewardsReceiver2.String(),
			Weight:  sdk.NewDecWithPrec(6, 1),
		},
		{
			Address: devRewardsReceiver3.String(),
			Weight:  sdk.NewDecWithPrec(4, 1),
		},
	}
	suite.app.MintKeeper.SetParams(suite.ctx, params)

	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin)
	suite.NoError(err)

	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).Mul(params.WeightedDeveloperRewardsReceivers[0].Weight).TruncateInt(),
		suite.app.BankKeeper.GetBalance(suite.ctx, devRewardsReceiver2, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.DeveloperRewards).Mul(params.WeightedDeveloperRewardsReceivers[1].Weight).TruncateInt(),
		suite.app.BankKeeper.GetBalance(suite.ctx, devRewardsReceiver3, "stake").Amount)
}

func (suite *KeeperTestSuite) TestDistrAssetToCommunityPoolWhenNoDeveloperRewardsAddr() {
	mintKeeper := suite.app.MintKeeper

	params := suite.app.MintKeeper.GetParams(suite.ctx)
	// At this time, there is no distr record, so the asset should be allocated to the community pool.
	mintCoin := sdk.NewCoin("stake", sdk.NewInt(100000))
	mintCoins := sdk.Coins{mintCoin}
	err := mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin)
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
	feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	// PoolIncentives + DeveloperRewards + CommunityPool => CommunityPool
	proportionToCommunity := params.DistributionProportions.PoolIncentives.
		Add(params.DistributionProportions.DeveloperRewards).
		Add(params.DistributionProportions.CommunityPool)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt(),
		suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(proportionToCommunity),
		feePool.CommunityPool.AmountOf("stake"))

	// Mint more and community pool should be increased
	err = mintKeeper.MintCoins(suite.ctx, mintCoins)
	suite.NoError(err)
	err = mintKeeper.DistributeMintedCoin(suite.ctx, mintCoin)
	suite.NoError(err)

	distribution.BeginBlocker(suite.ctx, abci.RequestBeginBlock{}, *suite.app.DistrKeeper)

	feePool = suite.app.DistrKeeper.GetFeePool(suite.ctx)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(params.DistributionProportions.Staking).TruncateInt().Mul(sdk.NewInt(2)),
		suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, "stake").Amount)
	suite.Equal(
		mintCoins[0].Amount.ToDec().Mul(proportionToCommunity).Mul(sdk.NewDec(2)),
		feePool.CommunityPool.AmountOf("stake"))
}
