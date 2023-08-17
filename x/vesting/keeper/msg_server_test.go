package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/evmos/evmos/v14/testutil"
	utiltx "github.com/evmos/evmos/v14/testutil/tx"
	evmostypes "github.com/evmos/evmos/v14/types"
	"github.com/evmos/evmos/v14/x/vesting/types"
)

var (
	vestAmount     = int64(1000)
	balances       = sdk.NewCoins(sdk.NewInt64Coin("test", vestAmount))
	quarter        = sdk.NewCoins(sdk.NewInt64Coin("test", 250))
	addr3          = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	addr4          = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	funder         = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	vestingAddr    = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	lockupPeriods  = sdkvesting.Periods{{Length: 5000, Amount: balances}}
	vestingPeriods = sdkvesting.Periods{
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
		{Length: 2000, Amount: quarter},
	}
)

func (suite *KeeperTestSuite) TestMsgFundVestingAccount() {
	testCases := []struct {
		name               string
		funder             sdk.AccAddress
		vestingAddr        sdk.AccAddress
		lockup             sdkvesting.Periods
		vesting            sdkvesting.Periods
		expectExtraBalance int64
		// initClawback determines if the clawback vesting account should be initialized for the test case
		initClawback bool
		// preFundClawback determines if the clawback vesting account should be already be funded before the test case
		// this is used to test merging new vesting amounts to existing lockup and vesting schedules
		preFundClawback bool
		expPass         bool
		errContains     string
	}{
		{
			name:         "pass - lockup and vesting defined",
			funder:       funder,
			vestingAddr:  vestingAddr,
			lockup:       lockupPeriods,
			vesting:      vestingPeriods,
			initClawback: true,
			expPass:      true,
		},
		{
			name:         "pass - only vesting",
			funder:       funder,
			vestingAddr:  vestingAddr,
			vesting:      vestingPeriods,
			initClawback: true,
			expPass:      true,
		},
		{
			name:         "pass - only lockup",
			funder:       funder,
			vestingAddr:  vestingAddr,
			lockup:       lockupPeriods,
			initClawback: true,
			expPass:      true,
		},
		{
			name:         "fail - account is no clawback account",
			funder:       funder,
			vestingAddr:  vestingAddr,
			lockup:       lockupPeriods,
			vesting:      vestingPeriods,
			initClawback: false,
			expPass:      false,
		},
		{
			name:               "true - fund existing vesting account",
			funder:             funder,
			vestingAddr:        vestingAddr,
			lockup:             lockupPeriods,
			vesting:            vestingPeriods,
			expectExtraBalance: vestAmount,
			initClawback:       true,
			preFundClawback:    true,
			expPass:            true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // Reset
			ctx := sdk.WrapSDKContext(suite.ctx)

			// fund the recipient account to set the account and then
			// send funds over to the funder account so balance is empty
			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, tc.vestingAddr, balances)
			suite.Require().NoError(err, "failed to fund target account")
			err = s.app.BankKeeper.SendCoins(s.ctx, tc.vestingAddr, tc.funder, balances)
			suite.Require().NoError(err, "failed to send tokens to funder account")

			// create a clawback vesting account if necessary
			if tc.initClawback {
				msgCreate := types.NewMsgCreateClawbackVestingAccount(tc.funder, tc.vestingAddr, false)
				resCreate, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msgCreate)
				suite.Require().NoError(err)
				suite.Require().Equal(&types.MsgCreateClawbackVestingAccountResponse{}, resCreate)
			}

			// fund the vesting account prior to actual test if desired
			if tc.preFundClawback {
				// in order to fund the vesting account additionally to the actual main test case, we need to
				// send it some more funds
				err = testutil.FundAccount(s.ctx, s.app.BankKeeper, tc.funder, balances)
				suite.Require().NoError(err, "failed to fund funder account")
				// fund vesting account
				msgFund := types.NewMsgFundVestingAccount(tc.funder, tc.vestingAddr, time.Now(), lockupPeriods, vestingPeriods)
				_, err = suite.app.VestingKeeper.FundVestingAccount(ctx, msgFund)
				suite.Require().NoError(err, "failed to fund vesting account")
			}

			// fund the vesting account
			msg := types.NewMsgFundVestingAccount(
				tc.funder,
				tc.vestingAddr,
				time.Now(),
				tc.lockup,
				tc.vesting,
			)

			res, err := suite.app.VestingKeeper.FundVestingAccount(ctx, msg)

			expRes := &types.MsgFundVestingAccountResponse{}
			balanceFunder := suite.app.BankKeeper.GetBalance(suite.ctx, tc.funder, "test")
			balanceVestingAddr := suite.app.BankKeeper.GetBalance(suite.ctx, tc.vestingAddr, "test")

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)

				accI := suite.app.AccountKeeper.GetAccount(suite.ctx, tc.vestingAddr)
				suite.Require().NotNil(accI)
				suite.Require().IsType(&types.ClawbackVestingAccount{}, accI)
				suite.Require().Equal(sdk.NewInt64Coin("test", 0), balanceFunder)
				suite.Require().Equal(sdk.NewInt64Coin("test", vestAmount+tc.expectExtraBalance), balanceVestingAddr)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().ErrorContains(err, tc.errContains)
			}
		})
	}
}

// NOTE: This function tests cases which require a different setup than the standard
// cases in TestMsgFundVestingAccount.
func (suite *KeeperTestSuite) TestMsgFundVestingAccountSpecialCases() {
	// ---------------------------
	// Test blocked address
	suite.Run("fail - blocked address", func() {
		suite.SetupTest()
		msg := &types.MsgFundVestingAccount{
			FunderAddress:  funder.String(),
			VestingAddress: authtypes.NewModuleAddress("transfer").String(),
			StartTime:      time.Now(),
			LockupPeriods:  lockupPeriods,
			VestingPeriods: vestingPeriods,
		}

		_, err = suite.app.VestingKeeper.FundVestingAccount(suite.ctx, msg)
		suite.Require().Error(err, "expected blocked address error")
		suite.Require().ErrorContains(err, "is not allowed to receive funds")
	})

	// ---------------------------
	// Test wrong funder by first creating a clawback vesting account
	// and then trying to fund it with a different funder
	suite.Run("fail - wrong funder", func() {
		suite.SetupTest()

		// fund the recipient account to set the account
		err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, vestingAddr, balances)
		suite.Require().NoError(err, "failed to fund target account")
		msgCreate := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, false)
		_, err = suite.app.VestingKeeper.CreateClawbackVestingAccount(suite.ctx, msgCreate)
		suite.Require().NoError(err, "failed to create clawback vesting account")

		msg := &types.MsgFundVestingAccount{
			FunderAddress:  addr3.String(),
			VestingAddress: vestingAddr.String(),
			StartTime:      time.Now(),
			LockupPeriods:  lockupPeriods,
			VestingPeriods: vestingPeriods,
		}
		_, err = suite.app.VestingKeeper.FundVestingAccount(suite.ctx, msg)
		suite.Require().Error(err, "expected wrong funder error")
		suite.Require().ErrorContains(err, fmt.Sprintf("%s can only accept grants from account %s", vestingAddr, funder))
	})
}

func (suite *KeeperTestSuite) TestMsgCreateClawbackVestingAccount() {
	funderAddr, _ := utiltx.NewAccAddressAndKey()
	vestingAddr, _ := utiltx.NewAccAddressAndKey()

	testcases := []struct {
		name        string
		malleate    func(funder sdk.AccAddress, vestingAddr sdk.AccAddress)
		funder      sdk.AccAddress
		vestingAddr sdk.AccAddress
		expPass     bool
		errContains string
	}{
		{
			name:        "fail - account does not exist",
			malleate:    func(funder sdk.AccAddress, vestingAddr sdk.AccAddress) {},
			funder:      funderAddr,
			vestingAddr: vestingAddr,
			expPass:     false,
			errContains: fmt.Sprintf("account %s does not exist", vestingAddr),
		},
		{
			name: "fail - account is not an eth account",
			malleate: func(funder sdk.AccAddress, vestingAddr sdk.AccAddress) {
				acc := authtypes.NewBaseAccountWithAddress(vestingAddr)
				s.app.AccountKeeper.SetAccount(s.ctx, acc)
			},
			funder:      funderAddr,
			vestingAddr: vestingAddr,
			expPass:     false,
			errContains: fmt.Sprintf("account %s is not an Ethereum account", vestingAddr),
		},
		{
			name: "fail - vesting account already exists",
			malleate: func(funder sdk.AccAddress, vestingAddr sdk.AccAddress) {
				// fund the funder and vesting accounts from Bankkeeper
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, balances)
				suite.Require().NoError(err)
				err = testutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr, balances)
				suite.Require().NoError(err)

				msg := types.NewMsgCreateClawbackVestingAccount(funderAddr, vestingAddr, false)
				_, err = suite.app.VestingKeeper.CreateClawbackVestingAccount(s.ctx, msg)
				suite.Require().NoError(err, "failed to create vesting account")
			},
			funder:      funderAddr,
			vestingAddr: vestingAddr,
			expPass:     false,
			errContains: "is already a clawback vesting account",
		},
		{
			name: "fail - vesting address is in the blocked addresses list",
			malleate: func(funder sdk.AccAddress, vestingAddr sdk.AccAddress) {
				// fund the funder and vesting accounts from Bankkeeper
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, balances)
				suite.Require().NoError(err)
			},
			funder:      funderAddr,
			vestingAddr: authtypes.NewModuleAddress("distribution"),
			expPass:     false,
			errContains: "is a blocked address and cannot be converted in a clawback vesting account",
		},
		{
			name: "success",
			malleate: func(funder sdk.AccAddress, vestingAddr sdk.AccAddress) {
				// fund the funder and vesting accounts from Bankkeeper
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, balances)
				suite.Require().NoError(err)
				err = testutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr, balances)
				suite.Require().NoError(err)
			},
			funder:      funderAddr,
			vestingAddr: vestingAddr,
			expPass:     true,
		},
	}

	for _, tc := range testcases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest() // Reset
			ctx := sdk.WrapSDKContext(suite.ctx)

			tc.malleate(tc.funder, tc.vestingAddr)

			msg := types.NewMsgCreateClawbackVestingAccount(tc.funder, tc.vestingAddr, false)
			res, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(&types.MsgCreateClawbackVestingAccountResponse{}, res)

				accI := suite.app.AccountKeeper.GetAccount(suite.ctx, tc.vestingAddr)
				suite.Require().NotNil(accI, "expected account to be created")
				suite.Require().IsType(&types.ClawbackVestingAccount{}, accI, "expected account to be a clawback vesting account")
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.errContains)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgClawback() {
	testCases := []struct {
		name        string
		malleate    func()
		funder      sdk.AccAddress
		vestingAddr sdk.AccAddress
		// clawbackDest is the address to send the coins that were clawed back to
		clawbackDest sdk.AccAddress
		// initClawback determines if the clawback account should be created during the test setup
		initClawback bool
		// initVesting determines if the vesting account should be created during the test setup
		initVesting bool
		startTime   time.Time
		expPass     bool
		errContains string
	}{
		{
			name:        "fail - account does not exist",
			malleate:    func() {},
			funder:      funder,
			vestingAddr: sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			startTime:   suite.ctx.BlockTime(),
			expPass:     false,
			errContains: "does not exist",
		},
		{
			name:        "fail - no clawback account",
			malleate:    func() {},
			funder:      funder,
			vestingAddr: vestingAddr,
			startTime:   suite.ctx.BlockTime(),
			expPass:     false,
			errContains: types.ErrNotSubjectToClawback.Error(),
		},
		{
			name: "fail - wrong account type",
			malleate: func() {
				// create a base vesting account instead of a clawback vesting account at the vesting address
				baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
				acc := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)
				s.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			funder:       funder,
			vestingAddr:  vestingAddr,
			clawbackDest: addr3,
			startTime:    suite.ctx.BlockTime(),
			expPass:      false,
			errContains:  types.ErrNotSubjectToClawback.Error(),
		},
		{
			name:         "fail - clawback vesting account has no vesting or lockup periods (not funded yet)",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			startTime:    suite.ctx.BlockTime(),
			initClawback: true,
			initVesting:  false,
			expPass:      false,
			errContains:  "has no vesting or lockup periods",
		},
		{
			name:         "fail - wrong funder",
			malleate:     func() {},
			funder:       addr3,
			vestingAddr:  vestingAddr,
			clawbackDest: addr3,
			startTime:    suite.ctx.BlockTime(),
			initClawback: true,
			initVesting:  true,
			expPass:      false,
			errContains:  "clawback can only be requested by original funder",
		},
		{
			name:         "fail - clawback destination is blocked",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			clawbackDest: authtypes.NewModuleAddress("transfer"),
			startTime:    suite.ctx.BlockTime(),
			initClawback: true,
			initVesting:  true,
			expPass:      false,
			errContains:  "is a blocked address and not allowed to receive funds",
		},
		{
			name:         "pass - before start time",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			startTime:    suite.ctx.BlockTime().Add(time.Hour),
			initClawback: true,
			initVesting:  true,
			expPass:      true,
		},
		{
			name:         "pass - with clawback destination",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			clawbackDest: addr3,
			startTime:    suite.ctx.BlockTime(),
			initClawback: true,
			initVesting:  true,
			expPass:      true,
		},
		{
			name:         "pass - without clawback destination",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			startTime:    suite.ctx.BlockTime(),
			initClawback: true,
			initVesting:  true,
			expPass:      true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)

			// fund the vesting target address to initialize it as an account and
			// then send all funds to the funder account
			err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, vestingAddr, balances)
			suite.Require().NoError(err, "failed to fund target account")
			err = suite.app.BankKeeper.SendCoins(suite.ctx, vestingAddr, funder, balances)
			suite.Require().NoError(err, "failed to send coins to funder account")

			// Create Clawback Vesting Account
			if tc.initClawback {
				createMsg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, true)
				createRes, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, createMsg)
				suite.Require().NoError(err)
				suite.Require().NotNil(createRes)
			}

			// Fund vesting account
			if tc.initVesting {
				fundMsg := types.NewMsgFundVestingAccount(funder, vestingAddr, tc.startTime, lockupPeriods, vestingPeriods)
				fundRes, err := suite.app.VestingKeeper.FundVestingAccount(ctx, fundMsg)
				suite.Require().NoError(err)
				suite.Require().NotNil(fundRes)

				balanceVestingAcc := suite.app.BankKeeper.GetBalance(suite.ctx, vestingAddr, "test")
				suite.Require().Equal(balanceVestingAcc, sdk.NewInt64Coin("test", 1000))
			}

			tc.malleate()

			// Perform clawback
			msg := types.NewMsgClawback(tc.funder, tc.vestingAddr, tc.clawbackDest)
			res, err := suite.app.VestingKeeper.Clawback(ctx, msg)

			balanceVestingAcc := suite.app.BankKeeper.GetBalance(suite.ctx, vestingAddr, "test")
			balanceClaw := suite.app.BankKeeper.GetBalance(suite.ctx, tc.clawbackDest, "test")
			if len(tc.clawbackDest) == 0 {
				balanceClaw = suite.app.BankKeeper.GetBalance(suite.ctx, tc.funder, "test")
			}

			if tc.expPass {
				suite.Require().NoError(err)

				expRes := &types.MsgClawbackResponse{Coins: balances}
				suite.Require().Equal(expRes, res, "expected full balances to be clawed back")
				suite.Require().Equal(sdk.NewInt64Coin("test", 0), balanceVestingAcc)
				suite.Require().Equal(balances[0], balanceClaw)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.errContains)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgUpdateVestingFunder() {
	newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	testCases := []struct {
		name       string
		malleate   func()
		funder     sdk.AccAddress
		vestingAcc sdk.AccAddress
		newFunder  sdk.AccAddress
		// initClawback determines if the clawback vesting account should be initialized for the test case
		initClawback bool
		expPass      bool
		errContains  string
	}{
		{
			name:         "fail - non-existent account",
			malleate:     func() {},
			funder:       funder,
			vestingAcc:   sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			newFunder:    newFunder,
			initClawback: false,
			expPass:      false,
			errContains:  "does not exist",
		},
		{
			name: "fail - wrong account type",
			malleate: func() {
				baseAccount := authtypes.NewBaseAccountWithAddress(addr4)
				acc := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)
				s.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			funder:       funder,
			vestingAcc:   vestingAddr,
			newFunder:    newFunder,
			initClawback: false,
			expPass:      false,
			errContains:  types.ErrNotSubjectToClawback.Error(),
		},
		{
			name:         "fail - wrong funder",
			malleate:     func() {},
			funder:       newFunder,
			vestingAcc:   vestingAddr,
			newFunder:    newFunder,
			initClawback: true,
			expPass:      false,
			errContains:  "is not the current funder and cannot update the funder address",
		},
		{
			name:         "fail - new funder is blocked",
			malleate:     func() {},
			funder:       funder,
			vestingAcc:   vestingAddr,
			newFunder:    authtypes.NewModuleAddress("transfer"),
			initClawback: true,
			expPass:      false,
			errContains:  "is a blocked address and not allowed to fund vesting accounts",
		},
		{
			name: "pass - update funder successfully",
			malleate: func() {
			},
			funder:       funder,
			vestingAcc:   vestingAddr,
			newFunder:    newFunder,
			initClawback: true,
			expPass:      true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx := sdk.WrapSDKContext(suite.ctx)

			// fund the account at the vesting address to initialize it and then sund all funds to the funder account
			err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, vestingAddr, balances)
			suite.Require().NoError(err)
			err = suite.app.BankKeeper.SendCoins(suite.ctx, vestingAddr, funder, balances)

			// Create Clawback Vesting Account
			if tc.initClawback {
				createMsg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, false)
				createRes, err := suite.app.VestingKeeper.CreateClawbackVestingAccount(ctx, createMsg)
				suite.Require().NoError(err)
				suite.Require().NotNil(createRes)
			}

			tc.malleate()

			// Perform Vesting account update
			msg := types.NewMsgUpdateVestingFunder(tc.funder, tc.newFunder, tc.vestingAcc)
			res, err := suite.app.VestingKeeper.UpdateVestingFunder(ctx, msg)

			expRes := &types.MsgUpdateVestingFunderResponse{}

			if tc.expPass {
				// get the updated vesting account
				vestingAcc := suite.app.AccountKeeper.GetAccount(suite.ctx, tc.vestingAcc)
				va, ok := vestingAcc.(*types.ClawbackVestingAccount)
				suite.Require().True(ok, "vesting account could not be casted to ClawbackVestingAccount")

				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(va.FunderAddress, tc.newFunder.String())
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.errContains)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackVestingAccountStore() {
	suite.SetupTest()

	// Create and set clawback vesting account
	vestingStart := s.ctx.BlockTime()
	funder := sdk.AccAddress(types.ModuleName)
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	baseAccount := authtypes.NewBaseAccountWithAddress(addr)
	acc := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	acc2 := suite.app.AccountKeeper.GetAccount(suite.ctx, acc.GetAddress())
	suite.Require().IsType(&types.ClawbackVestingAccount{}, acc2)
	suite.Require().Equal(acc.String(), acc2.String())
}

func (suite *KeeperTestSuite) TestClawbackVestingAccountMarshal() {
	suite.SetupTest()

	// Create and set clawback vesting account
	vestingStart := s.ctx.BlockTime()
	funder := sdk.AccAddress(types.ModuleName)
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	baseAccount := authtypes.NewBaseAccountWithAddress(addr)
	acc := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)

	bz, err := suite.app.AccountKeeper.MarshalAccount(acc)
	suite.Require().NoError(err)

	acc2, err := suite.app.AccountKeeper.UnmarshalAccount(bz)
	suite.Require().NoError(err)
	suite.Require().IsType(&types.ClawbackVestingAccount{}, acc2)
	suite.Require().Equal(acc.String(), acc2.String())

	// error on bad bytes
	_, err = suite.app.AccountKeeper.UnmarshalAccount(bz[:len(bz)/2])
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestConvertVestingAccount() {
	startTime := s.ctx.BlockTime().Add(-5 * time.Second)
	testCases := []struct {
		name     string
		malleate func() authtypes.AccountI
		expPass  bool
	}{
		{
			"fail - no account found",
			func() authtypes.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				return baseAcc
			},
			false,
		},
		{
			"fail - not a vesting account",
			func() authtypes.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)
				return baseAcc
			},
			false,
		},
		{
			"fail - unlocked & unvested",
			func() authtypes.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				lockupPeriods := sdkvesting.Periods{{Length: 0, Amount: balances}}
				vestingPeriods := sdkvesting.Periods{
					{Length: 0, Amount: quarter},
					{Length: 2000, Amount: quarter},
					{Length: 2000, Amount: quarter},
					{Length: 2000, Amount: quarter},
				}
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, startTime, lockupPeriods, vestingPeriods)
				suite.app.AccountKeeper.SetAccount(suite.ctx, vestingAcc)
				return vestingAcc
			},
			false,
		},
		{
			"fail - locked & vested",
			func() authtypes.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				vestingPeriods := sdkvesting.Periods{{Length: 0, Amount: balances}}
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, startTime, lockupPeriods, vestingPeriods)
				suite.app.AccountKeeper.SetAccount(suite.ctx, vestingAcc)
				return vestingAcc
			},
			false,
		},
		{
			"fail - locked & unvested",
			func() authtypes.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, suite.ctx.BlockTime(), lockupPeriods, vestingPeriods)
				suite.app.AccountKeeper.SetAccount(suite.ctx, vestingAcc)
				return vestingAcc
			},
			false,
		},
		{
			"success - unlocked & vested convert to base account",
			func() authtypes.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				vestingPeriods := sdkvesting.Periods{{Length: 0, Amount: balances}}
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, startTime, nil, vestingPeriods)
				suite.app.AccountKeeper.SetAccount(suite.ctx, vestingAcc)
				return vestingAcc
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			acc := tc.malleate()

			msg := types.NewMsgConvertVestingAccount(acc.GetAddress())
			res, err := suite.app.VestingKeeper.ConvertVestingAccount(sdk.WrapSDKContext(suite.ctx), msg)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				account := suite.app.AccountKeeper.GetAccount(suite.ctx, acc.GetAddress())

				_, ok := account.(vestingexported.VestingAccount)
				suite.Require().False(ok)

				_, ok = account.(evmostypes.EthAccountI)
				suite.Require().True(ok)

			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}
