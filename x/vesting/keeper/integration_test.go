package keeper_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/testutil"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"github.com/evmos/evmos/v16/x/vesting/types"
)

func TestKeeperTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

// TestClawbackAccount is a struct to store all relevant information that is corresponding
// to a clawback vesting account.
type TestClawbackAccount struct {
	privKey         cryptotypes.PrivKey
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
	var s *KeeperTestSuite
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

	accountGasCoverageNativeCoin := sdk.NewCoin(stakeDenom, math.NewInt(1e16))
	accountGasCoverage := sdk.NewCoins(accountGasCoverageNativeCoin)

	var (
		clawbackAccount   *types.ClawbackVestingAccount
		unvested          sdk.Coins
		vested            sdk.Coins
		twoThirdsOfVested sdk.Coins
	)

	dest := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	funder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	BeforeEach(func() {
		s = new(KeeperTestSuite)
		s.SetupTest()

		// Initialize all test accounts
		for i, account := range testAccounts {
			// Create and fund periodic vesting account
			vestingStart := s.network.GetContext().BlockTime()
			baseAccount := authtypes.NewBaseAccountWithAddress(account.address)
			clawbackAccount = types.NewClawbackVestingAccount(
				baseAccount,
				funder,
				vestingAmtTotal,
				vestingStart,
				lockupPeriods,
				vestingPeriods,
			)

			err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, account.address, vestingAmtTotal)
			s.Require().NoError(err)
			acc := s.network.App.AccountKeeper.NewAccount(s.network.GetContext(), clawbackAccount)
			s.network.App.AccountKeeper.SetAccount(s.network.GetContext(), acc)

			// Check if all tokens are unvested at vestingStart
			unvested = clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			Expect(vestingAmtTotal).To(Equal(unvested))
			s.Require().True(vested.IsZero())

			// Grant gas stipend to cover EVM fees
			err = testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, clawbackAccount.GetAddress(), accountGasCoverage)
			s.Require().NoError(err)
			granteeBalance, err := s.handler.GetBalance(account.address, stakeDenom)
			Expect(granteeBalance).To(Equal(accountGasCoverage[0].Add(vestingAmtTotal[0])))

			// Update testAccounts clawbackAccount reference
			testAccounts[i].clawbackAccount = clawbackAccount
		}
	})

	Context("before first vesting period", func() {
		BeforeEach(func() {
			// Ensure no tokens are vested
			vested := clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			unlocked := clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.ZeroInt()))
			Expect(zeroCoins).To(Equal(vested))
			Expect(zeroCoins).To(Equal(unlocked))
		})

		It("cannot delegate tokens", func() {
			err := s.factory.Delegate(
				testAccounts[0].privKey,
				s.network.GetValidators()[0].OperatorAddress,
				accountGasCoverageNativeCoin.Add(sdk.NewCoin(stakeDenom, math.NewInt(1))),
			)
			Expect(err).ToNot(BeNil())
		})

		It("can transfer spendable tokens", func() {
			account := testAccounts[0]
			// Fund account with new spendable tokens
			err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, account.address, unvested)
			Expect(err).To(BeNil())

			err = s.network.App.BankKeeper.SendCoins(
				s.network.GetContext(),
				account.address,
				dest,
				unvested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.network.App.BankKeeper.SendCoins(
				s.network.GetContext(),
				clawbackAccount.GetAddress(),
				dest,
				unvested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := testAccounts[0]
			// Fund account with new spendable tokens
			err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, account.address, unlockedPerLockup)
			Expect(err).To(BeNil())

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("cannot perform Ethereum tx with unvested balance", func() {
			account := testAccounts[0]
			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})
	})

	Context("after first vesting period and before lockup", func() {
		BeforeEach(func() {
			// Surpass cliff but none of lockup duration
			cliffDuration := time.Duration(cliffLength)
			s.network.NextBlockAfter(cliffDuration * time.Second)

			// Check if some, but not all tokens are vested
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			Expect(expVested).To(Equal(vested))

			twoThirdsOfVested = vested.Sub(vested.QuoInt(math.NewInt(3))...)
		})

		It("can delegate vested tokens and update spendable balance", func() {
			testAccount := testAccounts[0]
			// Verify that the total spendable coins decreases after staking
			// vested tokens.
			spendablePre := s.network.App.BankKeeper.SpendableCoins(s.network.GetContext(), testAccount.address)

			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err := s.factory.Delegate(
				testAccount.privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			spendablePost := s.network.App.BankKeeper.SpendableCoins(s.network.GetContext(), testAccount.address)
			Expect(spendablePost.AmountOf(stakeDenom).GT(spendablePre.AmountOf(stakeDenom)))
		})

		It("cannot delegate unvested tokens", func() {
			ok, vestedCoin := vestingAmtTotal.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err := s.factory.Delegate(
				testAccounts[0].privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot delegate unvested tokens in batches", func() {
			ok, vestedCoin := twoThirdsOfVested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err := s.factory.Delegate(
				testAccounts[0].privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).NotTo(BeNil())
		})

		It("cannot delegate then send tokens", func() {
			ok, vestedCoin := twoThirdsOfVested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err := s.factory.Delegate(
				testAccounts[0].privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).To(BeNil())

			err = s.network.App.BankKeeper.SendCoins(
				s.network.GetContext(),
				clawbackAccount.GetAddress(),
				dest,
				twoThirdsOfVested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot transfer vested tokens", func() {
			err := s.network.App.BankKeeper.SendCoins(
				s.network.GetContext(),
				clawbackAccount.GetAddress(),
				dest,
				vested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := testAccounts[0]
			// Fund account with new spendable tokens
			err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, account.address, unlockedPerLockup)
			Expect(err).To(BeNil())

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("cannot perform Ethereum tx with locked balance", func() {
			account := testAccounts[0]
			txAmount := vested.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})
	})

	Context("Between first and second lockup periods", func() {
		BeforeEach(func() {
			// Surpass first lockup
			vestDuration := time.Duration(lockupLength)
			s.network.NextBlockAfter(vestDuration * time.Second)

			// Check if some, but not all tokens are vested and unlocked
			for _, account := range testAccounts {
				vested := account.clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
				unlocked := account.clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup))))

				s.Require().NotEqual(vestingAmtTotal, vested)
				Expect(expVested).To(Equal(vested))
				Expect(unlocked).To(Equal(unlockedPerLockup))
			}
		})

		It("should enable access to unlocked EVM tokens (single-account, single-msg)", func() {
			account := testAccounts[0]

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("should enable access to unlocked EVM tokens (single-account, multiple-msgs)", func() {
			account := testAccounts[0]

			// Split the total unlocked amount into numTestMsgs equally sized tx's
			msgs := make([]sdk.Msg, numTestMsgs)
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()

			for i := 0; i < numTestMsgs; i++ {
				msgs[i], err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, i)
				Expect(err).To(BeNil())
			}

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, unlockedPerLockupAmt, stakeDenom, msgs...)
		})

		It("should enable access to unlocked EVM tokens (multi-account, single-msg)", func() {
			txAmount := unlockedPerLockupAmt.BigInt()

			msgs := make([]sdk.Msg, numTestAccounts)
			for i, grantee := range testAccounts {
				msgs[i], err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, grantee.privKey, grantee.address, dest, txAmount, 0)
				Expect(err).To(BeNil())
			}

			assertEthSucceeds(testAccounts, funder, dest, unlockedPerLockupAmt, stakeDenom, msgs...)
		})

		It("should enable access to unlocked EVM tokens (multi-account, multiple-msgs)", func() {
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()

			for _, grantee := range testAccounts {
				for j := 0; j < numTestMsgs; j++ {
					addedMsg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, grantee.privKey, grantee.address, dest, txAmount, j)
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
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, testAccount.privKey, testAccount.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})

		It("should not enable access to locked EVM tokens (single-account, multiple-msgs)", func() {
			msgs := make([]sdk.Msg, numTestMsgs+1)
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()
			testAccount := testAccounts[0]

			// Add additional message that exceeds unlocked balance
			for i := 0; i < numTestMsgs+1; i++ {
				msgs[i], err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, testAccount.privKey, testAccount.address, dest, txAmount, i)
				Expect(err).To(BeNil())
			}

			assertEthFails(msgs...)
		})

		It("should not enable access to locked EVM tokens (multi-account, single-msg)", func() {
			msgs := make([]sdk.Msg, numTestAccounts+1)
			txAmount := unlockedPerLockupAmt.BigInt()

			for i, account := range testAccounts {
				msgs[i], err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
				Expect(err).To(BeNil())
			}

			// Add additional message that exceeds unlocked balance
			msgs[numTestAccounts], err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, testAccounts[0].privKey, testAccounts[0].address, dest, txAmount, 1)
			Expect(err).To(BeNil())

			assertEthFails(msgs...)
		})

		It("should not enable access to locked EVM tokens (multi-account, multiple-msgs)", func() {
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockupAmt.QuoRaw(int64(numTestMsgs)).BigInt()
			var addedMsg sdk.Msg

			for _, account := range testAccounts {
				for j := 0; j < numTestMsgs; j++ {
					addedMsg, err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, j)
					msgs = append(msgs, addedMsg)
				}
			}

			// Add additional message that exceeds unlocked balance
			addedMsg, err = utiltx.CreateEthTx(s.network.GetContext(), s.network.App, testAccounts[0].privKey, testAccounts[0].address, dest, txAmount, numTestMsgs)
			Expect(err).To(BeNil())
			msgs = append(msgs, addedMsg)

			assertEthFails(msgs...)
		})

		It("should not short-circuit with a normal account", func() {
			account := testAccounts[0]
			address, privKey := utiltx.NewAccAddressAndKey()

			txAmount := vestingAmtTotal.AmountOf(stakeDenom).BigInt()

			// Fund a normal account to try to short-circuit the AnteHandler
			err = testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, address, vestingAmtTotal.MulInt(math.NewInt(2)))
			Expect(err).To(BeNil())
			normalAccMsg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, privKey, address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			// Attempt to spend entire balance
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())
			err = validateEthVestingTransactionDecorator(normalAccMsg, msg)
			Expect(err).ToNot(BeNil())

			_, err = testutil.DeliverEthTx(s.network.App, nil, msg)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("after first lockup and additional vest", func() {
		BeforeEach(func() {
			vestDuration := time.Duration(lockupLength + vestingLength)
			s.network.NextBlockAfter(vestDuration * time.Second)

			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup+1))))

			unlocked := clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			expUnlocked := unlockedPerLockup

			Expect(expVested).To(Equal(vested))
			Expect(expUnlocked).To(Equal(unlocked))
		})

		It("should enable access to unlocked EVM tokens", func() {
			testAccount := testAccounts[0]

			txAmount := unlockedPerLockupAmt.BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, testAccount.privKey, testAccount.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{testAccount}, funder, dest, unlockedPerLockupAmt, stakeDenom, msg)
		})

		It("should not enable access to locked EVM tokens", func() {
			testAccount := testAccounts[0]

			txAmount := vested.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, testAccount.privKey, testAccount.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})
	})

	Context("after half of vesting period and both lockups", func() {
		BeforeEach(func() {
			// Surpass lockup duration
			lockupDuration := time.Duration(lockupLength * numLockupPeriods)
			s.network.NextBlockAfter(lockupDuration * time.Second)

			// Check if some, but not all tokens are vested
			unvested = clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup*numLockupPeriods))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			Expect(expVested).To(Equal(vested))
		})

		It("can delegate vested tokens", func() {
			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err := s.factory.Delegate(
				testAccounts[0].privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())
			// TODO check balance change
		})

		It("cannot delegate unvested tokens", func() {
			ok, vestedCoin := vestingAmtTotal.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err := s.factory.Delegate(
				testAccounts[0].privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can transfer vested tokens", func() {
			err := s.network.App.BankKeeper.SendCoins(
				s.network.GetContext(),
				clawbackAccount.GetAddress(),
				sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.network.App.BankKeeper.SendCoins(
				s.network.GetContext(),
				clawbackAccount.GetAddress(),
				sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
				vestingAmtTotal,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := testAccounts[0]

			txAmount := vested.AmountOf(stakeDenom).BigInt()
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount, 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, vested.AmountOf(stakeDenom), stakeDenom, msg)
		})
	})

	Context("after entire vesting period and both lockups", func() {
		BeforeEach(func() {
			// Surpass vest duration
			vestDuration := time.Duration(vestingLength * periodsTotal)
			s.network.NextBlockAfter(vestDuration * time.Second)

			// Check that all tokens are vested and unlocked
			unvested = clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			locked := clawbackAccount.LockedCoins(s.network.GetContext().BlockTime())

			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.ZeroInt()))
			Expect(vestingAmtTotal).To(Equal(vested))
			Expect(zeroCoins).To(Equal(locked))
			Expect(zeroCoins).To(Equal(unvested))
		})

		It("can send entire balance", func() {
			account := testAccounts[0]

			txAmount := vestingAmtTotal.AmountOf(stakeDenom)
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount.BigInt(), 0)
			Expect(err).To(BeNil())

			assertEthSucceeds([]TestClawbackAccount{account}, funder, dest, txAmount, stakeDenom, msg)
		})

		It("cannot exceed balance", func() {
			account := testAccounts[0]

			txAmount := vestingAmtTotal.AmountOf(stakeDenom).Mul(math.NewInt(2))
			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, txAmount.BigInt(), 0)
			Expect(err).To(BeNil())

			assertEthFails(msg)
		})

		It("should short-circuit with zero balance", func() {
			account := testAccounts[0]
			balRes, err := s.handler.GetBalance(account.address, stakeDenom)
			balance := balRes.Balance

			// Drain account balance
			err := s.network.App.BankKeeper.SendCoins(s.network.GetContext(), account.address, dest, sdk.NewCoins(balance))
			Expect(err).To(BeNil())

			msg, err := utiltx.CreateEthTx(s.network.GetContext(), s.network.App, account.privKey, account.address, dest, big.NewInt(0), 0)
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
	var s *KeeperTestSuite
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
		s = new(KeeperTestSuite)
		s.SetupTest()
		vestingStart := s.network.GetContext().BlockTime()

		// Initialize account at vesting address by funding it with tokens
		// and then send them over to the vesting funder
		err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, vestingAddr, vestingAmtTotal)
		Expect(err).ToNot(HaveOccurred(), "failed to fund target account")
		err = s.network.App.BankKeeper.SendCoins(s.network.GetContext(), vestingAddr, funder, vestingAmtTotal)
		Expect(err).ToNot(HaveOccurred(), "failed to send coins to funder")

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		balanceDest := balRes.Balance
		Expect(balanceFunder).To(Equal(vestingAmtTotal[0]), "expected different funder balance")
		Expect(balanceGrantee.IsZero()).To(BeTrue(), "expected balance of vesting account to be zero")
		Expect(balanceDest.IsZero()).To(BeTrue(), "expected destination balance to be zero")

		msg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, true)

		_, err = s.network.App.VestingKeeper.CreateClawbackVestingAccount(s.network.GetContext(), msg)
		Expect(err).ToNot(HaveOccurred(), "expected creating clawback vesting account to succeed")

		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
		clawbackAccount, isClawback = acc.(*types.ClawbackVestingAccount)
		Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

		// fund the vesting account
		msgFund := types.NewMsgFundVestingAccount(funder, vestingAddr, vestingStart, lockupPeriods, vestingPeriods)
		_, err = s.network.App.VestingKeeper.FundVestingAccount(s.network.GetContext(), msgFund)
		Expect(err).ToNot(HaveOccurred(), "expected funding vesting account to succeed")

		acc = s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
		Expect(acc).ToNot(BeNil(), "expected account to exist")
		clawbackAccount, isClawback = acc.(*types.ClawbackVestingAccount)
		Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

		// Check if all tokens are unvested and locked at vestingStart
		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
		Expect(vesting).To(Equal(vestingAmtTotal), "expected difference vesting tokens")
		Expect(vested.IsZero()).To(BeTrue(), "expected no tokens to be vested")
		Expect(unlocked.IsZero()).To(BeTrue(), "expected no tokens to be unlocked")

		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee = balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		balanceDest = balRes.Balance

		Expect(bF.IsZero()).To(BeTrue(), "expected funder balance to be zero")
		Expect(balanceGrantee).To(Equal(vestingAmtTotal[0]), "expected all tokens to be locked")
		Expect(balanceDest.IsZero()).To(BeTrue(), "expected no tokens to be unlocked")
	})

	It("should fail if there is no vesting or lockup schedule set", func() {
		ctx := s.network.GetContext()
		emptyVestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
		err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, emptyVestingAddr, vestingAmtTotal)
		Expect(err).ToNot(HaveOccurred(), "failed to fund target account")

		msg := types.NewMsgCreateClawbackVestingAccount(funder, emptyVestingAddr, false)

		_, err = s.network.App.VestingKeeper.CreateClawbackVestingAccount(s.network.GetContext(), msg)
		Expect(err).ToNot(HaveOccurred(), "expected creating clawback vesting account to succeed")

		clawbackMsg := types.NewMsgClawback(funder, emptyVestingAddr, dest)
		_, err = s.network.App.VestingKeeper.Clawback(ctx, clawbackMsg)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("has no vesting or lockup periods"))
	})

	It("should claw back unvested amount before cliff", func() {
		ctx := s.network.GetContext()

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		balanceDest := balRes.Balance

		// Perform clawback before cliff
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

		// All initial vesting amount goes to dest
		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		bG := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		bD := balRes.Balance

		Expect(bF).To(Equal(balanceFunder), "expected funder balance to be unchanged")
		Expect(bG.IsZero()).To(BeTrue(), "expected all tokens to be clawed back")
		Expect(bD).To(Equal(balanceDest.Add(vestingAmtTotal[0])), "expected all tokens to be clawed back to the destination account")
	})

	It("should claw back any unvested amount after cliff before unlocking", func() {
		// Surpass cliff but not lockup duration
		cliffDuration := time.Duration(cliffLength)
		s.network.NextBlockAfter(cliffDuration * time.Second)

		// Check that all tokens are locked and some, but not all tokens are vested
		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
		free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
		expVestedAmount := amt.Mul(math.NewInt(cliff))
		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
		unvested := vestingAmtTotal.Sub(vested...)

		Expect(expVested).To(Equal(vested))
		s.Require().True(expVestedAmount.GT(math.NewInt(0)))
		s.Require().True(free.IsZero())
		Expect(vesting).To(Equal(vestingAmtTotal))

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		balanceDest := balRes.Balance

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := s.network.GetContext()
		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(unvested), "expected unvested coins to be clawed back")

		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		bG := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		bD := balRes.Balance

		expClawback := clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())

		// Any unvested amount is clawed back
		Expect(balanceFunder).To(Equal(bF))
		Expect(balanceGrantee.Sub(expClawback[0]).Amount).To(Equal(bG.Amount))
		Expect(balanceDest.Add(expClawback[0]).Amount).To(Equal(bD.Amount))
	})

	It("should claw back any unvested amount after cliff and unlocking", func() {
		// Surpass lockup duration
		// A strict `if t < clawbackTime` comparison is used in ComputeClawback
		// so, we increment the duration with 1 for the free token calculation to match
		lockupDuration := time.Duration(lockupLength + 1)
		s.network.NextBlockAfter(lockupDuration * time.Second)

		// Check if some, but not all tokens are vested and unlocked
		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
		free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
		expVestedAmount := amt.Mul(math.NewInt(lockup))
		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
		unvested := vestingAmtTotal.Sub(vested...)

		Expect(free).To(Equal(vested))
		Expect(expVested).To(Equal(vested))
		s.Require().True(expVestedAmount.GT(math.NewInt(0)))
		Expect(vesting).To(Equal(unvested))

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		balanceDest := balRes.Balance

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := s.network.GetContext()
		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(unvested), "expected only coins to be clawed back")

		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		bG := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		bD := balRes.Balance

		// Any unvested amount is clawed back
		Expect(balanceFunder).To(Equal(bF))
		Expect(balanceGrantee.Sub(vesting[0]).Amount).To(Equal(bG.Amount))
		Expect(balanceDest.Add(vesting[0]).Amount).To(Equal(bD.Amount))
	})

	It("should not claw back any amount after vesting periods end", func() {
		// Surpass vesting periods
		vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
		s.network.NextBlockAfter(vestingDuration * time.Second)

		// Check if some, but not all tokens are vested and unlocked
		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
		free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())

		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))
		unvested := vestingAmtTotal.Sub(vested...)

		Expect(free).To(Equal(vested))
		Expect(expVested).To(Equal(vested))
		Expect(expVested).To(Equal(vestingAmtTotal))
		Expect(unlocked).To(Equal(vestingAmtTotal))
		Expect(vesting).To(Equal(unvested))
		s.Require().True(vesting.IsZero())

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		balanceDest := balRes.Balance

		// Perform clawback
		msg := types.NewMsgClawback(funder, vestingAddr, dest)
		ctx := s.network.GetContext()
		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil(), "expected no error during clawback")
		Expect(res).ToNot(BeNil(), "expected response not to be nil")
		Expect(res.Coins).To(BeEmpty(), "expected nothing to be clawed back")

		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		bG := balRes.Balance
		balRes, err = s.handler.GetBalance(dest, stakeDenom)
		Expect(err).To(BeNil())
		bD := balRes.Balance

		// No amount is clawed back
		Expect(balanceFunder).To(Equal(bF))
		Expect(balanceGrantee).To(Equal(bG))
		Expect(balanceDest).To(Equal(bD))
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
				s.keyring.GetAddr(0).Bytes(),
			)
			Expect(err).ToNot(HaveOccurred(), "expected no error creating the proposal submission message")

			_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, s.keyring.GetPrivKey(0), nil, msgSubmitProposal)
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
				govClawbackProposal, deposit, s.keyring.GetAddr(0).Bytes(),
			)
			Expect(err).ToNot(HaveOccurred(), "expected no error creating the proposal submission message")
			// deliver the proposal
			_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, s.keyring.GetPrivKey(0), nil, msgSubmit)
			Expect(err).ToNot(HaveOccurred(), "expected no error during proposal submission")

			Expect(s.network.NextBlock()).To(BeNil())

			// Check if the proposal was submitted
			res, err := s.network.GetGovClient().Proposals(s.network.GetContext(), &govv1.QueryProposalsRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).ToNot(BeNil())

			Expect(len(res.Proposals)).To(Equal(2), "expected two proposals to be found")
			proposal := res.Proposals[len(res.Proposals)-1]
			clawbackProposalID = proposal.Id
			Expect(proposal.GetTitle()).To(Equal("test gov clawback"), "expected different proposal title")
			Expect(proposal.Status).To(Equal(govv1.StatusDepositPeriod), "expected proposal to be in deposit period")
		})

		Context("with deposit made", func() {
			BeforeEach(func() {
				params, err := s.network.App.GovKeeper.Params.Get(s.network.GetContext())
				Expect(err).ToNot(HaveOccurred())
				depositAmount := params.MinDeposit[0].Amount.Sub(math.NewInt(1))
				deposit := sdk.Coins{sdk.Coin{Denom: params.MinDeposit[0].Denom, Amount: depositAmount}}

				// Deliver the deposit
				msgDeposit := govv1beta1.NewMsgDeposit(s.keyring.GetAddr(0).Bytes(), clawbackProposalID, deposit)
				_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, s.keyring.GetPrivKey(0), nil, msgDeposit)
				Expect(err).ToNot(HaveOccurred(), "expected no error during proposal deposit")

				Expect(s.network.NextBlock()).To(BeNil())

				// Check the proposal is in voting period
				proposal, found := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), clawbackProposalID)
				Expect(found).To(BeTrue(), "expected proposal to be found")
				Expect(proposal.Status).To(Equal(govv1.StatusVotingPeriod), "expected proposal to be in voting period")

				// Check the store entry was set correctly
				hasActivePropposal := s.network.App.VestingKeeper.HasActiveClawbackProposal(s.network.GetContext(), vestingAddr)
				Expect(hasActivePropposal).To(BeTrue(), "expected an active clawback proposal for the vesting account")
			})

			It("should not allow clawback", func() {
				// Try to clawback tokens
				msgClawback := types.NewMsgClawback(funder, vestingAddr, dest)
				_, err = s.network.App.VestingKeeper.Clawback(s.network.GetContext(), msgClawback)
				Expect(err).To(HaveOccurred(), "expected error during clawback while there is an active governance proposal")
				Expect(err.Error()).To(ContainSubstring("clawback is disabled while there is an active clawback proposal"))

				// Check that the clawback was not performed
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback := acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

				balances, err := s.network.App.VestingKeeper.Balances(s.network.GetContext(), &types.QueryBalancesRequest{
					Address: vestingAddr.String(),
				})
				Expect(err).ToNot(HaveOccurred(), "expected no error during balances query")
				Expect(balances.Unvested).To(Equal(vestingAmtTotal), "expected no tokens to be clawed back")

				// Delegate some funds to the suite validators in order to vote on proposal with enough voting power
				// using only the suite private key
				priv, ok := s.keyring.GetPrivKey(0).(*ethsecp256k1.PrivKey)
				Expect(ok).To(BeTrue(), "expected private key to be of type ethsecp256k1.PrivKey")
				validators, err := s.network.App.StakingKeeper.GetBondedValidatorsByPower(s.network.GetContext())
				Expect(err).ToNot(HaveOccurred())
				err = testutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, s.keyring.GetAddr(0).Bytes(), 5e18)
				Expect(err).ToNot(HaveOccurred(), "expected no error during funding of account")
				for _, val := range validators {
					res, err := testutil.Delegate(s.network.GetContext(), s.network.App, priv, sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)), val)
					Expect(err).ToNot(HaveOccurred(), "expected no error during delegation")
					Expect(res.Code).To(BeZero(), "expected delegation to succeed")
				}

				// Vote on proposal
				res, err := testutil.Vote(s.network.GetContext(), s.network.App, priv, clawbackProposalID, govv1beta1.OptionYes)
				Expect(err).ToNot(HaveOccurred(), "failed to vote on proposal %d", clawbackProposalID)
				Expect(res.Code).To(BeZero(), "expected proposal voting to succeed")

				// Check that the funds are clawed back after the proposal has ended
				s.network.NextBlockAfter(time.Hour * 24 * 365) // one year
				// Commit again because EndBlocker is run with time of the previous block and gov proposals are ended in EndBlocker
				Expect(s.network.NextBlock()).To(BeNil())

				// Check that proposal has passed
				proposal, found := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), clawbackProposalID)
				Expect(found).To(BeTrue(), "expected proposal to exist")
				Expect(proposal.Status).ToNot(Equal(govv1.StatusVotingPeriod), "expected proposal to not be in voting period anymore")
				Expect(proposal.Status).To(Equal(govv1.StatusPassed), "expected proposal to have passed")

				// Check that the account was converted to a normal account
				acc = s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback = acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeFalse(), "expected account to be a normal account")

				hasActiveProposal := s.network.App.VestingKeeper.HasActiveClawbackProposal(s.network.GetContext(), vestingAddr)
				Expect(hasActiveProposal).To(BeFalse(), "expected no active clawback proposal")
			})

			It("should not allow changing the vesting funder", func() {
				msgUpdateFunder := types.NewMsgUpdateVestingFunder(funder, dest, vestingAddr)
				_, err = s.network.App.VestingKeeper.UpdateVestingFunder(s.network.GetContext(), msgUpdateFunder)
				Expect(err).To(HaveOccurred(), "expected error during update funder while there is an active governance proposal")
				Expect(err.Error()).To(ContainSubstring("cannot update funder while there is an active clawback proposal"))

				// Check that the funder was not updated
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
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
				err = testutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, newFunder, 5e18)
				Expect(err).ToNot(HaveOccurred(), "failed to fund target account")
				err = testutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, funder, 5e18)
				Expect(err).ToNot(HaveOccurred(), "failed to fund target account")

				msgUpdateFunder := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
				_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, funderPriv, nil, msgUpdateFunder)
				Expect(err).ToNot(HaveOccurred(), "expected no error during update funder while there is an active governance proposal")

				// Check that the funder was updated
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback := acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

				// Claw back tokens
				msgClawback := types.NewMsgClawback(newFunder, vestingAddr, funder)
				_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, newPriv, nil, msgClawback)
				Expect(err).ToNot(HaveOccurred(), "expected no error during clawback while there is no deposit made")

				// Check account is converted to a normal account
				acc = s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
				Expect(acc).ToNot(BeNil(), "expected account to exist")
				_, isClawback = acc.(*types.ClawbackVestingAccount)
				Expect(isClawback).To(BeFalse(), "expected account to be a normal account")
			})

			It("should remove the store entry after the deposit period ends", func() {
				s.network.NextBlockAfter(time.Hour * 24 * 365) // one year
				// Commit again because EndBlocker is run with time of the previous block and gov proposals are ended in EndBlocker
				Expect(s.network.NextBlock()).To(BeNil())

				// Check that the proposal has ended -- since deposit failed it's removed from the store
				_, found := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), clawbackProposalID)
				Expect(found).To(BeFalse(), "expected proposal not to be found")

				// Check that the store entry was removed
				hasActiveProposal := s.network.App.VestingKeeper.HasActiveClawbackProposal(s.network.GetContext(), vestingAddr)
				Expect(hasActiveProposal).To(BeFalse(),
					"expected no active clawback proposal for address %q",
					vestingAddr.String(),
				)
			})
		})
	})

	It("should update vesting funder and claw back unvested amount before cliff", func() {
		ctx := s.network.GetContext()
		newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
		Expect(err).To(BeNil())
		balanceNewFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee := balRes.Balance

		// Update clawback vesting account funder
		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
		_, err = s.network.App.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
		s.Require().NoError(err)

		// Perform clawback before cliff - funds should go to new funder (no dest address defined)
		msg := types.NewMsgClawback(newFunder, vestingAddr, sdk.AccAddress([]byte{}))
		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())
		Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

		// All initial vesting amount goes to funder
		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
		Expect(err).To(BeNil())
		bNewF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		bG := balRes.Balance

		// Original funder balance should not change
		Expect(bF).To(Equal(balanceFunder))
		// New funder should get the vested tokens
		Expect(balanceNewFunder.Add(vestingAmtTotal[0]).Amount).To(Equal(bNewF.Amount))
		Expect(balanceGrantee.Sub(vestingAmtTotal[0]).Amount).To(Equal(bG.Amount))
	})

	It("should update vesting funder and first funder cannot claw back unvested before cliff", func() {
		ctx := s.network.GetContext()
		newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

		balRes, err := s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		balanceFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
		Expect(err).To(BeNil())
		balanceNewFunder := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		balanceGrantee := balRes.Balance

		// Update clawback vesting account funder
		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
		_, err = s.network.App.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
		s.Require().NoError(err)

		// Original funder tries to perform clawback before cliff - is not the current funder
		msg := types.NewMsgClawback(funder, vestingAddr, sdk.AccAddress([]byte{}))
		_, err = s.network.App.VestingKeeper.Clawback(ctx, msg)
		s.Require().Error(err)

		// All balances should remain the same
		balRes, err = s.handler.GetBalance(funder, stakeDenom)
		Expect(err).To(BeNil())
		bF := balRes.Balance
		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
		Expect(err).To(BeNil())
		bNewF := balRes.Balance
		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
		Expect(err).To(BeNil())
		bG := balRes.Balance

		Expect(bF).To(Equal(balanceFunder))
		Expect(balanceNewFunder).To(Equal(bNewF))
		Expect(balanceGrantee).To(Equal(bG))
	})

	Context("governance clawback to community pool", func() {
		It("should claw back unvested amount before cliff", func() {
			ctx := s.network.GetContext()

			// initial balances
			balRes, err := s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			balanceFunder := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			balanceGrantee := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			balanceDest := balRes.Balance

			pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())
			balanceCommPool := pool.CommunityPool[0]

			// Perform clawback before cliff
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

			// All initial vesting amount goes to community pool instead of dest
			balRes, err = s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			bF := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			bG := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			bD := balRes.Balance

			pool, err = s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())
			bCP := pool.CommunityPool[0]

			Expect(bF).To(Equal(balanceFunder))
			Expect(balanceGrantee.Sub(vestingAmtTotal[0]).Amount).To(Equal(bG.Amount))
			// destination address should remain unchanged
			Expect(balanceDest.Amount).To(Equal(bD.Amount))
			// vesting amount should go to community pool
			Expect(balanceCommPool.Amount.Add(math.LegacyNewDec(vestingAmtTotal[0].Amount.Int64()))).To(Equal(bCP.Amount))
			Expect(stakeDenom).To(Equal(bCP.Denom))
		})

		It("should claw back any unvested amount after cliff before unlocking", func() {
			// Surpass cliff but not lockup duration
			cliffDuration := time.Duration(cliffLength)
			s.network.NextBlockAfter(cliffDuration * time.Second)

			// Check that all tokens are locked and some, but not all tokens are vested
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
			vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
			expVestedAmount := amt.Mul(math.NewInt(cliff))
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
			unvested := vestingAmtTotal.Sub(vested...)

			Expect(expVested).To(Equal(vested))
			s.Require().True(expVestedAmount.GT(math.NewInt(0)))
			s.Require().True(free.IsZero())
			Expect(vesting).To(Equal(vestingAmtTotal))

			balRes, err := s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			balanceFunder := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			balanceGrantee := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			balanceDest := balRes.Balance

			pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())

			balanceCommPool := pool.CommunityPool[0]

			testClawbackAccount := TestClawbackAccount{
				privKey:         nil,
				address:         vestingAddr,
				clawbackAccount: clawbackAccount,
			}
			// stake vested tokens
			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err = s.factory.Delegate(
				testClawbackAccount.privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// Perform clawback
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			ctx := s.network.GetContext()
			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(unvested), "expected unvested coins to be clawed back")

			balRes, err = s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			bF := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			bG := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			bD := balRes.Balance

			pool, err = s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())

			bCP := pool.CommunityPool[0]

			expClawback := clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())

			// Any unvested amount is clawed back to community pool
			Expect(balanceFunder).To(Equal(bF))
			Expect(balanceGrantee.Sub(expClawback[0]).Amount).To(Equal(bG.Amount))
			Expect(balanceDest.Amount).To(Equal(bD.Amount))
			// vesting amount should go to community pool
			Expect(balanceCommPool.Amount.Add(math.LegacyNewDec(expClawback[0].Amount.Int64()))).To(Equal(bCP.Amount))
			Expect(stakeDenom).To(Equal(bCP.Denom))
		})

		It("should claw back any unvested amount after cliff and unlocking", func() {
			// Surpass lockup duration
			// A strict `if t < clawbackTime` comparison is used in ComputeClawback
			// so, we increment the duration with 1 for the free token calculation to match
			lockupDuration := time.Duration(lockupLength + 1)
			s.network.NextBlockAfter(lockupDuration * time.Second)

			// Check if some, but not all tokens are vested and unlocked
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
			vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
			expVestedAmount := amt.Mul(math.NewInt(lockup))
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
			unvested := vestingAmtTotal.Sub(vested...)

			Expect(free).To(Equal(vested))
			Expect(expVested).To(Equal(vested))
			s.Require().True(expVestedAmount.GT(math.NewInt(0)))
			Expect(vesting).To(Equal(unvested))

			balRes, err := s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			balanceFunder := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			balanceGrantee := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			balanceDest := balRes.Balance

			testClawbackAccount := TestClawbackAccount{
				privKey:         nil,
				address:         vestingAddr,
				clawbackAccount: clawbackAccount,
			}
			// stake vested tokens
			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			err = s.factory.Delegate(
				testClawbackAccount.privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).To(BeNil())

			// Perform clawback
			msg := types.NewMsgClawback(funder, vestingAddr, dest)
			ctx := s.network.GetContext()
			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(unvested), "expected only unvested coins to be clawed back")

			balRes, err = s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			bF := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			bG := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			bD := balRes.Balance

			// Any unvested amount is clawed back
			Expect(balanceFunder).To(Equal(bF))
			Expect(balanceGrantee.Sub(vesting[0]).Amount).To(Equal(bG.Amount))
			Expect(balanceDest.Add(vesting[0]).Amount).To(Equal(bD.Amount))
		})

		It("should not claw back any amount after vesting periods end", func() {
			// Surpass vesting periods
			vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
			s.network.NextBlockAfter(vestingDuration * time.Second)

			// Check if some, but not all tokens are vested and unlocked
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
			vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())

			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))
			unvested := vestingAmtTotal.Sub(vested...)

			Expect(free).To(Equal(vested))
			Expect(expVested).To(Equal(vested))
			Expect(expVested).To(Equal(vestingAmtTotal))
			Expect(unlocked).To(Equal(vestingAmtTotal))
			Expect(vesting).To(Equal(unvested))
			Expect(vesting.IsZero()).To(BeTrue())

			balRes, err := s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			balanceFunder := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			balanceGrantee := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			balanceDest := balRes.Balance

			pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())

			balanceCommPool := pool.CommunityPool[0]

			testClawbackAccount := TestClawbackAccount{
				privKey:         nil,
				address:         vestingAddr,
				clawbackAccount: clawbackAccount,
			}
			// stake vested tokens
			ok, vestedCoin := vested.Find(stakeDenom)
			Expect(ok).To(BeTrue())
			err = s.factory.Delegate(
				testClawbackAccount.privKey,
				s.network.GetValidators()[0].OperatorAddress,
				vestedCoin,
			)
			Expect(err).To(BeNil())

			// Perform clawback
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			ctx := s.network.GetContext()
			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil(), "expected no error during clawback")
			Expect(res.Coins).To(BeEmpty(), "expected nothing to be clawed back after end of vesting schedules")

			balRes, err = s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			bF := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			bG := balRes.Balance
			balRes, err = s.handler.GetBalance(dest, stakeDenom)
			Expect(err).To(BeNil())
			bD := balRes.Balance

			pool, err = s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())

			bCP := pool.CommunityPool[0]

			// No amount is clawed back
			Expect(balanceFunder).To(Equal(bF))
			Expect(balanceGrantee).To(Equal(bG))
			Expect(balanceDest).To(Equal(bD))
			Expect(balanceCommPool.Amount).To(Equal(bCP.Amount))
		})

		It("should update vesting funder and claw back unvested amount before cliff", func() {
			ctx := s.network.GetContext()
			newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

			balRes, err := s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			balanceFunder := balRes.Balance
			balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
			Expect(err).To(BeNil())
			balanceNewFunder := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			balanceGrantee := balRes.Balance

			pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())

			balanceCommPool := pool.CommunityPool[0]

			// Update clawback vesting account funder
			updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
			_, err = s.network.App.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
			s.Require().NoError(err)

			// Perform clawback before cliff - funds should go to new funder (no dest address defined)
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, nil)
			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
			Expect(err).To(BeNil())
			Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

			// All initial vesting amount goes to funder
			balRes, err = s.handler.GetBalance(funder, stakeDenom)
			Expect(err).To(BeNil())
			bF := balRes.Balance
			balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
			Expect(err).To(BeNil())
			bNewF := balRes.Balance
			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
			Expect(err).To(BeNil())
			bG := balRes.Balance

			pool, err = s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
			Expect(err).To(BeNil())

			bCP := pool.CommunityPool[0]

			// Original funder balance should not change
			Expect(bF).To(Equal(balanceFunder))
			// New funder should not get the vested tokens
			Expect(balanceNewFunder.Amount).To(Equal(bNewF.Amount))
			Expect(balanceGrantee.Sub(vestingAmtTotal[0]).Amount).To(Equal(bG.Amount))
			// vesting amount should go to community pool
			Expect(balanceCommPool.Amount.Add(math.LegacyNewDec(vestingAmtTotal[0].Amount.Int64()))).To(Equal(bCP.Amount))
		})

		It("should not claw back when governance clawback is disabled", func() {
			// disable governance clawback
			s.network.App.VestingKeeper.SetGovClawbackDisabled(s.network.GetContext(), vestingAddr)

			// Perform clawback before cliff
			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
			_, err := s.network.App.VestingKeeper.Clawback(s.network.GetContext(), msg)
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
		s            *KeeperTestSuite
		contractAddr common.Address
		contract     evmtypes.CompiledContract
		err          error
	)

	BeforeEach(func() {
		s = new(KeeperTestSuite)
		s.SetupTest()
		contract = contracts.ERC20MinterBurnerDecimalsContract
		contractAddr, err = testutil.DeployContract(
			s.network.GetContext(),
			s.network.App,
			s.keyring.GetPrivKey(0),
			s.network.GetEvmClient(),
			contract,
			"Test", "TTT", uint8(18),
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")
	})

	It("should not convert a smart contract to a clawback vesting account", func() {
		msgCreate := types.NewMsgCreateClawbackVestingAccount(
			s.keyring.GetAccAddr(0),
			contractAddr.Bytes(),
			false,
		)
		_, err := s.network.App.VestingKeeper.CreateClawbackVestingAccount(s.network.GetContext(), msgCreate)
		Expect(err).To(HaveOccurred(), "expected error")
		Expect(err.Error()).To(ContainSubstring(
			fmt.Sprintf(
				"account %s is a contract account and cannot be converted in a clawback vesting account",
				sdk.AccAddress(contractAddr.Bytes()).String()),
		))

		// Check that the account was not converted
		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), contractAddr.Bytes())
		Expect(acc).ToNot(BeNil(), "smart contract should be found")
		_, ok := acc.(*types.ClawbackVestingAccount)
		Expect(ok).To(BeFalse(), "account should not be a clawback vesting account")

		// Check that the contract code was not deleted
		//
		// NOTE: When it was possible to create clawback vesting accounts for smart contracts,
		// the contract code was deleted from the EVM state. This checks that this is not the case.
		res, err := s.network.App.EvmKeeper.Code(s.network.GetContext(), &evmtypes.QueryCodeRequest{Address: contractAddr.String()})
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
		s *KeeperTestSuite
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
		s = new(KeeperTestSuite)
		s.SetupTest()

		// Initialize the account at the vesting address and the funder accounts by funding them
		fundedCoins := sdk.Coins{{Denom: utils.BaseDenom, Amount: math.NewInt(2e18)}} // fund more than what is sent to the vesting account for transaction fees
		err = testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, vestingAddr, fundedCoins)
		Expect(err).ToNot(HaveOccurred(), "failed to fund account")
		err = testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, funder, fundedCoins)
		Expect(err).ToNot(HaveOccurred(), "failed to fund account")

		// Create a clawback vesting account
		msgCreate := types.NewMsgCreateClawbackVestingAccount(
			funder,
			vestingAddr,
			false,
		)

		res, err := testutil.DeliverTx(s.network.GetContext(), s.network.App, vestingPriv, &gasPrice, msgCreate)
		Expect(err).ToNot(HaveOccurred(), "failed to create clawback vesting account")
		txCost = gasPrice.Int64() * res.GasWanted

		// Check clawback acccount was created
		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
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
					s.network.GetContext().BlockTime(),
					lockupPeriods,
					vestingPeriods,
				)

				// Deliver transaction with message
				res, err := testutil.DeliverTx(s.network.GetContext(), s.network.App, funderPriv, nil, msg)

				// Get account at the new address
				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
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
					balRes, err := s.handler.GetBalance(vestingAddr, utils.BaseDenom)
					Expect(err).To(BeNil())
					balance := balRes.Balance
					expBalance := int64(2e18) + int64(1e18) - txCost // fundedCoins + vestingCoins - txCost
					Expect(balance.Amount.Int64()).To(Equal(expBalance), "vesting account has incorrect balance")
				}
			})
		}
	})
})
