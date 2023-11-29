package keeper_test

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v15/testutil"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	"github.com/evmos/evmos/v15/utils"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	"github.com/evmos/evmos/v15/x/vesting/types"
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
	amt := math.NewInt(1e17)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
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
			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.ZeroInt()))
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
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)

			twoThirdsOfVested = vested.Sub(vested.QuoInt(math.NewInt(3))...)
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

			msgServer := stakingkeeper.NewMsgServerImpl(&s.app.StakingKeeper)
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
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup))))

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
			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, address, vestingAmtTotal.MulInt(math.NewInt(2)))
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
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup+1))))

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
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup*numLockupPeriods))))
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

			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.ZeroInt()))
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
	amt := math.NewInt(1)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingTotal := amt.Mul(math.NewInt(periodsTotal))
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, vestingTotal))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
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
	funder, funderPriv := utiltx.NewAccAddressAndKey()
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

		msg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, true)

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

		msg := types.NewMsgCreateClawbackVestingAccount(funder, emptyVestingAddr, false)

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
		res, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

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
		expVestedAmount := amt.Mul(math.NewInt(cliff))
		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
		unvested := vestingAmtTotal.Sub(vested...)

		s.Require().Equal(expVested, vested)
		s.Require().True(expVestedAmount.GT(math.NewInt(0)))
		s.Require().True(free.IsZero())
		s.Require().Equal(vesting, vestingAmtTotal)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		res, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(unvested), "expected unvested coins to be clawed back")

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
		s.Require().True(expVestedAmount.GT(math.NewInt(0)))
		s.Require().Equal(vesting, unvested)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		res, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(unvested), "expected only coins to be clawed back")

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
		res, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil(), "expected no error during clawback")
		Expect(res).ToNot(BeNil(), "expected response not to be nil")
		Expect(res.Coins).To(BeEmpty(), "expected nothing to be clawed back")

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// No amount is clawed back
		s.Require().Equal(balanceFunder, bF)
		s.Require().Equal(balanceGrantee, bG)
		s.Require().Equal(balanceDest, bD)
	})

	Context("while there is an active governance proposal for the vesting account", func() {
		var clawbackProposalID uint64

		BeforeEach(func() {
			// submit a different proposal to simulate having multiple proposals of different types
			// on chain.
			msgSubmitProposal, err := govv1beta1.NewMsgSubmitProposal(
				&erc20types.RegisterERC20Proposal{
					Title:          "test gov upgrade",
					Description:    "this is an example of a governance proposal to upgrade the evmos app",
					Erc20Addresses: []string{},
				},
				sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(1e9))),
				s.address.Bytes(),
			)
			Expect(err).ToNot(HaveOccurred(), "expected no error creating the proposal submission message")

			_, err = testutil.DeliverTx(s.ctx, s.app, s.priv, nil, msgSubmitProposal)
			Expect(err).ToNot(HaveOccurred(), "expected no error during proposal submission")

			// submit clawback proposal
			govClawbackProposal := &types.ClawbackProposal{
				Title:              "test gov clawback",
				Description:        "this is an example of a governance proposal to clawback vesting coins",
				Address:            vestingAddr.String(),
				DestinationAddress: funder.String(),
			}

			deposit := sdk.Coins{sdk.Coin{Denom: stakeDenom, Amount: math.NewInt(1)}}

			// Create the message to submit the proposal
			msgSubmit, err := govv1beta1.NewMsgSubmitProposal(
				govClawbackProposal, deposit, s.address.Bytes(),
			)
			Expect(err).ToNot(HaveOccurred(), "expected no error creating the proposal submission message")
			// deliver the proposal
			_, err = testutil.DeliverTx(s.ctx, s.app, s.priv, nil, msgSubmit)
			Expect(err).ToNot(HaveOccurred(), "expected no error during proposal submission")

			s.Commit()

			// Check if the proposal was submitted
			proposals := s.app.GovKeeper.GetProposals(s.ctx)
			Expect(len(proposals)).To(Equal(2), "expected two proposals to be found")
			proposal := proposals[len(proposals)-1]
			clawbackProposalID = proposal.Id
			Expect(proposal.GetTitle()).To(Equal("test gov clawback"), "expected different proposal title")
			Expect(proposal.Status).To(Equal(govv1.StatusDepositPeriod), "expected proposal to be in deposit period")
		})

		Context("with deposit made", func() {
			BeforeEach(func() {
				params := s.app.GovKeeper.GetParams(s.ctx)
				depositAmount := params.MinDeposit[0].Amount.Sub(math.NewInt(1))
				deposit := sdk.Coins{sdk.Coin{Denom: params.MinDeposit[0].Denom, Amount: depositAmount}}

				// Deliver the deposit
				msgDeposit := govv1beta1.NewMsgDeposit(s.address.Bytes(), clawbackProposalID, deposit)
				_, err := testutil.DeliverTx(s.ctx, s.app, s.priv, nil, msgDeposit)
				Expect(err).ToNot(HaveOccurred(), "expected no error during proposal deposit")

				s.Commit()

				// Check the proposal is in voting period
				proposal, found := s.app.GovKeeper.GetProposal(s.ctx, clawbackProposalID)
				Expect(found).To(BeTrue(), "expected proposal to be found")
				Expect(proposal.Status).To(Equal(govv1.StatusVotingPeriod), "expected proposal to be in voting period")

				// Check the store entry was set correctly
				hasActivePropposal := s.app.VestingKeeper.HasActiveClawbackProposal(s.ctx, vestingAddr)
				Expect(hasActivePropposal).To(BeTrue(), "expected an active clawback proposal for the vesting account")
			})

			It("should not allow clawback", func() {
				// Try to clawback tokens
				msgClawback := types.NewMsgClawback(funder, vestingAddr, dest)
				_, err = s.app.VestingKeeper.Clawback(sdk.WrapSDKContext(s.ctx), msgClawback)
				Expect(err).To(HaveOccurred(), "expected error during clawback while there is an active governance proposal")
				Expect(err.Error()).To(ContainSubstring("clawback is disabled while there is an active clawback proposal"))

				// Check that the clawback was not performed
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback := acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

				balances, err := s.app.VestingKeeper.Balances(s.ctx, &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				})
				Expect(err).ToNot(HaveOccurred(), "expected no error during balances query")
				Expect(balances.Unvested).To(Equal(vestingAmtTotal), "expected no tokens to be clawed back")

				// Delegate some funds to the suite validators in order to vote on proposal with enough voting power
				// using only the suite private key
				priv, ok := s.priv.(*ethsecp256k1.PrivKey)
				Expect(ok).To(BeTrue(), "expected private key to be of type ethsecp256k1.PrivKey")
				validators := s.app.StakingKeeper.GetBondedValidatorsByPower(s.ctx)
				err = testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, s.address.Bytes(), 5e18)
				Expect(err).ToNot(HaveOccurred(), "expected no error during funding of account")
				for _, val := range validators {
					res, err := testutil.Delegate(s.ctx, s.app, priv, sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)), val)
					Expect(err).ToNot(HaveOccurred(), "expected no error during delegation")
					Expect(res.Code).To(BeZero(), "expected delegation to succeed")
				}

				// Vote on proposal
				res, err := testutil.Vote(s.ctx, s.app, priv, clawbackProposalID, govv1beta1.OptionYes)
				Expect(err).ToNot(HaveOccurred(), "failed to vote on proposal %d", clawbackProposalID)
				Expect(res.Code).To(BeZero(), "expected proposal voting to succeed")

				// Check that the funds are clawed back after the proposal has ended
				s.CommitAfter(time.Hour * 24 * 365) // one year
				// Commit again because EndBlocker is run with time of the previous block and gov proposals are ended in EndBlocker
				s.Commit()

				// Check that proposal has passed
				proposal, found := s.app.GovKeeper.GetProposal(s.ctx, clawbackProposalID)
				Expect(found).To(BeTrue(), "expected proposal to exist")
				Expect(proposal.Status).ToNot(Equal(govv1.StatusVotingPeriod), "expected proposal to not be in voting period anymore")
				Expect(proposal.Status).To(Equal(govv1.StatusPassed), "expected proposal to have passed")

				// Check that the account was converted to a normal account
				acc = s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback = acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeFalse(), "expected account to be a normal account")

				hasActiveProposal := s.app.VestingKeeper.HasActiveClawbackProposal(s.ctx, vestingAddr)
				Expect(hasActiveProposal).To(BeFalse(), "expected no active clawback proposal")
			})

			It("should not allow changing the vesting funder", func() {
				msgUpdateFunder := types.NewMsgUpdateVestingFunder(funder, dest, vestingAddr)
				_, err = s.app.VestingKeeper.UpdateVestingFunder(sdk.WrapSDKContext(s.ctx), msgUpdateFunder)
				Expect(err).To(HaveOccurred(), "expected error during update funder while there is an active governance proposal")
				Expect(err.Error()).To(ContainSubstring("cannot update funder while there is an active clawback proposal"))

				// Check that the funder was not updated
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				clawbackAcc, isClawback := acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")
				Expect(clawbackAcc.FunderAddress).To(Equal(funder.String()), "expected funder to be unchanged")
			})
		})

		Context("without deposit made", func() {
			It("allows clawback and changing the funder before the deposit period ends", func() {
				newFunder, newPriv := utiltx.NewAccAddressAndKey()

				// fund accounts
				err = testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, newFunder, 5e18)
				Expect(err).ToNot(HaveOccurred(), "failed to fund target account")
				err = testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, funder, 5e18)
				Expect(err).ToNot(HaveOccurred(), "failed to fund target account")

				msgUpdateFunder := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
				_, err = testutil.DeliverTx(s.ctx, s.app, funderPriv, nil, msgUpdateFunder)
				Expect(err).ToNot(HaveOccurred(), "expected no error during update funder while there is an active governance proposal")

				// Check that the funder was updated
				acc := s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback := acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

				// Claw back tokens
				msgClawback := types.NewMsgClawback(newFunder, vestingAddr, funder)
				_, err = testutil.DeliverTx(s.ctx, s.app, newPriv, nil, msgClawback)
				Expect(err).ToNot(HaveOccurred(), "expected no error during clawback while there is no deposit made")

				// Check account is converted to a normal account
				acc = s.app.AccountKeeper.GetAccount(s.ctx, vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback = acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeFalse(), "expected account to be a normal account")
			})

			It("should remove the store entry after the deposit period ends", func() {
				s.CommitAfter(time.Hour * 24 * 365) // one year
				// Commit again because EndBlocker is run with time of the previous block and gov proposals are ended in EndBlocker
				s.Commit()

				// Check that the proposal has ended -- since deposit failed it's removed from the store
				_, found := s.app.GovKeeper.GetProposal(s.ctx, clawbackProposalID)
				Expect(found).To(BeFalse(), "expected proposal not to be found")

				// Check that the store entry was removed
				hasActiveProposal := s.app.VestingKeeper.HasActiveClawbackProposal(s.ctx, vestingAddr)
				Expect(hasActiveProposal).To(BeFalse(),
					"expected no active clawback proposal for address %q",
					vestingAddr.String(),
				)
			})
		})
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
		res, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

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

	Context("governance clawback to community pool", func() {
		It("should claw back unvested amount before cliff", func() {
			ctx := sdk.WrapSDKContext(s.ctx)

			// initial balances
			balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			pool := s.app.DistrKeeper.GetFeePool(s.ctx)
			balanceCommPool := pool.CommunityPool[0]

			// Perform clawback before cliff
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			res, err := s.app.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

			// All initial vesting amount goes to community pool instead of dest
			bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			pool = s.app.DistrKeeper.GetFeePool(s.ctx)
			bCP := pool.CommunityPool[0]

			s.Require().Equal(bF, balanceFunder)
			s.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
			// destination address should remain unchanged
			s.Require().Equal(balanceDest.Amount.Uint64(), bD.Amount.Uint64())
			// vesting amount should go to community pool
			s.Require().Equal(balanceCommPool.Amount.Add(math.LegacyNewDec(vestingAmtTotal[0].Amount.Int64())), bCP.Amount)
			s.Require().Equal(stakeDenom, bCP.Denom)
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
			expVestedAmount := amt.Mul(math.NewInt(cliff))
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
			unvested := vestingAmtTotal.Sub(vested...)

			s.Require().Equal(expVested, vested)
			s.Require().True(expVestedAmount.GT(math.NewInt(0)))
			s.Require().True(free.IsZero())
			s.Require().Equal(vesting, vestingAmtTotal)

			balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			pool := s.app.DistrKeeper.GetFeePool(s.ctx)
			balanceCommPool := pool.CommunityPool[0]

			testClawbackAccount := TestClawbackAccount{
				privKey:         nil,
				address:         vestingAddr,
				clawbackAccount: clawbackAccount,
			}
			// stake vested tokens
			_, err := delegate(testClawbackAccount, vested)
			Expect(err).To(BeNil())

			// Perform clawback
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			ctx := sdk.WrapSDKContext(s.ctx)
			res, err := s.app.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(unvested), "expected unvested coins to be clawed back")

			bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			pool = s.app.DistrKeeper.GetFeePool(s.ctx)
			bCP := pool.CommunityPool[0]

			expClawback := clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())

			// Any unvested amount is clawed back to community pool
			s.Require().Equal(balanceFunder, bF)
			s.Require().Equal(balanceGrantee.Sub(expClawback[0]).Amount.Uint64(), bG.Amount.Uint64())
			s.Require().Equal(balanceDest.Amount.Uint64(), bD.Amount.Uint64())
			// vesting amount should go to community pool
			s.Require().Equal(balanceCommPool.Amount.Add(math.LegacyNewDec(expClawback[0].Amount.Int64())), bCP.Amount)
			s.Require().Equal(stakeDenom, bCP.Denom)
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
			s.Require().True(expVestedAmount.GT(math.NewInt(0)))
			s.Require().Equal(vesting, unvested)

			balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			testClawbackAccount := TestClawbackAccount{
				privKey:         nil,
				address:         vestingAddr,
				clawbackAccount: clawbackAccount,
			}
			// stake vested tokens
			_, err := delegate(testClawbackAccount, vested)
			Expect(err).To(BeNil())

			// Perform clawback
			msg := types.NewMsgClawback(funder, vestingAddr, dest)
			ctx := sdk.WrapSDKContext(s.ctx)
			res, err := s.app.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(unvested), "expected only unvested coins to be clawed back")

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
			pool := s.app.DistrKeeper.GetFeePool(s.ctx)
			balanceCommPool := pool.CommunityPool[0]

			testClawbackAccount := TestClawbackAccount{
				privKey:         nil,
				address:         vestingAddr,
				clawbackAccount: clawbackAccount,
			}
			// stake vested tokens
			_, err := delegate(testClawbackAccount, vested)
			Expect(err).To(BeNil())

			// Perform clawback
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			ctx := sdk.WrapSDKContext(s.ctx)
			res, err := s.app.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil(), "expected no error during clawback")
			Expect(res.Coins).To(BeEmpty(), "expected nothing to be clawed back after end of vesting schedules")

			bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			pool = s.app.DistrKeeper.GetFeePool(s.ctx)
			bCP := pool.CommunityPool[0]

			// No amount is clawed back
			s.Require().Equal(balanceFunder, bF)
			s.Require().Equal(balanceGrantee, bG)
			s.Require().Equal(balanceDest, bD)
			s.Require().Equal(balanceCommPool.Amount, bCP.Amount)
		})

		It("should update vesting funder and claw back unvested amount before cliff", func() {
			ctx := sdk.WrapSDKContext(s.ctx)
			newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

			balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			balanceNewFunder := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
			balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			pool := s.app.DistrKeeper.GetFeePool(s.ctx)
			balanceCommPool := pool.CommunityPool[0]

			// Update clawback vesting account funder
			updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
			_, err := s.app.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
			s.Require().NoError(err)

			// Perform clawback before cliff - funds should go to new funder (no dest address defined)
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, nil)
			res, err := s.app.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

			// All initial vesting amount goes to funder
			bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			bNewF := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
			bG := s.app.BankKeeper.GetBalance(s.ctx, vestingAddr, stakeDenom)
			pool = s.app.DistrKeeper.GetFeePool(s.ctx)
			bCP := pool.CommunityPool[0]

			// Original funder balance should not change
			s.Require().Equal(bF, balanceFunder)
			// New funder should not get the vested tokens
			s.Require().Equal(balanceNewFunder.Amount.Uint64(), bNewF.Amount.Uint64())
			s.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
			// vesting amount should go to community pool
			s.Require().Equal(balanceCommPool.Amount.Add(math.LegacyNewDec(vestingAmtTotal[0].Amount.Int64())), bCP.Amount)
		})

		It("should not claw back when governance clawback is disabled", func() {
			// disable governance clawback
			s.app.VestingKeeper.SetGovClawbackDisabled(s.ctx, vestingAddr)

			// Perform clawback before cliff
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			_, err := s.app.VestingKeeper.Clawback(s.ctx, msg)
			Expect(err).To(HaveOccurred(), "expected error")
			Expect(err.Error()).To(ContainSubstring("%s: account does not have governance clawback enabled", vestingAddr.String()))
		})
	})
})

// Testing that smart contracts cannot be converted to clawback vesting accounts
//
// NOTE: For smart contracts, it is not possible to directly call keeper methods
// or send SDK transactions. They go exclusively through the EVM, which is tested
// in the precompiles package.
// The test here is just confirming the expected behavior on the module level.
var _ = Describe("Clawback Vesting Account - Smart contract", func() {
	var (
		contractAddr common.Address
		contract     evmtypes.CompiledContract
		err          error
	)

	BeforeEach(func() {
		contract = contracts.ERC20MinterBurnerDecimalsContract
		contractAddr, err = testutil.DeployContract(
			s.ctx,
			s.app,
			s.priv,
			s.queryClientEvm,
			contract,
			"Test", "TTT", uint8(18),
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")
	})

	It("should not convert a smart contract to a clawback vesting account", func() {
		msgCreate := types.NewMsgCreateClawbackVestingAccount(
			s.address.Bytes(),
			contractAddr.Bytes(),
			false,
		)
		_, err := s.app.VestingKeeper.CreateClawbackVestingAccount(s.ctx, msgCreate)
		Expect(err).To(HaveOccurred(), "expected error")
		Expect(err.Error()).To(ContainSubstring(
			fmt.Sprintf(
				"account %s is a contract account and cannot be converted in a clawback vesting account",
				sdk.AccAddress(contractAddr.Bytes()).String()),
		))

		// Check that the account was not converted
		acc := s.app.AccountKeeper.GetAccount(s.ctx, contractAddr.Bytes())
		Expect(acc).ToNot(BeNil(), "smart contract should be found")
		_, ok := acc.(*types.ClawbackVestingAccount)
		Expect(ok).To(BeFalse(), "account should not be a clawback vesting account")

		// Check that the contract code was not deleted
		//
		// NOTE: When it was possible to create clawback vesting accounts for smart contracts,
		// the contract code was deleted from the EVM state. This checks that this is not the case.
		res, err := s.app.EvmKeeper.Code(s.ctx, &evmtypes.QueryCodeRequest{Address: contractAddr.String()})
		Expect(err).ToNot(HaveOccurred(), "failed to query contract code")
		Expect(res.Code).ToNot(BeEmpty(), "contract code should not be empty")
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
			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)},
		}
		// coinsWithNegAmount is a Coins struct with a positive and a negative amount of the same
		// denomination.
		coinsWithNegAmount = sdk.Coins{
			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)},
			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(-1e18)},
		}
		// coinsWithZeroAmount is a Coins struct with a positive and a zero amount of the same
		// denomination.
		coinsWithZeroAmount = sdk.Coins{
			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)},
			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(0)},
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
		fundedCoins := sdk.Coins{{Denom: utils.BaseDenom, Amount: math.NewInt(2e18)}} // fund more than what is sent to the vesting account for transaction fees
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, vestingAddr, fundedCoins)
		Expect(err).ToNot(HaveOccurred(), "failed to fund account")
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, fundedCoins)
		Expect(err).ToNot(HaveOccurred(), "failed to fund account")

		// Create a clawback vesting account
		msgCreate := types.NewMsgCreateClawbackVestingAccount(
			funder,
			vestingAddr,
			false,
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

	Context("when funding a clawback vesting account", func() {
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
				errContains: errortypes.ErrInvalidCoins.Wrap(coinsWithNegAmount.String()).Error(),
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
				errContains: errortypes.ErrInvalidCoins.Wrap(coinsWithZeroAmount.String()).Error(),
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
