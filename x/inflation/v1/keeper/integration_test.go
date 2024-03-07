package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	epochstypes "github.com/evmos/evmos/v16/x/epochs/types"
	inflationkeeper "github.com/evmos/evmos/v16/x/inflation/v1/keeper"
	"github.com/evmos/evmos/v16/x/inflation/v1/types"
)

var (
	epochNumber int64
	skipped     uint64
	provision   math.LegacyDec
)

func TestKeeperIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

var _ = Describe("Inflation", Ordered, func() {
	var s *KeeperTestSuite
	BeforeEach(func() {
		s = new(KeeperTestSuite)
		s.SetupTest()
	})

	Describe("Committing a block", func() {
		addr := s.network.App.AccountKeeper.GetModuleAddress("incentives")

		Context("with inflation param enabled and exponential calculation params changed", func() {
			BeforeEach(func() {
				params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
				params.EnableInflation = true
				params.ExponentialCalculation = types.ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(60, 2), // 60%
					C:             math.LegacyNewDec(int64(6_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(66, 2), // 66%
					MaxVariance:   math.LegacyZeroDec(),             // 0%
				}
				params.InflationDistribution = types.DefaultInflationDistribution
				err := s.network.App.InflationKeeper.SetParams(s.network.GetContext(), params)
				Expect(err).ToNot(HaveOccurred(), "error while setting params")
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)    // Start Epoch
					s.network.NextBlockAfter(time.Hour * 23) // End Epoch
				})

				It("should not allocate funds to usage incentives (Deprecated)", func() {
					balance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), addr, denomMint)
					Expect(balance.IsZero()).To(BeTrue(), "balance should be zero")
				})
				It("should not allocate funds to the community pool", func() {
					pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
					Expect(err).To(BeNil())
					balance := pool.CommunityPool
					Expect(balance.IsZero()).To(BeTrue())
				})
			})

			Context("after an epoch ends", func() { //nolint:dupl // these tests are not duplicates
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)    // Start Epoch
					s.network.NextBlockAfter(time.Hour * 23) // End Epoch
				})

				It("should not allocate funds to usage incentives (deprecated)", func() {
					actual := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), addr, denomMint)

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.UsageIncentives //nolint:staticcheck
					expected := (provision.Mul(distribution)).TruncateInt()

					Expect(actual.IsZero()).To(BeTrue())
					Expect(actual.Amount).To(Equal(expected))
				})

				It("should allocate funds to the community pool", func() {
					pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
					Expect(err).To(BeNil())
					balanceCommunityPool := pool.CommunityPool

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.CommunityPool
					expected := provision.Mul(distribution)

					Expect(balanceCommunityPool.IsZero()).ToNot(BeTrue())
					Expect(balanceCommunityPool.AmountOf(denomMint).GT(expected)).To(BeTrue())
				})
			})
		})

		Context("with inflation param enabled and distribution params changed", func() {
			BeforeEach(func() {
				params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
				params.EnableInflation = true
				params.InflationDistribution = types.InflationDistribution{
					StakingRewards:  math.LegacyNewDecWithPrec(333333333, 9),
					CommunityPool:   math.LegacyNewDecWithPrec(666666667, 9),
					UsageIncentives: math.LegacyZeroDec(), // Deprecated
				}
				_ = s.network.App.InflationKeeper.SetParams(s.network.GetContext(), params)
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)    // Start Epoch
					s.network.NextBlockAfter(time.Hour * 23) // End Epoch
				})

				It("should not allocate funds to usage incentives", func() {
					balance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), addr, denomMint)
					Expect(balance.IsZero()).To(BeTrue())
				})

				It("should not allocate funds to the community pool", func() {
					pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
					Expect(err).To(BeNil())
					balance := pool.CommunityPool
					Expect(balance.IsZero()).To(BeTrue())
				})
			})

			Context("after an epoch ends", func() { //nolint:dupl
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)    // Start Epoch
					s.network.NextBlockAfter(time.Hour * 25) // End Epoch
				})

				It("should not allocate funds to usage incentives (deprecated)", func() {
					actual := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), addr, denomMint)

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.UsageIncentives //nolint:staticcheck
					expected := (provision.Mul(distribution)).TruncateInt()

					Expect(actual.IsZero()).To(BeTrue())
					Expect(actual.Amount).To(Equal(expected))
				})

				It("should allocate funds to the community pool", func() {
					pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
					Expect(err).To(BeNil())
					balanceCommunityPool := pool.CommunityPool

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.CommunityPool
					expected := provision.Mul(distribution)

					Expect(balanceCommunityPool.IsZero()).ToNot(BeTrue())
					Expect(balanceCommunityPool.AmountOf(denomMint).GT(expected)).To(BeTrue())
				})
			})
		})

		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
				params.EnableInflation = true
				s.network.App.InflationKeeper.SetParams(s.network.GetContext(), params) //nolint:errcheck
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)    // Start Epoch
					s.network.NextBlockAfter(time.Hour * 23) // End Epoch
				})

				It("should not allocate funds to usage incentives", func() {
					balance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), addr, denomMint)
					Expect(balance.IsZero()).To(BeTrue())
				})
				It("should not allocate funds to the community pool", func() {
					pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
					Expect(err).To(BeNil())
					balance := pool.CommunityPool
					Expect(balance.IsZero()).To(BeTrue())
				})
			})

			Context("after an epoch ends", func() { //nolint:dupl
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)    // Start Epoch
					s.network.NextBlockAfter(time.Hour * 25) // End Epoch
				})

				It("should not allocate funds to usage incentives (deprecated)", func() {
					actual := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), addr, denomMint)

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.UsageIncentives //nolint:staticcheck
					expected := (provision.Mul(distribution)).TruncateInt()

					Expect(actual.IsZero()).To(BeTrue())
					Expect(actual.Amount).To(Equal(expected))
				})
				It("should allocate funds to the community pool", func() {
					pool, err := s.network.App.DistrKeeper.FeePool.Get(s.network.GetContext())
					Expect(err).To(BeNil())
					balanceCommunityPool := pool.CommunityPool

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.CommunityPool
					expected := provision.Mul(distribution)

					Expect(balanceCommunityPool.IsZero()).ToNot(BeTrue())
					Expect(balanceCommunityPool.AmountOf(denomMint).GT(expected)).To(BeTrue())
				})
			})
		})

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
				params.EnableInflation = false
				s.network.App.InflationKeeper.SetParams(s.network.GetContext(), params) //nolint:errcheck
			})

			Context("after the network was offline for several days/epochs", func() {
				BeforeEach(func() {
					s.network.NextBlockAfter(time.Minute)        // Start Epoch
					s.network.NextBlockAfter(time.Hour * 24 * 5) // end epoch after several days
				})
				When("the epoch start time has not caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 3 blocks to trigger afterEpochEnd let EpochStartTime
						// catch up with BlockTime
						s.network.NextBlockAfter(time.Second * 6)
						s.network.NextBlockAfter(time.Second * 6)
						s.network.NextBlockAfter(time.Second * 6)

						epochInfo, found := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())

						// commit next block
						s.network.NextBlockAfter(time.Second * 6)
					})
					It("should increase the epoch number ", func() {
						epochInfo, _ := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber + 1))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())
						Expect(skippedAfter).To(Equal(skipped + 1))
					})
				})

				When("the epoch start time has caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 4 blocks to trigger afterEpochEnd hook several times
						// and let EpochStartTime catch up with BlockTime
						s.network.NextBlockAfter(time.Second * 6)
						s.network.NextBlockAfter(time.Second * 6)
						s.network.NextBlockAfter(time.Second * 6)
						s.network.NextBlockAfter(time.Second * 6)

						epochInfo, found := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())

						// commit next block
						s.network.NextBlockAfter(time.Second * 6)
					})
					It("should not increase the epoch number ", func() {
						epochInfo, _ := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())
						Expect(skippedAfter).To(Equal(skipped))
					})

					When("epoch number passes epochsPerPeriod + skippedEpochs and inflation re-enabled", func() {
						BeforeEach(func() {
							params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
							params.EnableInflation = true
							s.network.App.InflationKeeper.SetParams(s.network.GetContext(), params) //nolint:errcheck

							epochInfo, _ := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
							epochNumber := epochInfo.CurrentEpoch // 6

							epochsPerPeriod := int64(1)
							s.network.App.InflationKeeper.SetEpochsPerPeriod(s.network.GetContext(), epochsPerPeriod)
							skipped := s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())
							s.Require().Equal(epochNumber, epochsPerPeriod+int64(skipped))

							provision = s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())

							// commit before next full epoch
							s.network.NextBlockAfter(time.Hour * 23)
							provisionAfter := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
							s.Require().Equal(provisionAfter, provision)

							// commit after next full epoch
							s.network.NextBlockAfter(time.Hour * 2)
						})

						It("should recalculate the EpochMintProvision", func() {
							provisionAfter := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
							Expect(provisionAfter).ToNot(Equal(provision))
							Expect(provisionAfter).To(Equal(math.LegacyMustNewDecFromStr("159375000000000000000000000").Quo(math.LegacyNewDec(inflationkeeper.ReductionFactor))))
						})
					})
				})
			})
		})
	})
})
