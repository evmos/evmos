package keeper_test

import (
	"math/big"
	"strings"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/evmos/evmos/v13/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v13/testutil"
	utiltx "github.com/evmos/evmos/v13/testutil/tx"
	"github.com/evmos/evmos/v13/utils"
	"github.com/evmos/evmos/v13/x/vesting/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestClawbackAccount is a struct to store all relevant information that is corresponding
// to a clawback vesting account.
type TestClawbackAccount struct {
	privKey         *ethsecp256k1.PrivKey
	address         sdk.AccAddress
	clawbackAccount *types.ClawbackVestingAccount
}

// Initialize general error variable for easier handling in loops throughout this test suite.
var err error

// Clawback vesting with Cliff and Lock. In this case the cliff is reached
// before the lockup period is reached to represent the scenario in which an
// employee starts before mainnet launch (periodsCliff < lockupPeriod)
//
// Example:
// 21/10 Employee joins Evmos and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
var _ = Describe("Clawback Vesting Accounts", Ordered, func() {
	// Monthly vesting period
	stakeDenom := utils.BaseDenom
	amt := sdk.NewInt(1e17)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12 month lockup
	lockup := int64(12) // 12 months
	lockupLength := vestingLength * lockup
	// Unlock at 12 and 24 months
	numLockupPeriods := int64(2)
	// Unlock 1/4th of the total vest in each unlock event. By default, all tokens are
	// unlocked after surpassing the final period.
	unlockedPerLockup := vestingAmtTotal.QuoInt(math.NewInt(4))
	unlockedPerLockupAmt := unlockedPerLockup[0].Amount
	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: unlockedPerLockup}
	lockupPeriods := make(sdkvesting.Periods, numLockupPeriods)
	for i := range lockupPeriods {
		lockupPeriods[i] = lockupPeriod
	}

	// Create vesting periods with initial cliff
	vestingPeriods := sdkvesting.Periods{cliffPeriod}
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	// Create test accounts with private keys for signing
	numTestAccounts := 4
	testAccounts := make([]TestClawbackAccount, numTestAccounts)
	for i := range testAccounts {
		address, privKey := utiltx.NewAddrKey()
		testAccounts[i] = TestClawbackAccount{
			privKey: privKey,
			address: address.Bytes(),
		}
	}
	numTestMsgs := 3

	accountGasCoverage := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(1e16)))

	var (
		clawbackAccount   *types.ClawbackVestingAccount
		unvested          sdk.Coins
		vested            sdk.Coins
		twoThirdsOfVested sdk.Coins
	)

	dest := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	funder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	BeforeEach(func() {
		s.SetupTest()

		// Initialize all test accounts
		for i, account := range testAccounts {
			// Create and fund periodic vesting account
			vestingStart := s.ctx.BlockTime()
			baseAccount := authtypes.NewBaseAccountWithAddress(account.address)
			clawbackAccount = types.NewClawbackVestingAccount(
				baseAccount,
				funder,
				vestingAmtTotal,
				vestingStart,
				lockupPeriods,
				vestingPeriods,
			)

			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, account.address, vestingAmtTotal)
			s.Require().NoError(err)
			acc := s.app.AccountKeeper.NewAccount(s.ctx, clawbackAccount)
			s.app.AccountKeeper.SetAccount(s.ctx, acc)

			// Check if all tokens are unvested at vestingStart
			unvested = clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			s.Require().Equal(vestingAmtTotal, unvested)
			s.Require().True(vested.IsZero())

			// Grant gas stipend to cover EVM fees
			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, clawbackAccount.GetAddress(), accountGasCoverage)
			s.Require().NoError(err)
			granteeBalance := s.app.BankKeeper.GetBalance(s.ctx, account.address, stakeDenom)
			s.Require().Equal(granteeBalance, accountGasCoverage[0].Add(vestingAmtTotal[0]))

			// Update testAccounts clawbackAccount reference
			testAccounts[i].clawbackAccount = clawbackAccount
		}
	})

	Context("before first vesting period", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			s.Commit()

			// Ensure no tokens are vested
			vested := clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			unlocked := clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.ZeroInt()))
			s.Require().Equal(zeroCoins, vested)
			s.Require().Equal(zeroCoins, unlocked)
		})

		It("cannot delegate tokens", func() {
			_, err := delegate(testAccounts[0], accountGasCoverage.Add(sdk.NewCoin(stakeDenom, math.NewInt(1))))
			Expect(err).ToNot(BeNil())
		})

		It("can transfer spendable tokens", func() {
			account := testAccounts[0]
			// Fund account with new spendable tokens
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, account.address, unvested)
			Expect(err).To(BeNil())

			err = s.app.BankKeeper.SendCoins(
				s.ctx,
				account.address,
				dest,
				unvested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				dest,
				unvested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := testAccounts[0]
			// Fund account with new spendable tokens
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, account.address, unlockedPerLockup)
			Expect(err).To(BeNil())

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("cannot perform Ethereum tx with unvested balance", func() {
			account := testAccounts[0]
			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})
	})

	Context("after first vesting period and before lockup", func() {
		BeforeEach(func() {
			// Surpass cliff but none of lockup duration
			cliffDuration := time.Duration(cliffLength)
			s.CommitAfter(cliffDuration * time.Second)

			// Check if some, but not all tokens are vested
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)

			twoThirdsOfVested = vested.Sub(vested.QuoInt(sdk.NewInt(3))...)
		})

		It("can delegate vested tokens and update spendable balance", func() {
			testAccount := testAccounts[0]
			// Verify that the total spendable coins decreases after staking
			// vested tokens.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, testAccount.address)

			_, err := delegate(testAccount, vested)
			Expect(err).To(BeNil())

			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, testAccount.address)
			Expect(spendablePost.AmountOf(stakeDenom).GT(spendablePre.AmountOf(stakeDenom)))
		})

		It("cannot delegate unvested tokens", func() {
			_, err := delegate(testAccounts[0], vestingAmtTotal)
			Expect(err).ToNot(BeNil())
		})

		It("cannot delegate unvested tokens in batches", func() {
			msg, err := delegate(testAccounts[0], twoThirdsOfVested)
			Expect(err).To(BeNil())

			msgServer := stakingkeeper.NewMsgServerImpl(s.app.StakingKeeper)
			_, err = msgServer.Delegate(s.ctx, msg)
			Expect(err).ToNot(HaveOccurred(), "error while executing the delegate message")

			_, err = delegate(testAccounts[0], twoThirdsOfVested)
			Expect(err).ToNot(BeNil())
		})

		It("cannot delegate then send tokens", func() {
			_, err := delegate(testAccounts[0], twoThirdsOfVested)
			Expect(err).To(BeNil())

			err = s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				dest,
				twoThirdsOfVested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				dest,
				vested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := testAccounts[0]
			// Fund account with new spendable tokens
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, account.address, unlockedPerLockup)
			Expect(err).To(BeNil())

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("cannot perform Ethereum tx with locked balance", func() {
			account := testAccounts[0]
			txAmount := vested.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})
	})

	Context("Between first and second lockup periods", func() {
		BeforeEach(func() {
			// Surpass first lockup
			vestDuration := time.Duration(lockupLength)
			s.CommitAfter(vestDuration * time.Second)

			// Check if some, but not all tokens are vested and unlocked
			for _, account := range testAccounts {
				vested := account.clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
				unlocked := account.clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup))))

				s.Require().NotEqual(vestingAmtTotal, vested)
				s.Require().Equal(expVested, vested)
				s.Require().Equal(unlocked, unlockedPerLockup)
			}
		})

		It("should enable access to unlocked EVM tokens (single-account, single-msg)", func() {
			account := testAccounts[0]

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("should enable access to unlocked EVM tokens (single-account, multiple-msgs)", func() {
			account := testAccounts[0]

			// Split the total unlocked amount into numTestMsgs equally sized tx's
			msgs := make([]sdk.Msg, numTestMsgs)
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()

			for i := 0; i < numTestMsgs; i++ {
				msgs[i], err = utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, i)
				Expect(err).To(BeNil())
			}

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msgs...)
		})

		It("should enable access to unlocked EVM tokens (multi-account, single-msg)", func() {
			txAmount := unlockedPerLockupAmt.BigInt()

			msgs := make([]sdk.Msg, numTestAccounts)
			for i, grantee := range testAccounts {
				msgs[i], err = utiltx.CreateEthTx(s.ctx, s.app, grantee.privKey, grantee.address, dest, txAmount, 0)
				Expect(err).To(BeNil())
			}

			assertEthSucceeds(testAccounts, funder, dest, unlockedPerLockupAmt, stakeDenom, msgs...)
		})

		It("should enable access to unlocked EVM tokens (multi-account, multiple-msgs)", func() {
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()

			for _, grantee := range testAccounts {
				for j := 0; j < numTestMsgs; j++ {
					addedMsg, err := utiltx.CreateEthTx(s.ctx, s.app, grantee.privKey, grantee.address, dest, txAmount, j)
					Expect(err).To(BeNil())
					msgs = append(msgs, addedMsg)
				}
			}

			assertEthSucceeds(testAccounts, funder, dest, unlockedPerLockupAmt, stakeDenom, msgs...)
		})

		It("should not enable access to locked EVM tokens (single-account, single-msg)", func() {
			testAccount := testAccounts[0]
			// Attempt to spend entire vesting balance
			txAmount := vestingAmtTotal.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, testAccount.privKey, testAccount.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})

		It("should not enable access to locked EVM tokens (single-account, multiple-msgs)", func() {
			msgs := make([]sdk.Msg, numTestMsgs+1)
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()
			testAccount := testAccounts[0]

			// Add additional message that exceeds unlocked balance
			for i := 0; i < numTestMsgs+1; i++ {
				msgs[i], err = utiltx.CreateEthTx(s.ctx, s.app, testAccount.privKey, testAccount.address, dest, txAmount, i)
				Expect(err).To(BeNil())
			}

			assertEthFails(msgs...)
		})

		It("should not enable access to locked EVM tokens (multi-account, single-msg)", func() {
			msgs := make([]sdk.Msg, numTestAccounts+1)
			txAmount := unlockedPerLockupAmt.BigInt()

			for i, account := range testAccounts {
				msgs[i], err = utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
				Expect(err).To(BeNil())
			}

			// Add additional message that exceeds unlocked balance
			msgs[numTestAccounts], err = utiltx.CreateEthTx(s.ctx, s.app, testAccounts[0].privKey, testAccounts[0].address, dest, txAmount, 1)
			Expect(err).To(BeNil())

			assertEthFails(msgs...)
		})

		It("should not enable access to locked EVM tokens (multi-account, multiple-msgs)", func() {
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()
			var addedMsg sdk.Msg

			for _, account := range testAccounts {
				for j := 0; j < numTestMsgs; j++ {
					addedMsg, err = utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, j)
					msgs = append(msgs, addedMsg)
				}
			}

			// Add additional message that exceeds unlocked balance
			addedMsg, err = utiltx.CreateEthTx(s.ctx, s.app, testAccounts[0].privKey, testAccounts[0].address, dest, txAmount, numTestMsgs)
			Expect(err).To(BeNil())
			msgs = append(msgs, addedMsg)

			assertEthFails(msgs...)
		})

		It("should not short-circuit with a normal account", func() {
			account := testAccounts[0]
			address, privKey := utiltx.NewAccAddressAndKey()

			txAmount := vestingAmtTotal.AmountOf(stakeDenom).BigInt()

			// Fund a normal account to try to short-circuit the AnteHandler
			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, address, vestingAmtTotal.MulInt(sdk.NewInt(2)))
			Expect(err).To(BeNil())
			normalAccMsg, err := utiltx.CreateEthTx(s.ctx, s.app, privKey, address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			// Attempt to spend entire balance
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())
			err = validateEthVestingTransactionDecorator(normalAccMsg, msg)
			Expect(err).ToNot(BeNil())

			_, err = testutil.DeliverEthTx(s.app, nil, msg)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("after first lockup and additional vest", func() {
		BeforeEach(func() {
			vestDuration := time.Duration(lockupLength + vestingLength)
			s.CommitAfter(vestDuration * time.Second)

			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup+1))))

			unlocked := clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
			expUnlocked := unlockedPerLockup

			s.Require().Equal(expVested, vested)
			s.Require().Equal(expUnlocked, unlocked)
		})

		It("should enable access to unlocked EVM tokens", func() {
			testAccount := testAccounts[0]

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, testAccount.privKey, testAccount.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{testAccount}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("should not enable access to locked EVM tokens", func() {
			testAccount := testAccounts[0]

			txAmount := vested.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, testAccount.privKey, testAccount.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})
	})

	Context("after half of vesting period and both lockups", func() {
		BeforeEach(func() {
			// Surpass lockup duration
			lockupDuration := time.Duration(lockupLength * numLockupPeriods)
			s.CommitAfter(lockupDuration * time.Second)

			// Check if some, but not all tokens are vested
			unvested = clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup*numLockupPeriods))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)
		})

		It("can delegate vested tokens", func() {
			_, err := delegate(testAccounts[0], vested)
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			_, err := delegate(testAccounts[0], vestingAmtTotal)
			Expect(err).ToNot(BeNil())
		})

		It("can transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
				vestingAmtTotal,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := testAccounts[0]

			txAmount := vested.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, vested.AmountOf(stakeDenom), stakeDenom, msg)
		})
	})

	Context("after entire vesting period and both lockups", func() {
		BeforeEach(func() {
			// Surpass vest duration
			vestDuration := time.Duration(vestingLength * periodsTotal)
			s.CommitAfter(vestDuration * time.Second)

			// Check that all tokens are vested and unlocked
			unvested = clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			locked := clawbackAccount.LockedCoins(s.ctx.BlockTime())

			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.ZeroInt()))
			s.Require().Equal(vestingAmtTotal, vested)
			s.Require().Equal(zeroCoins, locked)
			s.Require().Equal(zeroCoins, unvested)
		})

		It("can send entire balance", func() {
			account := testAccounts[0]

			txAmount := vestingAmtTotal.AmountOf(stakeDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount.BigInt(), 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, txAmount, stakeDenom, msg)
		})

		It("cannot exceed balance", func() {
			account := testAccounts[0]

			txAmount := vestingAmtTotal.AmountOf(stakeDenom).Mul(math.NewInt(2))
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, txAmount.BigInt(), 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})

		It("should short-circuit with zero balance", func() {
			account := testAccounts[0]
			balance := s.app.BankKeeper.GetBalance(s.ctx, account.address, stakeDenom)

			// Drain account balance
			err := s.app.BankKeeper.SendCoins(s.ctx, account.address, dest, sdk.NewCoins(balance))
			Expect(err).To(BeNil())

			msg, err := utiltx.CreateEthTx(s.ctx, s.app, account.privKey, account.address, dest, big.NewInt(0), 0)
			Expect(err).To(BeNil())
			err = validateEthVestingTransactionDecorator(msg)
			Expect(err).ToNot(BeNil())
			Expect(strings.Contains(err.Error(), "no balance")).To(BeTrue())
		})
	})
})

// Example:
// 21/10 Employee joins Evmos and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
var _ = Describe("Clawback Vesting Accounts - claw back tokens", func() {
	// Monthly vesting period
	stakeDenom := utils.BaseDenom
	amt := sdk.NewInt(1)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingTotal := amt.Mul(sdk.NewInt(periodsTotal))
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, vestingTotal))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12 month lockup
	lockup := int64(12) // 12 year
	lockupLength := vestingLength * lockup
	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: vestingAmtTotal}
	lockupPeriods := sdkvesting.Periods{lockupPeriod}

	// Create vesting periods with initial cliff
	vestingPeriods := sdkvesting.Periods{cliffPeriod}
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	var (
		clawbackAccount *types.ClawbackVestingAccount
		vesting         sdk.Coins
		vested          sdk.Coins
		unlocked        sdk.Coins
		free            sdk.Coins
		isClawback      bool
	)
	vestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	funder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	dest := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	BeforeEach(func() {
		s.SetupTest()
		vestingStart := s.ctx.BlockTime()

		// Initialize account at vesting address by funding it with tokens
		// and then send them over to the vesting funder
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr, vestingAmtTotal)
		Expect(err).ToNot(HaveOccurred(), "failed to fund target account")
		err = s.app.BankKeeper.SendCoins(s.ctx, vestingAddr, funder, vestingAmtTotal)
		Expect(err).ToNot(HaveOccurred(), "failed to send coins to funder")

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
		Expect(balanceFunder).To(Equal(vestingAmtTotal[0]), "expected different funder balance")
		Expect(balanceGrantee.IsZero()).To(BeTrue(), "expected balance of vesting account to be zero")
		Expect(balanceDest.IsZero()).To(BeTrue(), "expected destination balance to be zero")

		msg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr)

		_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(sdk.WrapSDKContext(s.ctx), msg)
		Expect(err).ToNot(HaveOccurred(), "expected creating clawback vesting account to succeed")

		acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
		clawbackAccount, isClawback = acc.(*types.ClawbackVestingAccount)
		Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

		// fund the vesting account
		msgFund := types.NewMsgFundVestingAccount(funder, vestingAddr, vestingStart, lockupPeriods, vestingPeriods)
		_, err = s.app.VestingKeeper.FundVestingAccount(sdk.WrapSDKContext(s.ctx), msgFund)
		Expect(err).ToNot(HaveOccurred(), "expected funding vesting account to succeed")

		acc = s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
		Expect(acc).ToNot(BeNil(), "expected account to exist")
		clawbackAccount, isClawback = acc.(*types.ClawbackVestingAccount)
		Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

		// Check if all tokens are unvested and locked at vestingStart
		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
		Expect(vesting).To(Equal(vestingAmtTotal), "expected difference vesting tokens")
		Expect(vested.IsZero()).To(BeTrue(), "expected no tokens to be vested")
		Expect(unlocked.IsZero()).To(BeTrue(), "expected no tokens to be unlocked")

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee = s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest = s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		Expect(bF.IsZero()).To(BeTrue(), "expected funder balance to be zero")
		Expect(balanceGrantee).To(Equal(vestingAmtTotal[0]), "expected all tokens to be locked")
		Expect(balanceDest.IsZero()).To(BeTrue(), "expected no tokens to be unlocked")
	})

	It("should fail if there is no vesting or lockup schedule set", func() {
		ctx := sdk.WrapSDKContext(s.ctx)
		emptyVestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, emptyVestingAddr, vestingAmtTotal)
		Expect(err).ToNot(HaveOccurred(), "failed to fund target account")

		msg := types.NewMsgCreateClawbackVestingAccount(funder, emptyVestingAddr)

		_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(sdk.WrapSDKContext(s.ctx), msg)
		Expect(err).ToNot(HaveOccurred(), "expected creating clawback vesting account to succeed")

		clawbackMsg := types.NewMsgClawback(funder, emptyVestingAddr, dest)
		_, err = s.app.VestingKeeper.Clawback(ctx, clawbackMsg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("has no vesting or lockup periods"))
	})

	It("should claw back unvested amount before cliff", func() {
		ctx := sdk.WrapSDKContext(s.ctx)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback before cliff
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		_, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		// All initial vesting amount goes to dest
		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		Expect(bF).To(Equal(balanceFunder), "expected funder balance to be unchanged")
		Expect(bG.IsZero()).To(BeTrue(), "expected all tokens to be clawed back")
		Expect(bD).To(Equal(balanceDest.Add(vestingAmtTotal[0])), "expected all tokens to be clawed back to the destination account")
	})

	It("should claw back any unvested amount after cliff before unlocking", func() {
		// Surpass cliff but not lockup duration
		cliffDuration := time.Duration(cliffLength)
		s.CommitAfter(cliffDuration * time.Second)

		// Check that all tokens are locked and some, but not all tokens are vested
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
		free = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
		expVestedAmount := amt.Mul(sdk.NewInt(cliff))
		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))

		s.Require().Equal(expVested, vested)
		s.Require().True(expVestedAmount.GT(sdk.NewInt(0)))
		s.Require().True(free.IsZero())
		s.Require().Equal(vesting, vestingAmtTotal)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		expClawback := clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())

		// Any unvested amount is clawed back
		s.Require().Equal(balanceFunder, bF)
		s.Require().Equal(balanceGrantee.Sub(expClawback[0]).Amount.Uint64(), bG.Amount.Uint64())
		s.Require().Equal(balanceDest.Add(expClawback[0]).Amount.Uint64(), bD.Amount.Uint64())
	})

	It("should claw back any unvested amount after cliff and unlocking", func() {
		// Surpass lockup duration
		// A strict `if t < clawbackTime` comparison is used in ComputeClawback
		// so, we increment the duration with 1 for the free token calculation to match
		lockupDuration := time.Duration(lockupLength + 1)
		s.CommitAfter(lockupDuration * time.Second)

		// Check if some, but not all tokens are vested and unlocked
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
		free = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
		expVestedAmount := amt.Mul(math.NewInt(lockup))
		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
		unvested := vestingAmtTotal.Sub(vested...)

		s.Require().Equal(free, vested)
		s.Require().Equal(expVested, vested)
		s.Require().True(expVestedAmount.GT(sdk.NewInt(0)))
		s.Require().Equal(vesting, unvested)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Any unvested amount is clawed back
		s.Require().Equal(balanceFunder, bF)
		s.Require().Equal(balanceGrantee.Sub(vesting[0]).Amount.Uint64(), bG.Amount.Uint64())
		s.Require().Equal(balanceDest.Add(vesting[0]).Amount.Uint64(), bD.Amount.Uint64())
	})

	It("should not claw back any amount after vesting periods end", func() {
		// Surpass vesting periods
		vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
		s.CommitAfter(vestingDuration * time.Second)

		// Check if some, but not all tokens are vested and unlocked
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
		free = clawbackAccount.GetVestedCoins(s.ctx.BlockTime())
		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())

		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))
		unvested := vestingAmtTotal.Sub(vested...)

		s.Require().Equal(free, vested)
		s.Require().Equal(expVested, vested)
		s.Require().Equal(expVested, vestingAmtTotal)
		s.Require().Equal(unlocked, vestingAmtTotal)
		s.Require().Equal(vesting, unvested)
		s.Require().True(vesting.IsZero())

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// No amount is clawed back
		s.Require().Equal(balanceFunder, bF)
		s.Require().Equal(balanceGrantee, bG)
		s.Require().Equal(balanceDest, bD)
	})

	It("should update vesting funder and claw back unvested amount before cliff", func() {
		ctx := sdk.WrapSDKContext(s.ctx)
		newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceNewFunder := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)

		// Update clawback vesting account funder
		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
		_, err := s.app.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
		s.Require().NoError(err)

		// Perform clawback before cliff - funds should go to new funder (no dest address defined)
		msg := types.NewMsgClawback(newFunder, vestingAddr, sdk.AccAddress([]byte{}))
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		s.Require().NoError(err)

		// All initial vesting amount goes to funder
		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bNewF := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)

		// Original funder balance should not change
		s.Require().Equal(bF, balanceFunder)
		// New funder should get the vested tokens
		s.Require().Equal(balanceNewFunder.Add(vestingAmtTotal[0]).Amount.Uint64(), bNewF.Amount.Uint64())
		s.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
	})

	It("should update vesting funder and first funder cannot claw back unvested before cliff", func() {
		ctx := sdk.WrapSDKContext(s.ctx)
		newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceNewFunder := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)

		// Update clawback vesting account funder
		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
		_, err := s.app.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
		s.Require().NoError(err)

		// Original funder tries to perform clawback before cliff - is not the current funder
		msg := types.NewMsgClawback(funder, vestingAddr, sdk.AccAddress([]byte{}))
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		s.Require().Error(err)

		// All balances should remain the same
		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bNewF := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)

		s.Require().Equal(bF, balanceFunder)
		s.Require().Equal(balanceNewFunder, bNewF)
		s.Require().Equal(balanceGrantee, bG)
	})
})

// Trying to replicate the faulty behavior in MsgCreateClawbackVestingAccount,
// that was disclosed as a potential attack vector in relation to the Barberry
// security patch.
//
// It was possible to fund a clawback vesting account with negative amounts.
// Avoiding this requires an additional validation of the amount in the
// MsgFundVestingAccount's ValidateBasic method.
var _ = Describe("Clawback Vesting Account - Barberry bug", func() {
	var (
		// coinsNoNegAmount is a Coins struct with a positive and a negative amount of the same
		// denomination.
		coinsNoNegAmount = sdk.Coins{
			sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(1e18)},
		}
		// coinsWithNegAmount is a Coins struct with a positive and a negative amount of the same
		// denomination.
		coinsWithNegAmount = sdk.Coins{
			sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(1e18)},
			sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(-1e18)},
		}
		// coinsWithZeroAmount is a Coins struct with a positive and a zero amount of the same
		// denomination.
		coinsWithZeroAmount = sdk.Coins{
			sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(1e18)},
			sdk.Coin{Denom: utils.BaseDenom, Amount: sdk.NewInt(0)},
		}
		// emptyCoins is an Coins struct
		emptyCoins = sdk.Coins{}
		// funder and funderPriv are the address and private key of the account funding the vesting account
		funder, funderPriv = utiltx.NewAccAddressAndKey()
		// gasPrice is the gas price to be used in the transactions executed by the vesting account so that
		// the transaction fees can be deducted from the expected account balance
		gasPrice = math.NewInt(1e9)
		// vestingAddr and vestingPriv are the address and private key of the vesting account to be created
		vestingAddr, vestingPriv = utiltx.NewAccAddressAndKey()
		// vestingLength is a period of time in seconds to be used for the creation of the vesting
		// account.
		vestingLength = int64(60 * 60 * 24 * 30) // 30 days in seconds

		// txCost is the cost of a transaction to be deducted from the expected account balance
		txCost int64
	)

	BeforeEach(func() {
		s.SetupTest()

		// Initialize the account at the vesting address and the funder accounts by funding them
		fundedCoins := sdk.Coins{{Denom: utils.BaseDenom, Amount: sdk.NewInt(2e18)}} // fund more than what is sent to the vesting account for transaction fees
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr, fundedCoins)
		Expect(err).ToNot(HaveOccurred(), "failed to fund account")
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, fundedCoins)
		Expect(err).ToNot(HaveOccurred(), "failed to fund account")

		// Create a clawback vesting account
		msgCreate := types.NewMsgCreateClawbackVestingAccount(
			funder,
			vestingAddr,
		)

		res, err := testutil.DeliverTx(s.ctx, s.app, vestingPriv, &gasPrice, msgCreate)
		Expect(err).ToNot(HaveOccurred(), "failed to create clawback vesting account")
		txCost = gasPrice.Int64() * res.GasWanted

		// Check clawback acccount was created
		acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
		Expect(acc).ToNot(BeNil(), "clawback vesting account not created")
		_, ok := acc.(*types.ClawbackVestingAccount)
		Expect(ok).To(BeTrue(), "account is not a clawback vesting account")
	})

	Context("when creating a clawback vesting account", func() {
		testcases := []struct {
			name         string
			lockupCoins  sdk.Coins
			vestingCoins sdk.Coins
			expError     bool
			errContains  string
		}{
			{
				name:        "pass - positive amounts for the lockup period",
				lockupCoins: coinsNoNegAmount,
				expError:    false,
			},
			{
				name:         "pass - positive amounts for the vesting period",
				vestingCoins: coinsNoNegAmount,
				expError:     false,
			},
			{
				name:         "pass - positive amounts for both the lockup and vesting periods",
				lockupCoins:  coinsNoNegAmount,
				vestingCoins: coinsNoNegAmount,
				expError:     false,
			},
			{
				name:        "fail - negative amounts for the lockup period",
				lockupCoins: coinsWithNegAmount,
				expError:    true,
				errContains: "invalid amount in lockup periods, amounts must be positive",
			},
			{
				name:         "fail - negative amounts for the vesting period",
				vestingCoins: coinsWithNegAmount,
				expError:     true,
				errContains:  "invalid coins: invalid request",
			},
			{
				name:        "fail - zero amount for the lockup period",
				lockupCoins: coinsWithZeroAmount,
				expError:    true,
				errContains: "invalid amount in lockup periods, amounts must be positive",
			},
			{
				name:         "fail - zero amount for the vesting period",
				vestingCoins: coinsWithZeroAmount,
				expError:     true,
				errContains:  "invalid coins: invalid request",
			},
			{
				name:         "fail - empty amount for both the lockup and vesting periods",
				lockupCoins:  emptyCoins,
				vestingCoins: emptyCoins,
				expError:     true,
				errContains:  "vesting and/or lockup schedules must be present",
			},
		}

		for _, tc := range testcases {
			tc := tc
			It(tc.name, func() {
				var (
					lockupPeriods  sdkvesting.Periods
					vestingPeriods sdkvesting.Periods
				)

				if !tc.lockupCoins.Empty() {
					lockupPeriods = sdkvesting.Periods{
						sdkvesting.Period{Length: vestingLength, Amount: tc.lockupCoins},
					}
				}

				if !tc.vestingCoins.Empty() {
					vestingPeriods = sdkvesting.Periods{
						sdkvesting.Period{Length: vestingLength, Amount: tc.vestingCoins},
					}
				}

				// Fund the clawback vesting account at the given address
				msg := types.NewMsgFundVestingAccount(
					funder,
					vestingAddr,
					s.ctx.BlockTime(),
					lockupPeriods,
					vestingPeriods,
				)

				// Deliver transaction with message
				res, err := testutil.DeliverTx(s.ctx, s.app, funderPriv, nil, msg)

				// Get account at the new address
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
				vacc, _ := acc.(*types.ClawbackVestingAccount)

				if tc.expError {
					Expect(err).To(HaveOccurred(), "expected funding the vesting account to have failed")
					Expect(err.Error()).To(ContainSubstring(tc.errContains), "expected funding the vesting account to have failed")

					Expect(vacc.LockupPeriods).To(BeEmpty(), "expected clawback vesting account to not have been funded")
				} else {
					Expect(err).ToNot(HaveOccurred(), "failed to fund clawback vesting account")
					Expect(res.Code).To(Equal(uint32(0)), "failed to fund clawback vesting account")
					Expect(vacc.LockupPeriods).ToNot(BeEmpty(), "vesting account should have been funded")

					// Check that the vesting account has the correct balance
					balance := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, utils.BaseDenom)
					expBalance := int64(2e18) + int64(1e18) - txCost // fundedCoins + vestingCoins - txCost
					Expect(balance.Amount.Int64()).To(Equal(expBalance), "vesting account has incorrect balance")
				}
			})
		}
	})
})
