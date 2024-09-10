package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/testutil"
	evmosfactory "github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/utils"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/evmos/evmos/v20/x/vesting/types"
)

var (
	vestAmount     = int64(1000)
	balances       = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseDenom, vestAmount))
	quarter        = sdk.NewCoins(sdk.NewInt64Coin(utils.BaseDenom, 250))
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

func TestMsgFundVestingAccount(t *testing.T) {
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
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			nw := network.NewUnitTestNetwork()
			ctx := nw.GetContext()

			// fund the recipient account to set the account and then
			// send funds over to the funder account so balance is empty
			err := testutil.FundAccount(ctx, nw.App.BankKeeper, tc.vestingAddr, balances)
			require.NoError(t, err, "failed to fund target account")
			err = nw.App.BankKeeper.SendCoins(ctx, tc.vestingAddr, tc.funder, balances)
			require.NoError(t, err, "failed to send tokens to funder account")

			// create a clawback vesting account if necessary
			if tc.initClawback {
				msgCreate := types.NewMsgCreateClawbackVestingAccount(tc.funder, tc.vestingAddr, false)
				resCreate, err := nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msgCreate)
				require.NoError(t, err)
				require.Equal(t, &types.MsgCreateClawbackVestingAccountResponse{}, resCreate)
			}

			// fund the vesting account prior to actual test if desired
			if tc.preFundClawback {
				// in order to fund the vesting account additionally to the actual main test case, we need to
				// send it some more funds
				err = testutil.FundAccount(ctx, nw.App.BankKeeper, tc.funder, balances)
				require.NoError(t, err, "failed to fund funder account")
				// fund vesting account
				msgFund := types.NewMsgFundVestingAccount(tc.funder, tc.vestingAddr, time.Now(), lockupPeriods, vestingPeriods)
				_, err = nw.App.VestingKeeper.FundVestingAccount(ctx, msgFund)
				require.NoError(t, err, "failed to fund vesting account")
			}

			// fund the vesting account
			msg := types.NewMsgFundVestingAccount(
				tc.funder,
				tc.vestingAddr,
				time.Now(),
				tc.lockup,
				tc.vesting,
			)

			res, err := nw.App.VestingKeeper.FundVestingAccount(ctx, msg)

			expRes := &types.MsgFundVestingAccountResponse{}
			balanceFunder := nw.App.BankKeeper.GetBalance(ctx, tc.funder, utils.BaseDenom)
			balanceVestingAddr := nw.App.BankKeeper.GetBalance(ctx, tc.vestingAddr, utils.BaseDenom)

			if tc.expPass {
				require.NoError(t, err, tc.name)
				require.Equal(t, expRes, res)

				accI := nw.App.AccountKeeper.GetAccount(ctx, tc.vestingAddr)
				require.NotNil(t, accI)
				require.IsType(t, &types.ClawbackVestingAccount{}, accI)
				require.Equal(t, sdk.NewInt64Coin(utils.BaseDenom, 0), balanceFunder)
				require.Equal(t, sdk.NewInt64Coin(utils.BaseDenom, vestAmount+tc.expectExtraBalance), balanceVestingAddr)
			} else {
				require.Error(t, err, tc.name)
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

// NOTE: This function tests cases which require a different setup than the standard
// cases in TestMsgFundVestingAccount.
func TestMsgFundVestingAccountSpecialCases(t *testing.T) {
	// ---------------------------
	// Test blocked address
	t.Run("fail - blocked address", func(t *testing.T) {
		nw := network.NewUnitTestNetwork()
		ctx := nw.GetContext()

		msg := &types.MsgFundVestingAccount{
			FunderAddress:  funder.String(),
			VestingAddress: authtypes.NewModuleAddress("transfer").String(),
			StartTime:      time.Now(),
			LockupPeriods:  lockupPeriods,
			VestingPeriods: vestingPeriods,
		}

		_, err := nw.App.VestingKeeper.FundVestingAccount(ctx, msg)
		require.Error(t, err, "expected blocked address error")
		require.ErrorContains(t, err, "is not allowed to receive funds")
	})

	// ---------------------------
	// Test wrong funder by first creating a clawback vesting account
	// and then trying to fund it with a different funder
	t.Run("fail - wrong funder", func(t *testing.T) {
		nw := network.NewUnitTestNetwork()
		ctx := nw.GetContext()

		// fund the recipient account to set the account
		err := testutil.FundAccount(ctx, nw.App.BankKeeper, vestingAddr, balances)
		require.NoError(t, err, "failed to fund target account")
		msgCreate := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, false)
		_, err = nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msgCreate)
		require.NoError(t, err, "failed to create clawback vesting account")

		msg := &types.MsgFundVestingAccount{
			FunderAddress:  addr3.String(),
			VestingAddress: vestingAddr.String(),
			StartTime:      time.Now(),
			LockupPeriods:  lockupPeriods,
			VestingPeriods: vestingPeriods,
		}
		_, err = nw.App.VestingKeeper.FundVestingAccount(ctx, msg)
		require.Error(t, err, "expected wrong funder error")
		require.ErrorContains(t, err, fmt.Sprintf("%s can only accept grants from account %s", vestingAddr, funder))
	})
}

func TestMsgCreateClawbackVestingAccount(t *testing.T) {
	var (
		ctx     sdk.Context
		nw      *network.UnitTestNetwork
		handler grpc.Handler
		factory evmosfactory.TxFactory
	)
	funderAddr, funderPriv := utiltx.NewAccAddressAndKey()
	vestingAddr, _ := utiltx.NewAccAddressAndKey()

	testcases := []struct {
		name        string
		malleate    func(funder sdk.AccAddress) sdk.AccAddress
		funder      sdk.AccAddress
		expPass     bool
		errContains string
	}{
		{
			name: "fail - account does not exist",
			malleate: func(sdk.AccAddress) sdk.AccAddress {
				return vestingAddr
			},
			funder:      funderAddr,
			expPass:     false,
			errContains: fmt.Sprintf("account %s does not exist", vestingAddr),
		},
		{
			name: "fail - account is a smart contract",
			malleate: func(_ sdk.AccAddress) sdk.AccAddress {
				contractAddr, err := factory.DeployContract(
					funderPriv,
					evmtypes.EvmTxArgs{},
					evmosfactory.ContractDeploymentData{
						Contract:        contracts.ERC20MinterBurnerDecimalsContract,
						ConstructorArgs: []interface{}{"TestToken", "TTK", uint8(18)},
					},
				)
				require.NoError(t, err)
				require.NoError(t, nw.NextBlock())
				ctx = nw.GetContext()

				return utils.EthToCosmosAddr(contractAddr)
			},
			funder:      funderAddr,
			expPass:     false,
			errContains: "is a contract account and cannot be converted in a clawback vesting account",
		},
		{
			name: "fail - vesting account already exists",
			malleate: func(funder sdk.AccAddress) sdk.AccAddress {
				// fund the funder and vesting accounts from Bankkeeper
				err := testutil.FundAccount(ctx, nw.App.BankKeeper, funder, balances)
				require.NoError(t, err)
				err = testutil.FundAccount(ctx, nw.App.BankKeeper, vestingAddr, balances)
				require.NoError(t, err)

				msg := types.NewMsgCreateClawbackVestingAccount(funderAddr, vestingAddr, false)
				_, err = nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
				require.NoError(t, err, "failed to create vesting account")
				return vestingAddr
			},
			funder:      funderAddr,
			expPass:     false,
			errContains: "is already a clawback vesting account",
		},
		{
			name: "fail - vesting address is in the blocked addresses list",
			malleate: func(funder sdk.AccAddress) sdk.AccAddress {
				// fund the funder and vesting accounts from Bankkeeper
				err := testutil.FundAccount(ctx, nw.App.BankKeeper, funder, balances)
				require.NoError(t, err)
				return authtypes.NewModuleAddress("distribution")
			},
			funder:      funderAddr,
			expPass:     false,
			errContains: "is a blocked address and cannot be converted in a clawback vesting account",
		},
		{
			name: "success",
			malleate: func(funder sdk.AccAddress) sdk.AccAddress {
				// fund the funder and vesting accounts from Bankkeeper
				err := testutil.FundAccount(ctx, nw.App.BankKeeper, funder, balances)
				require.NoError(t, err)
				err = testutil.FundAccount(ctx, nw.App.BankKeeper, vestingAddr, balances)
				require.NoError(t, err)

				return vestingAddr
			},
			funder:  funderAddr,
			expPass: true,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork(network.WithPreFundedAccounts(funderAddr))
			handler = grpc.NewIntegrationHandler(nw)
			factory = evmosfactory.New(nw, handler)
			ctx = nw.GetContext()

			vestingAddr := tc.malleate(tc.funder)

			msg := types.NewMsgCreateClawbackVestingAccount(tc.funder, vestingAddr, false)
			res, err := nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)

			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, &types.MsgCreateClawbackVestingAccountResponse{}, res)

				accI := nw.App.AccountKeeper.GetAccount(ctx, vestingAddr)
				require.NotNil(t, accI, "expected account to be created")
				require.IsType(t, &types.ClawbackVestingAccount{}, accI, "expected account to be a clawback vesting account")
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

func TestMsgClawback(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	now := time.Now()
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
			name: "fail - account does not exist",
			malleate: func() {
				vestingAddr = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			},
			funder:      funder,
			vestingAddr: sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
			startTime:   now,
			expPass:     false,
			errContains: "does not exist",
		},
		{
			name:        "fail - no clawback account",
			malleate:    func() {},
			funder:      funder,
			vestingAddr: vestingAddr,
			startTime:   now,
			expPass:     false,
			errContains: types.ErrNotSubjectToClawback.Error(),
		},
		{
			name: "fail - wrong account type",
			malleate: func() {
				// create a base vesting account instead of a clawback vesting account at the vesting address
				baseAccount := authtypes.NewBaseAccountWithAddress(vestingAddr)
				baseAccount.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				acc, err := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)
				require.NoError(t, err)
				nw.App.AccountKeeper.SetAccount(ctx, acc)
			},
			funder:       funder,
			vestingAddr:  vestingAddr,
			clawbackDest: addr3,
			startTime:    now,
			expPass:      false,
			errContains:  types.ErrNotSubjectToClawback.Error(),
		},
		{
			name:         "fail - clawback vesting account has no vesting or lockup periods (not funded yet)",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			startTime:    now,
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
			startTime:    now,
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
			startTime:    now,
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
			startTime:    now.Add(time.Hour),
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
			startTime:    now,
			initClawback: true,
			initVesting:  true,
			expPass:      true,
		},
		{
			name:         "pass - without clawback destination",
			malleate:     func() {},
			funder:       funder,
			vestingAddr:  vestingAddr,
			startTime:    now,
			initClawback: true,
			initVesting:  true,
			expPass:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			vestingAddr = tc.vestingAddr

			// fund the vesting target address to initialize it as an account and
			// then send all funds to the funder account
			err := testutil.FundAccount(ctx, nw.App.BankKeeper, vestingAddr, balances)
			require.NoError(t, err, "failed to fund target account")
			err = nw.App.BankKeeper.SendCoins(ctx, vestingAddr, funder, balances)
			require.NoError(t, err, "failed to send coins to funder account")

			// Create Clawback Vesting Account
			if tc.initClawback {
				createMsg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, true)
				createRes, err := nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, createMsg)
				require.NoError(t, err)
				require.NotNil(t, createRes)
			}

			// Fund vesting account
			if tc.initVesting {
				fundMsg := types.NewMsgFundVestingAccount(funder, vestingAddr, tc.startTime, lockupPeriods, vestingPeriods)
				fundRes, err := nw.App.VestingKeeper.FundVestingAccount(ctx, fundMsg)
				require.NoError(t, err)
				require.NotNil(t, fundRes)

				balanceVestingAcc := nw.App.BankKeeper.GetBalance(ctx, vestingAddr, utils.BaseDenom)
				require.Equal(t, balanceVestingAcc, sdk.NewInt64Coin(utils.BaseDenom, 1000))
			}

			tc.malleate()

			// Perform clawback
			msg := types.NewMsgClawback(tc.funder, vestingAddr, tc.clawbackDest)
			res, err := nw.App.VestingKeeper.Clawback(ctx, msg)

			balanceVestingAcc := nw.App.BankKeeper.GetBalance(ctx, vestingAddr, utils.BaseDenom)
			balanceClaw := nw.App.BankKeeper.GetBalance(ctx, tc.clawbackDest, utils.BaseDenom)
			if len(tc.clawbackDest) == 0 {
				balanceClaw = nw.App.BankKeeper.GetBalance(ctx, tc.funder, utils.BaseDenom)
			}

			if tc.expPass {
				require.NoError(t, err)

				expRes := &types.MsgClawbackResponse{Coins: balances}
				require.Equal(t, expRes, res, "expected full balances to be clawed back")
				require.Equal(t, sdk.NewInt64Coin(utils.BaseDenom, 0), balanceVestingAcc)
				require.Equal(t, balances[0], balanceClaw)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errContains)
				require.Nil(t, res)
			}
		})
	}
}

func TestMsgUpdateVestingFunder(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
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
				baseAccount.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				acc, err := sdkvesting.NewBaseVestingAccount(baseAccount, balances, 500000)
				require.NoError(t, err)
				nw.App.AccountKeeper.SetAccount(ctx, acc)
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
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			// fund the account at the vesting address to initialize it and then sund all funds to the funder account
			err := testutil.FundAccount(ctx, nw.App.BankKeeper, vestingAddr, balances)
			require.NoError(t, err)
			err = nw.App.BankKeeper.SendCoins(ctx, vestingAddr, funder, balances)
			require.NoError(t, err)

			// Create Clawback Vesting Account
			if tc.initClawback {
				createMsg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, false)
				createRes, err := nw.App.VestingKeeper.CreateClawbackVestingAccount(ctx, createMsg)
				require.NoError(t, err)
				require.NotNil(t, createRes)
			}

			tc.malleate()

			// Perform Vesting account update
			msg := types.NewMsgUpdateVestingFunder(tc.funder, tc.newFunder, tc.vestingAcc)
			res, err := nw.App.VestingKeeper.UpdateVestingFunder(ctx, msg)

			expRes := &types.MsgUpdateVestingFunderResponse{}

			if tc.expPass {
				// get the updated vesting account
				vestingAcc := nw.App.AccountKeeper.GetAccount(ctx, tc.vestingAcc)
				va, ok := vestingAcc.(*types.ClawbackVestingAccount)
				require.True(t, ok, "vesting account could not be casted to ClawbackVestingAccount")

				require.NoError(t, err)
				require.Equal(t, expRes, res)
				require.Equal(t, va.FunderAddress, tc.newFunder.String())
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errContains)
				require.Nil(t, res)
			}
		})
	}
}

func TestClawbackVestingAccountStore(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	// Create and set clawback vesting account
	vestingStart := ctx.BlockTime()
	funder := sdk.AccAddress(types.ModuleName)
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	baseAccount := authtypes.NewBaseAccountWithAddress(addr)
	baseAccount.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
	acc := types.NewClawbackVestingAccount(baseAccount, funder, balances, vestingStart, lockupPeriods, vestingPeriods)
	nw.App.AccountKeeper.SetAccount(ctx, acc)

	acc2 := nw.App.AccountKeeper.GetAccount(ctx, acc.GetAddress())
	require.IsType(t, &types.ClawbackVestingAccount{}, acc2)
	require.Equal(t, acc.String(), acc2.String())
}

func TestConvertVestingAccount(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
	now := time.Now()
	startTime := now.Add(-5 * time.Second)
	testCases := []struct {
		name     string
		malleate func() sdk.AccountI
		expPass  bool
	}{
		{
			"fail - no account found",
			func() sdk.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				return baseAcc
			},
			false,
		},
		{
			"fail - not a vesting account",
			func() sdk.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				baseAcc.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				nw.App.AccountKeeper.SetAccount(ctx, baseAcc)
				return baseAcc
			},
			false,
		},
		{
			"fail - unlocked & unvested",
			func() sdk.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				baseAcc.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				lockupPeriods := sdkvesting.Periods{{Length: 0, Amount: balances}}
				vestingPeriods := sdkvesting.Periods{
					{Length: 0, Amount: quarter},
					{Length: 2000, Amount: quarter},
					{Length: 2000, Amount: quarter},
					{Length: 2000, Amount: quarter},
				}
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, startTime, lockupPeriods, vestingPeriods)
				nw.App.AccountKeeper.SetAccount(ctx, vestingAcc)
				return vestingAcc
			},
			false,
		},
		{
			"fail - locked & vested",
			func() sdk.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				vestingPeriods := sdkvesting.Periods{{Length: 0, Amount: balances}}
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				baseAcc.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, startTime, lockupPeriods, vestingPeriods)
				nw.App.AccountKeeper.SetAccount(ctx, vestingAcc)
				return vestingAcc
			},
			false,
		},
		{
			"fail - locked & unvested",
			func() sdk.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				baseAcc.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, ctx.BlockTime(), lockupPeriods, vestingPeriods)
				nw.App.AccountKeeper.SetAccount(ctx, vestingAcc)
				return vestingAcc
			},
			false,
		},
		{
			"success - unlocked & vested convert to base account",
			func() sdk.AccountI {
				from, priv := utiltx.NewAccAddressAndKey()
				baseAcc := authtypes.NewBaseAccount(from, priv.PubKey(), 1, 5)
				baseAcc.AccountNumber = nw.App.AccountKeeper.NextAccountNumber(ctx)
				vestingPeriods := sdkvesting.Periods{{Length: 0, Amount: balances}}
				vestingAcc := types.NewClawbackVestingAccount(baseAcc, from, balances, startTime, nil, vestingPeriods)
				nw.App.AccountKeeper.SetAccount(ctx, vestingAcc)
				return vestingAcc
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			acc := tc.malleate()

			msg := types.NewMsgConvertVestingAccount(acc.GetAddress())
			res, err := nw.App.VestingKeeper.ConvertVestingAccount(ctx, msg)

			if tc.expPass {
				require.NoError(t, err)
				require.NotNil(t, res)

				account := nw.App.AccountKeeper.GetAccount(ctx, acc.GetAddress())

				_, ok := account.(vestingexported.VestingAccount)
				require.False(t, ok)

				_, ok = account.(*authtypes.BaseAccount)
				require.True(t, ok)

			} else {
				require.Error(t, err)
				require.Nil(t, res)
			}
		})
	}
}
