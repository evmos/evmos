package keeper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	incentivestypes "github.com/tharsis/evmos/v2/x/incentives/types"
)

var _ = Describe("Integration", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()
	})

	Describe("Commiting a block", func() {
		addr := s.app.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)

		Context("before an epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)    // Start Epoch
				s.CommitAfter(time.Hour * 23) // End Epoch
			})

			It("should not allocate funds to usage incentives", func() {
				balance := s.app.BankKeeper.GetBalance(s.ctx, addr, denomMint)
				Expect(balance.IsZero()).To(BeTrue())
			})
			It("should not allocate funds to the community pool", func() {
				balance := s.app.DistrKeeper.GetFeePoolCommunityCoins(s.ctx)
				Expect(balance.IsZero()).To(BeTrue())
			})
		})

		Context("after an epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)    // Start Epoch
				s.CommitAfter(time.Hour * 25) // End Epoch
			})

			It("should allocate funds to usage incentives", func() {
				actual := s.app.BankKeeper.GetBalance(s.ctx, addr, denomMint)

				provision, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
				params := s.app.InflationKeeper.GetParams(s.ctx)
				distribution := params.InflationDistribution.UsageIncentives
				expected := (provision.Mul(distribution)).TruncateInt()

				Expect(actual.IsZero()).ToNot(BeTrue())
				Expect(actual.Amount).To(Equal(expected))
			})
			It("should allocate funds to the community pool", func() {
				balanceCommunityPool := s.app.DistrKeeper.GetFeePoolCommunityCoins(s.ctx)

				provision, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
				params := s.app.InflationKeeper.GetParams(s.ctx)
				distribution := params.InflationDistribution.CommunityPool
				expected := provision.Mul(distribution)

				Expect(balanceCommunityPool.IsZero()).ToNot(BeTrue())
				Expect(balanceCommunityPool.AmountOf(denomMint).GT(expected)).To(BeTrue())
			})
		})
	})
})
