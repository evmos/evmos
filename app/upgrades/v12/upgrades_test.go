package v12_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/cometbft/cometbft/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v15/testutil"
	feemarkettypes "github.com/evmos/evmos/v15/x/feemarket/types"

	"github.com/evmos/evmos/v15/app"
	v12 "github.com/evmos/evmos/v15/app/upgrades/v12"
	"github.com/evmos/evmos/v15/utils"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Evmos
	consAddress sdk.ConsAddress
}

func (suite *UpgradeTestSuite) SetupTest(chainID string) {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState(), chainID)
	suite.ctx = suite.app.BaseApp.NewContext(
		checkTx,
		testutil.NewHeader(
			1,
			time.Date(2022, 5, 9, 8, 0, 0, 0, time.UTC),
			chainID,
			suite.consAddress.Bytes(),
			tmhash.Sum([]byte("block_id")),
			tmhash.Sum([]byte("validators")),
		),
	)

	cp := suite.app.BaseApp.GetConsensusParams(suite.ctx)
	suite.ctx = suite.ctx.WithConsensusParams(cp)
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (suite *UpgradeTestSuite) TestReturnFundsFromCommunityPool() {
	suite.SetupTest(utils.TestnetChainID + "-2")

	// send funds to the community pool
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	sender := sdk.AccAddress(priv.PubKey().Address().Bytes())

	res, ok := math.NewIntFromString(v12.MaxRecover)
	suite.Require().True(ok)

	coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, res))
	err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, sender, coins)
	suite.Require().NoError(err)
	err = suite.app.DistrKeeper.FundCommunityPool(suite.ctx, coins, sender)
	suite.Require().NoError(err)

	balanceBefore := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(balanceBefore.AmountOf(utils.BaseDenom), math.LegacyNewDecFromInt(res))

	// return funds to accounts affected
	err = v12.ReturnFundsFromCommunityPool(suite.ctx, suite.app.DistrKeeper)
	suite.Require().NoError(err)

	// store the addresses on a map to check if there're
	// duplicated addresses
	uniqueAddrs := make(map[string]bool)
	// check balance of affected accounts
	for i := range v12.Accounts {
		addr := sdk.MustAccAddressFromBech32(v12.Accounts[i][0])
		// check for duplicated addresses
		found := uniqueAddrs[v12.Accounts[i][0]]
		suite.Require().False(found, "found account %s duplicated", v12.Accounts[i][0])
		uniqueAddrs[v12.Accounts[i][0]] = true

		res, ok := math.NewIntFromString(v12.Accounts[i][1])
		suite.Require().True(ok)
		suite.Require().True(res.IsPositive())
		balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
		suite.Require().Equal(balance.Amount, res)
	}

	balanceAfter := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().True(balanceAfter.IsZero(), "Community pool balance should be zero after the distribution, but is %d", balanceAfter.AmountOf(utils.BaseDenom))
}
