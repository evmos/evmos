package keeper_test

import (
	"math/big"
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

type GranteeSignerAccount struct {
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
	amt := sdk.NewInt(1)
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
		unvested        sdk.Coins
		vested          sdk.Coins
	)

	BeforeEach(func() {
		s.SetupTest()

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime()
		baseAccount := authtypes.NewBaseAccountWithAddress(addr)
		funder := sdk.AccAddress(types.ModuleName)
		clawbackAccount = types.NewClawbackVestingAccount(
			baseAccount,
			funder,
			vestingAmtTotal,
			vestingStart,
			lockupPeriods,
			vestingPeriods,
		)
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, addr, vestingAmtTotal)
		s.Require().NoError(err)
		acc := s.app.AccountKeeper.NewAccount(s.ctx, clawbackAccount)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)

		// Check if all tokens are unvested at vestingStart
		unvested = clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		s.Require().Equal(vestingAmtTotal, unvested)
		s.Require().True(vested.IsZero())
	})

	Context("before first vesting period", func() {
		It("cannot delegate tokens", func() {
			err := delegate(clawbackAccount, 100)
			Expect(err).ToNot(BeNil())
		})

		It("cannot transfer tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
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
			err := delegate(clawbackAccount, vested.AmountOf(stakeDenom).Int64())
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			err := delegate(clawbackAccount, vestingAmtTotal.AmountOf(stakeDenom).Int64())
			Expect(err).ToNot(BeNil())
		})

		It("cannot transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot perform Ethereum tx", func() {
			err := validateAnteForEthTx(clawbackAccount, nil)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("after first vesting period and lockup", func() {
		BeforeEach(func() {
			// Surpass lockup duration
			lockupDuration := time.Duration(lockupLength)
			s.CommitAfter(lockupDuration * time.Second)

			// Check if some, but not all tokens are vested
			unvested = clawbackAccount.GetUnvestedOnly(s.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)
		})

		It("can delegate vested tokens", func() {
			err := delegate(clawbackAccount, vested.AmountOf(stakeDenom).Int64())
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			err := delegate(clawbackAccount, vestingAmtTotal.AmountOf(stakeDenom).Int64())
			Expect(err).ToNot(BeNil())
		})

		It("can transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				unvested,
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
		err := delegate(clawbackAccount, vested.AmountOf(stakeDenom).Int64())
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
		err := delegate(clawbackAccount, vested.AmountOf(stakeDenom).Int64())
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
		err := delegate(clawbackAccount, vested.AmountOf(stakeDenom).Int64())
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

// Ensure the vesting account has access to unlocked tokens in EVM interactions
var _ = Describe("Clawback Vesting Accounts - Unlocked EVM Tokens", Ordered, func() {
	// Monthly vesting period
	stakeDenom := claimstypes.DefaultParams().ClaimsDenom
	amt := sdk.NewInt(1e18)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 2 years vesting total
	periodsTotal := int64(24)
	vestingTotal := amt.Mul(sdk.NewInt(periodsTotal))
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, vestingTotal))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12-month and 24-month lockups
	lockup := int64(12) // 12 months
	lockupLength := vestingLength * lockup
	numLockupPeriods := 2
	// Unlock the corresponding fraction of the total in each period
	unlockedPerLockup := vestingAmtTotal.QuoInt(math.NewInt(int64(numLockupPeriods)))
	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: unlockedPerLockup}
	lockupPeriods := sdkvesting.Periods{}
	for i := 0; i < numLockupPeriods; i++ {
		lockupPeriods = append(lockupPeriods, lockupPeriod)
	}

	// Create vesting periods with initial cliff
	vestingPeriods := sdkvesting.Periods{cliffPeriod}
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	var (
		vesting  sdk.Coins
		vested   sdk.Coins
		unlocked sdk.Coins
	)

	// Test with multiple accounts to ensure that none of them exceed the locked total
	numAccounts := 2
	clawbackAccounts := make([]*types.ClawbackVestingAccount, numAccounts)

	granteeAccounts := make([]GranteeSignerAccount, numAccounts)
	for i := range granteeAccounts {
		address, privKey := tests.NewAddrKey()
		granteeAccounts[i] = GranteeSignerAccount{
			privKey: &ethsecp256k1.PrivKey{
				Key: privKey.Bytes(),
			},
			address: address.Bytes(),
		}
	}

	granteeGasStipend := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(1e17)))

	funder := sdk.AccAddress(tests.GenerateAddress().Bytes())
	dest := sdk.AccAddress(tests.GenerateAddress().Bytes())

	BeforeEach(func() {
		s.SetupTest()
		ctx := sdk.WrapSDKContext(s.ctx)

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime()
		funderTotalBalance := vestingAmtTotal.MulInt(sdk.NewInt(int64(numAccounts)))
		err := testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, funderTotalBalance)
		s.Require().NoError(err)

		balanceFunder := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
		balanceDest := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)
		s.Require().True(balanceFunder.IsGTE(vestingAmtTotal[0]))
		s.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))

		// Initialize all ClawbackVestingAccounts
		for i := 0; i < numAccounts; i++ {
			grantee := granteeAccounts[i].address
			balanceGrantee := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			s.Require().Equal(balanceGrantee, sdk.NewInt64Coin(stakeDenom, 0))

			msg := types.NewMsgCreateClawbackVestingAccount(funder, grantee, vestingStart, lockupPeriods, vestingPeriods, true)

			_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(ctx, msg)
			s.Require().NoError(err)

			acc := s.app.AccountKeeper.GetAccount(s.ctx, grantee)
			clawbackAccount, _ := acc.(*types.ClawbackVestingAccount)
			// Set reference to clawbackAccount
			granteeAccounts[i].clawbackAccount = clawbackAccount

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

			totalSpentByFunder := vestingAmtTotal[0].Amount.Mul(math.NewInt(int64(i + 1)))

			s.Require().True(bF.IsGTE(balanceFunder.SubAmount(totalSpentByFunder)))
			s.Require().True(balanceGrantee.IsGTE(vestingAmtTotal[0]))
			s.Require().Equal(balanceDest, sdk.NewInt64Coin(stakeDenom, 0))

			clawbackAccounts[i] = clawbackAccount

			// Grant gas stipend to cover EVM fees
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, clawbackAccount.GetAddress(), granteeGasStipend)
			s.Require().NoError(err)
			balanceGrantee = s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			s.Require().Equal(balanceGrantee, granteeGasStipend[0].Add(vestingAmtTotal[0]))
		}
	})

	Context("After first unlock", func() {
		BeforeEach(func() {
			// Surpass cliff and first lockup
			vestDuration := time.Duration(lockupLength)
			s.CommitAfter(vestDuration * time.Second)

			// Check if some, but not all tokens are vested and unlocked
			for _, clawbackAccount := range clawbackAccounts {
				vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
				unlocked = clawbackAccount.GetUnlockedOnly(s.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup))))

				s.Require().NotEqual(vestingAmtTotal, vested)
				s.Require().Equal(expVested, vested)
				s.Require().Equal(unlocked, unlockedPerLockup)
			}
		})

		It("should enable access to unlocked EVM tokens (single-account)", func() {
			clawbackAccount := granteeAccounts[0].clawbackAccount
			grantee := granteeAccounts[0].address

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			granteeBalance := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			txAmount := unlockedPerLockup[0].Amount
			msg := createEthTx(nil, clawbackAccount, dest, txAmount.BigInt(), 0)
			err := validateAnteForEthTxs(msg)
			Expect(err).To(BeNil())

			// Deliver Eth Tx
			err = deliverEthTxs(granteeAccounts[0].privKey, msg)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			gb := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().GreaterOrEqual(granteeBalance.Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Uint64(), db.Amount.Uint64())
		})

		It("should enable access to unlocked EVM tokens (single-account, multiple-msgs)", func() {
			clawbackAccount := clawbackAccounts[0]
			grantee := granteeAccounts[0].address

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			granteeBalance := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			// Split the total unlocked amount into numMsgs tx's
			numMsgs := 3
			msgs := make([]sdk.Msg, numMsgs)
			txAmount := unlockedPerLockup[0].Amount.QuoRaw(int64(numMsgs))

			for i := 0; i < numMsgs; i++ {
				msgs[i] = createEthTx(nil, clawbackAccount, dest, txAmount.BigInt(), i)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Deliver Eth Tx
			err = deliverEthTxs(granteeAccounts[0].privKey, msgs...)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			gb := s.app.BankKeeper.GetBalance(s.ctx, grantee, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().GreaterOrEqual(granteeBalance.Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Uint64(), db.Amount.Uint64())
		})

		It("should enable access to unlocked EVM tokens (multi-account)", func() {
			txAmount := unlockedPerLockup[0].Amount

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			granteeBalances := make(sdk.Coins, numAccounts)
			msgs := make([]sdk.Msg, numAccounts)
			for i, grantee := range granteeAccounts {
				granteeBalances[i] = s.app.BankKeeper.GetBalance(s.ctx, grantee.address, stakeDenom)
				msgs[i] = createEthTx(grantee.privKey, grantee.clawbackAccount, dest, txAmount.BigInt(), 0)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Deliver Eth Tx
			err = deliverEthTxs(nil, msgs...)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Mul(math.NewInt(int64(numAccounts))), db.Amount)

			for i, clawbackAccount := range clawbackAccounts {
				gb := s.app.BankKeeper.GetBalance(s.ctx, clawbackAccount.GetAddress(), stakeDenom)
				s.Require().GreaterOrEqual(granteeBalances[i].Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			}
		})

		It("should enable access to unlocked EVM tokens (multi-account, multiple-msgs)", func() {
			numMsgs := 3
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockup[0].Amount.QuoRaw(int64(numMsgs))

			funderBalance := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			destBalance := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			granteeBalances := make(sdk.Coins, numAccounts)
			for i, grantee := range granteeAccounts {
				granteeBalances[i] = s.app.BankKeeper.GetBalance(s.ctx, grantee.address, stakeDenom)
				for j := 0; j < numMsgs; j++ {
					msgs = append(msgs, createEthTx(grantee.privKey, grantee.clawbackAccount, dest, txAmount.BigInt(), j))
				}
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Deliver Eth Tx
			err = deliverEthTxs(nil, msgs...)
			Expect(err).To(BeNil())

			fb := s.app.BankKeeper.GetBalance(s.ctx, funder, stakeDenom)
			db := s.app.BankKeeper.GetBalance(s.ctx, dest, stakeDenom)

			s.Require().Equal(funderBalance, fb)
			s.Require().Equal(destBalance.Add(unlockedPerLockup[0]).Amount.Mul(math.NewInt(int64(numAccounts))), db.Amount)

			for i, clawbackAccount := range clawbackAccounts {
				gb := s.app.BankKeeper.GetBalance(s.ctx, clawbackAccount.GetAddress(), stakeDenom)
				s.Require().GreaterOrEqual(granteeBalances[i].Sub(unlockedPerLockup[0]).Amount.Uint64(), gb.Amount.Uint64())
			}
		})

		It("should not enable access to locked EVM tokens (single-account)", func() {
			clawbackAccount := clawbackAccounts[0]

			// Run Tx spending entire balance
			txAmount := vestingAmtTotal[0].Amount
			msg := createEthTx(nil, clawbackAccount, dest, txAmount.BigInt(), 0)
			err := validateAnteForEthTxs(msg)
			Expect(err).To(BeNil())

			// Delivery Fails
			err = deliverEthTxs(nil, msg)
			Expect(err).ToNot(BeNil())
		})

		It("should not enable access to locked EVM tokens (single-account, multiple-msgs)", func() {
			clawbackAccount := clawbackAccounts[0]
			numMsgs := 3
			msgs := make([]sdk.Msg, numMsgs+1)
			txAmount := unlockedPerLockup[0].Amount.QuoRaw(int64(numMsgs))

			// Add message that exceeds unlocked balance
			for i := 0; i < numMsgs+1; i++ {
				msgs[i] = createEthTx(nil, clawbackAccount, dest, txAmount.BigInt(), i)
			}

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Delivery Fails
			err = deliverEthTxs(nil, msgs...)
			Expect(err).ToNot(BeNil())
		})

		It("should not enable access to locked EVM tokens (multi-account)", func() {
			msgs := make([]sdk.Msg, numAccounts+1)
			txAmount := unlockedPerLockup[0].Amount

			for i, grantee := range granteeAccounts {
				msgs[i] = createEthTx(grantee.privKey, grantee.clawbackAccount, dest, txAmount.BigInt(), 0)
			}

			// Add message that exceeds unlocked balance
			msgs[numAccounts] = createEthTx(nil, clawbackAccounts[0], dest, txAmount.BigInt(), 1)

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Delivery Fails
			err = deliverEthTxs(nil, msgs...)
			Expect(err).ToNot(BeNil())
		})

		It("should enable access to unlocked EVM tokens (multi-account, multiple-msgs)", func() {
			numMsgs := 3
			msgs := []sdk.Msg{}
			txAmount := unlockedPerLockup[0].Amount.QuoRaw(int64(numMsgs))

			for _, grantee := range granteeAccounts {
				for j := 0; j < numMsgs; j++ {
					msgs = append(msgs, createEthTx(grantee.privKey, grantee.clawbackAccount, dest, txAmount.BigInt(), j))
				}
			}

			// Add message that exceeds unlocked balance
			msgs = append(msgs, createEthTx(nil, clawbackAccounts[0], dest, txAmount.BigInt(), numMsgs))

			err := validateAnteForEthTxs(msgs...)
			Expect(err).To(BeNil())

			// Delivery Fails
			err = deliverEthTxs(nil, msgs...)
			Expect(err).ToNot(BeNil())
		})
	})
})

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func delegate(clawbackAccount *types.ClawbackVestingAccount, amount int64) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)
	//
	val, err := sdk.ValAddressFromBech32("evmosvaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn")
	s.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(claimstypes.DefaultParams().ClaimsDenom, sdk.NewInt(amount)))
	err = txBuilder.SetMsgs(delegateMsg)
	s.Require().NoError(err)
	tx := txBuilder.GetTx()

	dec := ante.NewVestingDelegationDecorator(s.app.AccountKeeper, s.app.StakingKeeper, types.ModuleCdc)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

func createEthTx(privKey *ethsecp256k1.PrivKey, clawbackAccount *types.ClawbackVestingAccount, dest sdk.AccAddress, amount *big.Int, nonceIncrement int) *evmtypes.MsgEthereumTx {
	toAddr := common.BytesToAddress(dest.Bytes())
	fromAddr := common.BytesToAddress(clawbackAccount.GetAddress().Bytes())
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
	dec := ante.NewEthVestingTransactionDecorator(s.app.AccountKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

func deliverEthTxs(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) error {
	_, err := testutil.DeliverEthTx(s.ctx, s.app, priv, msgs...)
	return err
}

// validateAnteForEthTx checks a simple single-message Ethereum transaction against the EVM Vesting AnteHandler
func validateAnteForEthTx(clawbackAccount *types.ClawbackVestingAccount, amount *big.Int) error {
	msg := createEthTx(nil, clawbackAccount, clawbackAccount.GetAddress(), amount, 0)

	return validateAnteForEthTxs(msg)
}
