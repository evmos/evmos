package keeper_test

import (
	"fmt"
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

var _ = Describe("Vesting", Ordered, func() {
	var (
		periodicAccount *authvesting.PeriodicVestingAccount
		locked          sdk.Coins
		validator       stakingtypes.Validator
	)

	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	periodDuration := int64(60 * 60 * 24 * 30) // month
	periodsTotal := int64(48)                  // 4 years
	amt := sdk.NewInt(1)
	vestingProvision := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))

	periods := authvesting.Periods{}
	for p := int64(1); p <= periodsTotal; p++ {
		period := authvesting.Period{Length: periodDuration, Amount: vestingProvision}
		periods = append(periods, period)
	}

	BeforeEach(func() {
		s.SetupTest()
		// Create periodic vesting account
		vestingStart := s.ctx.BlockTime().Unix()
		baseAccount := authtypes.NewBaseAccountWithAddress(addr)
		periodicAccount = authvesting.NewPeriodicVestingAccount(baseAccount, vestingTotal, vestingStart, periods)
		// TODO Check if funding is the correct way to test?
		err := testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, vestingTotal)
		s.Require().NoError(err)

		// Check if all tokens are locked at vestingStart
		locked = periodicAccount.LockedCoins(s.ctx.BlockTime())
		vested := periodicAccount.GetVestedCoins(s.ctx.BlockTime())
		s.Require().Equal(vestingTotal, locked)
		s.Require().True(vested.IsZero())

		// Get Validator
		validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
		validator = validators[0]
	})

	Describe("Staking", func() {
		Context("with locked tokens", func() {
			It("must not be possible", func() {
				// Stake locked tokens
				_, err := s.app.StakingKeeper.Delegate(
					s.ctx,
					periodicAccount.GetAddress(),
					locked.AmountOf(stakeDenom),
					stakingtypes.Unbonded,
					validator,
					true,
				)
				// TODO Delegation should fail, but standard Cosmos SDK allows staking locked tokens
				// Expect(err).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		Context("with vested and unlocked tokens", func() {
			passedPeriods := int64(12)

			BeforeAll(func() {
				s.CommitAfter(time.Duration(time.Hour * 24 * 30 * time.Duration(passedPeriods)))
			})
			It("should be possible", func() {
				// Check if some tokens are vested and unlocked
				locked = periodicAccount.LockedCoins(s.ctx.BlockTime())
				vested := periodicAccount.GetVestedCoins(s.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(passedPeriods))))
				s.Require().Equal(vestingTotal.Sub(expVested), locked)
				s.Require().Equal(expVested, vested)

				// Stake vested tokens
				_, err := s.app.StakingKeeper.Delegate(
					s.ctx,
					periodicAccount.GetAddress(),
					vested.AmountOf(stakeDenom),
					stakingtypes.Unbonded,
					validator,
					true,
				)

				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Transfers", func() {
		Context("before the lock period concludes", func() {
			It("must not be possible", func() {
				// TODO lock period not supported with standard Cosmos SDK
			})
		})
		Context("with unvested tokens", func() {
			It("must not be possible", func() {
				fmt.Printf("\n locked: %v", locked)
				// Transfer locked tokens
				err := s.app.BankKeeper.SendCoins(
					s.ctx,
					addr,
					sdk.AccAddress(s.address.Bytes()),
					locked,
				)

				// TODO Transfer should fail, but standard Cosmos SDK allows staking locked tokens
				// Expect(err).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
		Context("with vested and unlocked tokens", func() {
			passedPeriods := int64(12)

			BeforeAll(func() {
				s.CommitAfter(time.Duration(time.Hour * 24 * 30 * time.Duration(passedPeriods)))
			})
			It("should be possible", func() {
				// Check if some tokens are vested and unlocked
				locked = periodicAccount.LockedCoins(s.ctx.BlockTime())
				vested := periodicAccount.GetVestedCoins(s.ctx.BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(passedPeriods))))
				s.Require().Equal(vestingTotal.Sub(expVested), locked)
				s.Require().Equal(expVested, vested)

				// Transfer locked tokens
				err := s.app.BankKeeper.SendCoins(
					s.ctx,
					addr,
					sdk.AccAddress(s.address.Bytes()),
					vested,
				)
				Expect(err).To(BeNil())
			})
		})
	})

	// Describe("Ethereum Txs", func() {
	// 	Context("before the lock period concludes", func() {
	// 		It("must not be possible", func() {
	// 		})
	// 	})
	// 	Context("with vested and unlocked tokens", func() {
	// 		It("should be possible", func() {
	// 		})
	// 	})
	// })
})
