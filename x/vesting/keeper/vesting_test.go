package keeper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/tharsis/evmos/x/vesting/types"
)

var _ = Describe("Clawback Vesting Accounts", Ordered, func() {
	addr := sdk.AccAddress(s.address.Bytes())

	// Periodic vesting case In this case the cliff is reached before the locked
	// period is reached to represent the scenario in which an employee starts
	// before mainnet launch (periodsCliff < lockupPeriod)
	//
	// Example:
	// 21/10 Employee joins Evmos and vesting starts
	// 22/03 Mainnet launch
	// 22/09 Cliff ends
	// 23/02 Lock ends

	// Monthly vesting period
	stakeDenom := stakingtypes.DefaultParams().BondDenom
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
		vesting         sdk.Coins
		vested          sdk.Coins
	)

	BeforeEach(func() {
		s.SetupTest()

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime().Unix()
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
		err := testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, vestingAmtTotal)
		s.Require().NoError(err)
		s.app.AccountKeeper.SetAccount(s.ctx, clawbackAccount)

		// Check if all tokens are vesting at vestingStart
		vesting = s.app.BankKeeper.LockedCoins(s.ctx, addr)
		vested = s.app.BankKeeper.SpendableCoins(s.ctx, addr)
		s.Require().Equal(vestingAmtTotal, vesting)
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
			lockingDuration := time.Duration(lockupLength)
			s.CommitAfter(lockingDuration * time.Second)

			// Check if some, but not all tokens are vested
			vested = s.app.BankKeeper.SpendableCoins(s.ctx, addr)
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)
		})

		It("cannot delegate vesting tokens", func() {
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				addr,
				vestingAmtTotal.AmountOf(stakeDenom),
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
				vestingAmtTotal,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can stake vested tokens", func() {
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				clawbackAccount.GetAddress(),
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
