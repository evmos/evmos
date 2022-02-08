package keeper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	incentivestypes "github.com/tharsis/evmos/x/incentives/types"
)

var _ = Describe("Integration", Ordered, func() {

	BeforeEach(func() {
		s.SetupTest()
	})

	Describe("Commiting a block", func() {
		addr := s.app.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
		before := s.app.BankKeeper.GetBalance(s.ctx, addr, denomMint)

		Context("before an epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)    // Start Epoch
				s.CommitAfter(time.Hour * 23) // End Epoch
			})

			It("should not allocate funds to usage incentives", func() {
				actual := s.app.BankKeeper.GetBalance(s.ctx, addr, denomMint)
				Expect(actual.IsZero()).To(BeTrue())
			})
		})

		Context("after an epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)   // Start Epoch
				s.CommitAfter(time.Hour * 4) // End Epoch
			})

			It("should allocate funds to usage incentives", func() {
				actual := s.app.BankKeeper.GetBalance(s.ctx, addr, denomMint)

				provision, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
				params := s.app.InflationKeeper.GetParams(s.ctx)
				distribution := params.InflationDistribution.UsageIncentives
				expected := (provision.Mul(distribution)).TruncateInt().Sub(before.Amount)

				Expect(actual.IsZero()).ToNot(BeTrue())
				Expect(actual.Amount).To(Equal(expected))
			})
		})
	})
})
