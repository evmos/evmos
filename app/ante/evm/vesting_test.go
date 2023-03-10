package evm_test

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	ethante "github.com/evmos/evmos/v12/app/ante/evm"
	"github.com/evmos/evmos/v12/testutil"
	testutiltx "github.com/evmos/evmos/v12/testutil/tx"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v12/x/vesting/types"
)

// global variables used for testing the eth vesting ante handler
var (
	balances       = sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	quarter        = sdk.NewCoins(sdk.NewInt64Coin("test", 250))
	lockupPeriods  = sdkvesting.Periods{{Length: 5000, Amount: balances}}
	vestingPeriods = sdkvesting.Periods{
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
	}
	vestingCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000)))
)

// TestEthVestingTransactionDecorator tests the EthVestingTransactionDecorator ante handler.
func (suite *AnteTestSuite) TestEthVestingTransactionDecorator() {
	addr := testutiltx.GenerateAddress()

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		To:       &addr,
		Amount:   big.NewInt(1000000000),
		GasLimit: 100000,
		GasPrice: big.NewInt(1000000000),
	}
	tx := evmtypes.NewTx(ethTxParams)
	tx.From = addr.Hex()

	testcases := []struct {
		name        string
		tx          sdk.Tx
		malleate    func()
		expPass     bool
		errContains string
	}{
		{
			"pass - valid transaction, no vesting account",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			true,
			"",
		},
		{
			"fail - invalid transaction",
			&testutiltx.InvalidTx{},
			func() {},
			false,
			"invalid message type",
		},
		{
			"fail - from address not found",
			tx,
			func() {},
			false,
			"does not exist: unknown address",
		},
		{
			"pass - valid transaction, vesting account",
			tx,
			func() {
				baseAcc := authtypes.NewBaseAccountWithAddress(addr.Bytes())
				vestingAcc := vestingtypes.NewClawbackVestingAccount(
					baseAcc, addr.Bytes(), vestingCoins, time.Now(), lockupPeriods, vestingPeriods,
				)
				acc := suite.app.AccountKeeper.NewAccount(suite.ctx, vestingAcc)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				denom := suite.app.EvmKeeper.GetParams(suite.ctx).EvmDenom
				coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000000000)))
				err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, addr.Bytes(), coins)
				suite.Require().NoError(err, "failed to fund account")
			},
			true,
			"",
		},
		{
			"fail - valid transaction, vesting account, no balance",
			tx,
			func() {
				baseAcc := authtypes.NewBaseAccountWithAddress(addr.Bytes())
				vestingAcc := vestingtypes.NewClawbackVestingAccount(
					baseAcc, addr.Bytes(), vestingCoins, time.Now(), lockupPeriods, vestingPeriods,
				)
				acc := suite.app.AccountKeeper.NewAccount(suite.ctx, vestingAcc)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			"account has no balance to execute transaction",
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.malleate()

			dec := ethante.NewEthVestingTransactionDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper)
			_, err := dec.AnteHandle(suite.ctx, tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
			} else {
				suite.Require().ErrorContains(err, tc.errContains, tc.name)
			}
		})
	}
}
