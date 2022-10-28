package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	epochstypes "github.com/evmos/evmos/v9/x/epochs/types"
	incentivestypes "github.com/evmos/evmos/v9/x/incentives/types"
)

var (
	epochNumber int64
	skipped     uint64
	provision   sdk.Dec
	found       bool
)

var _ = Describe("Inflation", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()
	})

	Describe("Commiting a block", func() {
		addr := s.app.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)

		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = true
				s.app.InflationKeeper.SetParams(s.ctx, params)
			})

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

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = false
				s.app.InflationKeeper.SetParams(s.ctx, params)
			})

			Context("after the network was offline for several days/epochs", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)        // start initial epoch
					s.CommitAfter(time.Hour * 24 * 5) // end epoch after several days
				})
				When("the epoch start time has not caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 3 blocks to trigger afterEpochEnd let EpochStartTime
						// catch up with BlockTime
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 6) // commit next block
					})
					It("should increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber + 1))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped + 1))
					})
				})

				When("the epoch start time has caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 4 blocks to trigger afterEpochEnd hook several times
						// and let EpochStartTime catch up with BlockTime
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 6) // commit next block
					})
					It("should not increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped))
					})

					When("epoch number passes epochsPerPeriod + skippedEpochs and inflation re-enabled", func() {
						BeforeEach(func() {
							params := s.app.InflationKeeper.GetParams(s.ctx)
							params.EnableInflation = true
							s.app.InflationKeeper.SetParams(s.ctx, params)

							epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
							epochNumber := epochInfo.CurrentEpoch // 6

							epochsPerPeriod := int64(1)
							s.app.InflationKeeper.SetEpochsPerPeriod(s.ctx, epochsPerPeriod)
							skipped := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
							s.Require().Equal(epochNumber, epochsPerPeriod+int64(skipped))

							provision, found = s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().True(found)
							fmt.Println(provision)

							s.CommitAfter(time.Hour * 23) // commit before next full epoch
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().Equal(provisionAfter, provision)

							s.CommitAfter(time.Hour * 2) // commit after next full epoch
						})

						It("should recalculate the EpochMintProvision", func() {
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							Expect(provisionAfter).ToNot(Equal(provision))
							Expect(provisionAfter).To(Equal(sdk.MustNewDecFromStr("159375000000000000000000000")))
						})
					})
				})
			})
		})
	})
})
