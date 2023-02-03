package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/evmos/v11/tests"
	"github.com/evmos/evmos/v11/testutil"
	"github.com/evmos/evmos/v11/x/claims/types"
	vestingtypes "github.com/evmos/evmos/v11/x/vesting/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {
	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"claim disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
		},
		{
			"not claim time",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
		},
		{
			"claim enabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Time{}
				params.DurationUntilDecay = time.Hour
				params.DurationOfDecay = time.Hour
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimsKeeper.EndBlocker(suite.ctx)
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackEmptyAccounts() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr3 := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name       string
		expBalance int64
		malleate   func()
	}{
		{
			"no claims records",
			0,
			func() {
			},
		},
		{
			"no account",
			0,
			func() {
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"sequence not zero",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 1))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"no balance",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance non zero, vesting account is ignored",
			0,
			func() {
				bAcc := authtypes.NewBaseAccount(addr, nil, 0, 0)
				funder := sdk.AccAddress(tests.GenerateAddress().Bytes())
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust)))

				vestingAcc := vestingtypes.NewClawbackVestingAccount(bAcc, funder, coins, time.Now().UTC(), nil, nil)
				suite.app.AccountKeeper.SetAccount(suite.ctx, vestingAcc)

				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance non zero, base account is ignored",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance non zero, module account is ignored",
			0,
			func() {
				ba := authtypes.NewBaseAccount(addr, nil, 0, 0)
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewModuleAccount(ba, "testmodule"))

				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance non zero, eth account",
			types.GenesisDust,
			func() {
				baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)
				ethAccount := ethermint.EthAccount{
					BaseAccount: baseAccount,
					CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
				}
				suite.app.AccountKeeper.SetAccount(suite.ctx, &ethAccount)

				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance zero on claims denoms and non zero in other denoms, is ignored",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin("testcoin", sdk.NewInt(types.GenesisDust)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance more than dust, is ignored",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust+100000)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"multiple denoms, only claims denom is clawed back",
			types.GenesisDust,
			func() {
				ethAccount := newEthAccount(authtypes.NewBaseAccount(addr, nil, 0, 0))
				suite.app.AccountKeeper.SetAccount(suite.ctx, &ethAccount)

				coin1 := sdk.NewCoin("testcoin", sdk.NewInt(types.GenesisDust))
				coin2 := sdk.NewCoin("testcoin1", sdk.NewInt(types.GenesisDust))
				coin3 := sdk.NewCoin("testcoin2", sdk.NewInt(types.GenesisDust))
				coin4 := sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust))

				coins := sdk.NewCoins(coin1, coin2, coin3, coin4)

				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"multiple accounts, all clawed back",
			types.GenesisDust * 3,
			func() {
				ethAccount1 := newEthAccount(authtypes.NewBaseAccount(addr, nil, 0, 0))
				ethAccount2 := newEthAccount(authtypes.NewBaseAccount(addr, nil, 0, 0))
				ethAccount3 := newEthAccount(authtypes.NewBaseAccount(addr, nil, 0, 0))
				suite.app.AccountKeeper.SetAccount(suite.ctx, &ethAccount1)
				suite.app.AccountKeeper.SetAccount(suite.ctx, &ethAccount2)
				suite.app.AccountKeeper.SetAccount(suite.ctx, &ethAccount3)

				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(types.GenesisDust)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr, coins)
				suite.Require().NoError(err)
				err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr2, coins)
				suite.Require().NoError(err)
				err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr3, coins)
				suite.Require().NoError(err)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr2, types.ClaimsRecord{})
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr3, types.ClaimsRecord{})
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimsKeeper.ClawbackEmptyAccounts(suite.ctx, types.DefaultClaimsDenom)

			moduleAcc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, distrtypes.ModuleName)
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, moduleAcc.GetAddress(), types.DefaultClaimsDenom)
			suite.Require().Equal(tc.expBalance, balance.Amount.Int64())

			// test that all claims records are deleted
			claimsRecords := suite.app.ClaimsKeeper.GetClaimsRecords(suite.ctx)
			suite.Require().Len(claimsRecords, 0)
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackEscrowedTokensABCI() {
	var amount int64 = 10000

	testCases := []struct {
		name     string
		funds    int64
		malleate func()
	}{
		{
			"no balance",
			0,
			func() {
			},
		},
		{
			"balance on module account",
			amount,
			func() {
				coins := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(amount)))
				err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.ClaimsKeeper.ClawbackEscrowedTokens(suite.ctx)
			suite.Require().NoError(err)

			acc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, distrtypes.ModuleName)
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, acc.GetAddress(), types.DefaultClaimsDenom)
			suite.Require().Equal(balance.Amount, sdk.NewInt(tc.funds))
		})
	}
}

func newEthAccount(baseAccount *authtypes.BaseAccount) ethermint.EthAccount {
	return ethermint.EthAccount{
		BaseAccount: baseAccount,
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
}
