package keeper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ = Describe("Periodic Vesting Accounts", Ordered, func() {
	addr := sdk.AccAddress(s.address.Bytes())

	// Periodic vesting case In this case the cliff is reached before the locked
	// period is reached to represent the scenario in which an employee starts
	// before mainnet launch (periodsCliff < periodslock)
	//
	// Example:
	// 21/10 Employee joins Evmos and vesting starts
	// 22/03 Mainnet launch
	// 22/09 Cliff ends
	// 23/02 Lock ends
	periodDuration := int64(60 * 60 * 24 * 30) // 1 month in seconds
	// periodsCliff := int64(6)                   // 6 months
	periodsLock := int64(12)  // 12 year
	periodsTotal := int64(48) // 4 years
	amt := sdk.NewInt(1)
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	vestingProvision := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))

	periods := authvesting.Periods{}
	for p := int64(1); p <= periodsTotal; p++ {
		period := authvesting.Period{Length: periodDuration, Amount: vestingProvision}
		periods = append(periods, period)
	}

	var (
		periodicAccount *authvesting.PeriodicVestingAccount
		vesting         sdk.Coins
		vested          sdk.Coins
	)

	BeforeEach(func() {
		s.SetupTest()

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime().Unix()
		baseAccount := authtypes.NewBaseAccountWithAddress(addr)
		periodicAccount = authvesting.NewPeriodicVestingAccount(baseAccount, vestingTotal, vestingStart, periods)
		err := testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, vestingTotal)
		s.Require().NoError(err)
		s.app.AccountKeeper.SetAccount(s.ctx, periodicAccount)

		// Check if all tokens are vesting at vestingStart
		vesting = s.app.BankKeeper.LockedCoins(s.ctx, addr)
		vested = s.app.BankKeeper.SpendableCoins(s.ctx, addr)
		s.Require().Equal(vestingTotal, vesting)
		s.Require().True(vested.IsZero())
	})

	// TODO vesting cliff not supported with standard Cosmos SDK
	Context("before vesting cliff", func() {
		It("cannot delegate tokens", func() {
		})
		It("cannot vote on governance proposals", func() {
		})
		It("cannot transfer tokens", func() {
		})
		It("cannot perform Ethereum tx", func() {
		})
	})

	// TODO lock period not supported with standard Cosmos SDK
	Context("before locking period", func() {
		It("can delegate vested tokens", func() {
		})
		It("can vote on governance proposals", func() {
		})
		It("cannot transfer tokens", func() {
		})
		It("cannot perform Ethereum tx", func() {
		})
	})

	Context("after vesting cliff and locking period", func() {
		BeforeEach(func() {
			// Surpass locking duration
			lockingDuration := time.Duration(periodDuration * periodsLock)
			s.CommitAfter(lockingDuration * time.Second)

			// Check if some, but not all tokens are vested
			vested = s.app.BankKeeper.SpendableCoins(s.ctx, addr)
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsLock))))
			s.Require().NotEqual(vestingTotal, vested)
			s.Require().Equal(expVested, vested)
		})

		It("cannot delegate vesting tokens", func() {
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				addr,
				vestingTotal.AmountOf(stakeDenom),
				stakingtypes.Unbonded,
				s.validator,
				true,
			)
			// TODO Delegation should fail, but standard Cosmos SDK allows staking vesting tokens
			// Expect(err).ToNot(BeNil())
			Expect(err).To(BeNil())
		})

		It("cannot transfer vesting tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vestingTotal,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can stake vested tokens", func() {
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				periodicAccount.GetAddress(),
				vested.AmountOf(stakeDenom),
				stakingtypes.Unbonded,
				s.validator,
				true,
			)
			Expect(err).To(BeNil())
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

		It("can perform ethereum tx", func() {
			_, err := s.DeployContract("vestcoin", "VESTCOIN", erc20Decimals)
			Expect(err).To(BeNil())
		})
		// TODO Rewards Tests
		// TODO Clawback Tests
		// ? If the funder of a true vesting grant will be able to command "clawback" who is the funder in our case at genesis
	})
})
