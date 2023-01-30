package keeper_test

import (
	"math/big"
	"strings"
	"time"

	"cosmossdk.io/math"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/app/ante"
	"github.com/evmos/evmos/v11/testutil"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	claimstypes "github.com/evmos/evmos/v11/x/claims/types"

	"github.com/evmos/evmos/v11/x/vesting/types"
)

type TestClawbackAccount struct {
	privKey         *ethsecp256k1.PrivKey
	address         sdk.AccAddress
	clawbackAccount *types.ClawbackVestingAccount
}

// Clawback vesting with Cliff and Lock. In this case the cliff is reached
// before the lockup period is reached to represent the scenario in which an
// employee starts before mainnet launch (periodsCliff < lockupPeriod)

// Example:
// 21/10 Employee joins Evmos and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
var _ = Describe("Clawback Vesting Accounts", Ordered, func() {
	// Monthly vesting period
	stakeDenom := claimstypes.DefaultParams().ClaimsDenom
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
	lockup := int64(12) // 12 month
	lockupLength := vestingLength * lockup
	// Unlock at 12 and 24 months
	numLockupPeriods := int64(2)
	// Unlock 1/4th of the total vest in each unlock event
	unlockedPerLockup := vestingAmtTotal.QuoInt(math.NewInt(4))
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
		address, privKey := createAddressKey()
		testAccounts[i] = TestClawbackAccount{
			privKey: privKey,
			address: address,
		}
	}

	accountGasCoverage := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(1e16)))

	var (
		clawbackAccount *types.ClawbackVestingAccount
		unvested        sdk.Coins
		vested          sdk.Coins
	)

	dest := sdk.AccAddress(tests.GenerateAddress().Bytes())
	funder := sdk.AccAddress(tests.GenerateAddress().Bytes())

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
		It("cannot delegate tokens", func() {
			err := delegate(clawbackAccount, math.NewInt(100))
			Expect(err).ToNot(BeNil())
		})

		It("cannot transfer tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				dest,
				unvested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot perform Ethereum tx", func() {
			err := validateAnteForEthTx(clawbackAccount, nil)
			Expect(err).ToNot(BeNil())
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
		})

		It("can delegate vested tokens", func() {
			err := delegate(clawbackAccount, vested.AmountOf(stakeDenom))
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			err := delegate(clawbackAccount, vestingAmtTotal.AmountOf(stakeDenom))
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

		It("cannot perform Ethereum tx", func() {
			err := validateAnteForEthTx(clawbackAccount, nil)
			Expect(err).ToNot(BeNil())
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
			grantee := testAccounts[0].address

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			granteeBalance := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			txAmount := unlockedPerLockup.AmountOf(stakeDenom)
			msg := createEthTx(nil, grantee, dest, txAmount.BigInt(), 0)
			err := validateAnteForEthTxs(msg)
			Expect(err).To(BeNil())

			// Expect delivery to succeed, then compare balances
			err = deliverEthTxs(testAccounts[0].privKey, msg)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			gb := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().GreaterOrEqual(granteeBalance.Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Uint64(), db.Amount.Uint64())
		})

		It("should enable access to unlocked EVM tokens (single-account, multiple-msgs)", func() {
			account := testAccounts[0]

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			granteeBalance := s.app.BankKeeper.GetBalance(s.ctx, account.address, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			// Split the total unlocked amount into numMsgs equally sized tx's
			numMsgs := 3
			msgs := make([]sdk.Msg, numMsgs)
			txAmount := unlockedPerLockup.AmountOf(stakeDenom).QuoRaw(int64(numMsgs))

			for i := 0; i < numMsgs; i++ {
				msgs[i] = createEthTx(nil, account.address, dest, txAmount.BigInt(), i)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Expect delivery to succeed, then compare balances
			err = deliverEthTxs(testAccounts[0].privKey, msgs...)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			gb := s.app.BankKeeper.GetBalance(s.ctx, account.address, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().GreaterOrEqual(granteeBalance.Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Uint64(), db.Amount.Uint64())
		})

		It("should enable access to unlocked EVM tokens (multi-account, single-msg)", func() {
			txAmount := unlockedPerLockup.AmountOf(stakeDenom)

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			granteeBalances := make(sdk.Coins, numTestAccounts)

			msgs := make([]sdk.Msg, numTestAccounts)
			for i, grantee := range testAccounts {
				msgs[i] = createEthTx(grantee.privKey, grantee.address, dest, txAmount.BigInt(), 0)
				granteeBalances[i] = s.app.BankKeeper.GetBalance(s.ctx, grantee.address, stakeDenom)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Expect delivery to succeed, then compare balances
			err = deliverEthTxs(nil, msgs...)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Mul(math.NewInt(int64(numTestAccounts))), db.Amount)

			for i, account := range testAccounts {
				gb := s.app.BankKeeper.GetBalance(s.ctx, account.address, stakeDenom)
				s.Require().GreaterOrEqual(granteeBalances[i].Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			}
		})

		It("should enable access to unlocked EVM tokens (multi-account, multiple-msgs)", func() {
			numMsgs := 3
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockup.AmountOf(stakeDenom).QuoRaw(int64(numMsgs))

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
			granteeBalances := make(sdk.Coins, numTestAccounts)

			for i, grantee := range testAccounts {
				for j := 0; j < numMsgs; j++ {
					msgs = append(msgs, createEthTx(grantee.privKey, grantee.address, dest, txAmount.BigInt(), j))
				}
				granteeBalances[i] = s.app.BankKeeper.GetBalance(s.ctx, grantee.address, stakeDenom)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Expect delivery to succeed, then compare balances
			err = deliverEthTxs(nil, msgs...)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Mul(math.NewInt(int64(numTestAccounts))), db.Amount)

			for i, account := range testAccounts {
				gb := s.app.BankKeeper.GetBalance(s.ctx, account.address, stakeDenom)
				s.Require().GreaterOrEqual(granteeBalances[i].Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			}
		})

		It("should not enable access to locked EVM tokens (single-account, single-msg)", func() {
			// Attempt to spend entire balance
			txAmount := vestingAmtTotal.AmountOf(stakeDenom)
			msg := createEthTx(nil, clawbackAccount.GetAddress(), dest, txAmount.BigInt(), 0)
			err := validateAnteForEthTxs(msg)
			Expect(err).ToNot(BeNil())

			err = deliverEthTxs(nil, msg)
			Expect(err).ToNot(BeNil())
		})

		It("should not enable access to locked EVM tokens (single-account, multiple-msgs)", func() {
			numMsgs := 3
			msgs := make([]sdk.Msg, numMsgs+1)
			txAmount := unlockedPerLockup.AmountOf(stakeDenom).QuoRaw(int64(numMsgs))

			// Add additional message that exceeds unlocked balance
			for i := 0; i < numMsgs+1; i++ {
				msgs[i] = createEthTx(nil, clawbackAccount.GetAddress(), dest, txAmount.BigInt(), i)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).ToNot(BeNil())

			err = deliverEthTxs(nil, msgs...)
			Expect(err).ToNot(BeNil())
		})

		It("should not enable access to locked EVM tokens (multi-account, single-msg)", func() {
			msgs := make([]sdk.Msg, numTestAccounts+1)
			txAmount := unlockedPerLockup.AmountOf(stakeDenom)

			for i, account := range testAccounts {
				msgs[i] = createEthTx(account.privKey, account.address, dest, txAmount.BigInt(), 0)
			}

			// Add additional message that exceeds unlocked balance
			msgs[numTestAccounts] = createEthTx(nil, clawbackAccount.GetAddress(), dest, txAmount.BigInt(), 1)

			err := validateAnteForEthTxs(msgs...)
			Expect(err).ToNot(BeNil())

			err = deliverEthTxs(nil, msgs...)
			Expect(err).ToNot(BeNil())
		})

		It("should not enable access to locked EVM tokens (multi-account, multiple-msgs)", func() {
			numMsgs := 3
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockup.AmountOf(stakeDenom).QuoRaw(int64(numMsgs))

			for _, account := range testAccounts {
				for j := 0; j < numMsgs; j++ {
					msgs = append(msgs, createEthTx(account.privKey, account.address, dest, txAmount.BigInt(), j))
				}
			}

			// Add additional message that exceeds unlocked balance
			msgs = append(msgs, createEthTx(nil, clawbackAccount.GetAddress(), dest, txAmount.BigInt(), numMsgs))

			err := validateAnteForEthTxs(msgs...)
			Expect(err).ToNot(BeNil())

			err = deliverEthTxs(nil, msgs...)
			Expect(err).ToNot(BeNil())
		})

		It("should not short-circuit the AnteHandler with a normal account", func() {
			account := testAccounts[0]
			address, privKey := createAddressKey()

			txAmount := vestingAmtTotal.AmountOf(stakeDenom)

			// Fund a normal account to try to short-circuit the AnteHandler
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, address, vestingAmtTotal.MulInt(sdk.NewInt(2)))
			Expect(err).To(BeNil())
			normalAccMsg := createEthTx(privKey, address, dest, txAmount.BigInt(), 0)

			// Attempt to spend entire balance
			msg := createEthTx(account.privKey, account.address, dest, txAmount.BigInt(), 0)
			err = validateAnteForEthTxs(normalAccMsg, msg)
			Expect(err).ToNot(BeNil())

			err = deliverEthTxs(nil, msg)
			Expect(err).ToNot(BeNil())
		})

		It("should short-circuit with zero balance", func() {
			address, privKey := createAddressKey()
			// Fund a normal account to create it in state
			balance := vestingAmtTotal
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, address, balance)
			Expect(err).To(BeNil())

			// Drain account balance
			err = s.app.BankKeeper.SendCoins(s.ctx, address, dest, balance)
			Expect(err).ToNot(BeNil())

			msg := createEthTx(privKey, address, dest, big.NewInt(0), 0)
			err = validateAnteForEthTxs(msg)
			Expect(err).ToNot(BeNil())
			Expect(strings.Contains(err.Error(), "no balance")).To(BeTrue())
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
			err := delegate(clawbackAccount, vested.AmountOf(stakeDenom))
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			err := delegate(clawbackAccount, vestingAmtTotal.AmountOf(stakeDenom))
			Expect(err).ToNot(BeNil())
		})

		It("can transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				clawbackAccount.GetAddress(),
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vestingAmtTotal,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform ethereum tx", func() {
			err := validateAnteForEthTx(clawbackAccount, nil)
			Expect(err).To(BeNil())
		})
	})
})

// Example:
// 21/10 Employee joins Evmos and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
var _ = Describe("Clawback Vesting Accounts - claw back tokens", Ordered, func() {
	// Monthly vesting period
	stakeDenom := claimstypes.DefaultParams().ClaimsDenom
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
	)
	grantee := sdk.AccAddress(tests.GenerateAddress().Bytes())
	funder := sdk.AccAddress(tests.GenerateAddress().Bytes())
	dest := sdk.AccAddress(tests.GenerateAddress().Bytes())

	BeforeEach(func() {
		s.SetupTest()
		ctx := sdk.WrapSDKContext(s.ctx)

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime()
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, vestingAmtTotal)
		s.Require().NoError(err)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
		s.Require().True(balanceFunder.IsGTE(vestingAmtTotal[0]))
		s.Require().Equal(balanceGrantee, sdk.NewInt64Coin(stakeDenom, 0))
		s.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))

		msg := types.NewMsgCreateClawbackVestingAccount(funder, grantee, vestingStart, lockupPeriods, vestingPeriods, true)

		_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
		s.Require().NoError(err)

		acc := s.app.AccountKeeper.GetAccount(s.ctx, grantee)
		clawbackAccount, _ = acc.(*types.ClawbackVestingAccount)

		// Check if all tokens are unvested and locked at vestingStart
		vesting = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
		s.Require().Equal(vestingAmtTotal, vesting)
		s.Require().True(vested.IsZero())
		s.Require().True(unlocked.IsZero())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee = s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		balanceDest = s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		s.Require().True(bF.IsGTE(balanceFunder.Sub(vestingAmtTotal[0])))
		s.Require().True(balanceGrantee.IsGTE(vestingAmtTotal[0]))
		s.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))
	})

	It("should claw back unvested amount before cliff", func() {
		ctx := sdk.WrapSDKContext(s.ctx)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// Perform clawback before cliff
		msg := types.NewMsgClawback(funder, grantee, dest)
		_, err := s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		// All initial vesting amount goes to dest
		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		s.Require().Equal(bF, balanceFunder)
		s.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
		s.Require().Equal(balanceDest.Add(vestingAmtTotal[0]).Amount.Uint64(), bD.Amount.Uint64())
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
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// stake vested tokens
		err := delegate(clawbackAccount, vested.AmountOf(stakeDenom))
		Expect(err).To(BeNil())

		// Perform clawback
		msg := types.NewMsgClawback(funder, grantee, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
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
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// stake vested tokens
		err := delegate(clawbackAccount, vested.AmountOf(stakeDenom))
		Expect(err).To(BeNil())

		// Perform clawback
		msg := types.NewMsgClawback(funder, grantee, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
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
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// stake vested tokens
		err := delegate(clawbackAccount, vested.AmountOf(stakeDenom))
		Expect(err).To(BeNil())

		// Perform clawback
		msg := types.NewMsgClawback(funder, grantee, dest)
		ctx := sdk.WrapSDKContext(s.ctx)
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		Expect(err).To(BeNil())

		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
		bD := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

		// No amount is clawed back
		s.Require().Equal(balanceFunder, bF)
		s.Require().Equal(balanceGrantee, bG)
		s.Require().Equal(balanceDest, bD)
	})

	It("should update vesting funder and claw back unvested amount before cliff", func() {
		ctx := sdk.WrapSDKContext(s.ctx)
		newFunder := sdk.AccAddress(tests.GenerateAddress().Bytes())

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceNewFunder := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)

		// Update clawback vesting account funder
		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, grantee)
		_, err := s.app.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
		s.Require().NoError(err)

		// Perform clawback before cliff - funds should go to new funder (no dest address defined)
		msg := types.NewMsgClawback(newFunder, grantee, sdk.AccAddress([]byte{}))
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		s.Require().NoError(err)

		// All initial vesting amount goes to funder
		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bNewF := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)

		// Original funder balance should not change
		s.Require().Equal(bF, balanceFunder)
		// New funder should get the vested tokens
		s.Require().Equal(balanceNewFunder.Add(vestingAmtTotal[0]).Amount.Uint64(), bNewF.Amount.Uint64())
		s.Require().Equal(balanceGrantee.Sub(vestingAmtTotal[0]).Amount.Uint64(), bG.Amount.Uint64())
	})

	It("should update vesting funder and first funder cannot claw back unvested before cliff", func() {
		ctx := sdk.WrapSDKContext(s.ctx)
		newFunder := sdk.AccAddress(tests.GenerateAddress().Bytes())

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceNewFunder := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)

		// Update clawback vesting account funder
		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, grantee)
		_, err := s.app.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
		s.Require().NoError(err)

		// Original funder tries to perform clawback before cliff - is not the current funder
		msg := types.NewMsgClawback(funder, grantee, sdk.AccAddress([]byte{}))
		_, err = s.app.VestingKeeper.Clawback(ctx, msg)
		s.Require().Error(err)

		// All balances should remain the same
		bF := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		bNewF := s.app.BankKeeper.GetBalance(s.ctx, newFunder, stakeDenom)
		bG := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)

		s.Require().Equal(bF, balanceFunder)
		s.Require().Equal(balanceNewFunder, bNewF)
		s.Require().Equal(balanceGrantee, bG)
	})
})

func createAddressKey() (sdk.AccAddress, *ethsecp256k1.PrivKey) {
	address, privKey := tests.NewAddrKey()
	return address.Bytes(), &ethsecp256k1.PrivKey{Key: privKey.Bytes()}
}

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func delegate(clawbackAccount *types.ClawbackVestingAccount, amount math.Int) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)
	//
	val, err := sdk.ValAddressFromBech32("evmosvaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn")
	s.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(claimstypes.DefaultParams().ClaimsDenom, amount))
	err = txBuilder.SetMsgs(delegateMsg)
	s.Require().NoError(err)
	tx := txBuilder.GetTx()

	dec := ante.NewVestingDelegationDecorator(s.app.AccountKeeper, s.app.StakingKeeper, types.ModuleCdc)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

func createEthTx(privKey *ethsecp256k1.PrivKey, from sdk.AccAddress, dest sdk.AccAddress, amount *big.Int, nonceIncrement int) *evmtypes.MsgEthereumTx {
	toAddr := common.BytesToAddress(dest.Bytes())
	fromAddr := common.BytesToAddress(from.Bytes())
	chainID := s.app.EvmKeeper.ChainID()

	// When we send multiple Ethereum Tx's in one Cosmos Tx, we need to increment the nonce for each one.
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, fromAddr) + uint64(nonceIncrement)
	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &toAddr, amount, 100000, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = fromAddr.String()

	// If we are creating multiple eth Tx's with different senders, we need to sign here rather than later.
	if privKey != nil {
		signer := ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())
		err := msgEthereumTx.Sign(signer, tests.NewSigner(privKey))
		s.Require().NoError(err)
	}

	return msgEthereumTx
}

func validateAnteForEthTxs(msgs ...sdk.Msg) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	tx := txBuilder.GetTx()

	// Call Ante decorator
	dec := ante.NewEthVestingTransactionDecorator(s.app.AccountKeeper, s.app.BankKeeper, s.app.EvmKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

func deliverEthTxs(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) error {
	_, err := testutil.DeliverEthTx(s.ctx, s.app, priv, msgs...)
	return err
}

// validateAnteForEthTx checks a simple single-message Ethereum transaction against the EVM Vesting AnteHandler
func validateAnteForEthTx(clawbackAccount *types.ClawbackVestingAccount, amount *big.Int) error {
	address := clawbackAccount.GetAddress()
	msg := createEthTx(nil, address, address, amount, 0)

	return validateAnteForEthTxs(msg)
}
