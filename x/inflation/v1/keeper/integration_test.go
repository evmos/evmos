package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	integrationutils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
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

	Context("Committing a block", func() {
		var (
			prevCommPoolBalanceAmt math.LegacyDec
			addr                   = authtypes.NewModuleAddress("incentives")
		)

		Context("with inflation param enabled and exponential calculation params changed", func() {
			BeforeEach(func() {
				params := types.DefaultParams()
				params.EnableInflation = true
				params.ExponentialCalculation = types.ExponentialCalculation{
					A:             math.LegacyNewDec(int64(300_000_000)),
					R:             math.LegacyNewDecWithPrec(60, 2), // 60%
					C:             math.LegacyNewDec(int64(6_375_000)),
					BondingTarget: math.LegacyNewDecWithPrec(66, 2), // 66%
					MaxVariance:   math.LegacyZeroDec(),             // 0%
				}
				params.InflationDistribution = types.DefaultInflationDistribution
				err := integrationutils.UpdateInflationParams(
					integrationutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      s.keyring.GetPrivKey(0),
						Params:  params,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "error while setting params")
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					prevCommPoolBalanceAmt = res.Pool.AmountOf(denomMint)

					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil())    // Start Epoch
					Expect(s.network.NextBlockAfter(time.Hour * 23)).To(BeNil()) // End Epoch
				})

				It("should not allocate funds to usage incentives (Deprecated)", func() {
					res, err := s.handler.GetBalance(addr, denomMint)
					Expect(err).To(BeNil())
					balance := res.Balance
					Expect(balance.IsZero()).To(BeTrue(), "balance should be zero")
				})
				It("should not allocate funds to the community pool", func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					finalAmt := res.Pool.AmountOf(denomMint)
					Expect(finalAmt.Sub(prevCommPoolBalanceAmt).TruncateInt64()).To(Equal(int64(0)))
				})
			})

			Context("after an epoch ends", func() { //nolint:dupl // these tests are not duplicates
				BeforeEach(func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					prevCommPoolBalanceAmt = res.Pool.AmountOf(denomMint)

					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil())    // Start Epoch
					Expect(s.network.NextBlockAfter(time.Hour * 25)).To(BeNil()) // End Epoch
				})

				It("should not allocate funds to usage incentives (deprecated)", func() {
					res, err := s.handler.GetBalance(addr, denomMint)
					Expect(err).To(BeNil())
					actual := res.Balance

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.UsageIncentives //nolint:staticcheck
					expected := (provision.Mul(distribution)).TruncateInt()

					Expect(actual.IsZero()).To(BeTrue())
					Expect(actual.Amount).To(Equal(expected))
				})

				It("should allocate funds to the community pool", func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					balanceCommunityPoolAmt := res.Pool.AmountOf(denomMint)

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.CommunityPool
					expected := provision.Mul(distribution)

					allocatedAmt := balanceCommunityPoolAmt.Sub(prevCommPoolBalanceAmt)
					Expect(allocatedAmt.IsZero()).ToNot(BeTrue())
					Expect(allocatedAmt.GT(expected)).To(BeTrue())
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
				err := integrationutils.UpdateInflationParams(
					integrationutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      s.keyring.GetPrivKey(0),
						Params:  params,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "error while setting params")
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					prevCommPoolBalanceAmt = res.Pool.AmountOf(denomMint)

					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil())    // Start Epoch
					Expect(s.network.NextBlockAfter(time.Hour * 23)).To(BeNil()) // End Epoch
				})

				It("should not allocate funds to usage incentives", func() {
					res, err := s.handler.GetBalance(addr, denomMint)
					Expect(err).To(BeNil())
					balance := res.Balance
					Expect(balance.IsZero()).To(BeTrue())
				})

				It("should not allocate funds to the community pool", func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					finalAmt := res.Pool.AmountOf(denomMint)
					Expect(finalAmt.Sub(prevCommPoolBalanceAmt).IsZero()).To(BeTrue())
				})
			})

			Context("after an epoch ends", func() { //nolint:dupl
				BeforeEach(func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					prevCommPoolBalanceAmt = res.Pool.AmountOf(denomMint)

					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil())    // Start Epoch
					Expect(s.network.NextBlockAfter(time.Hour * 25)).To(BeNil()) // End Epoch
				})

				It("should not allocate funds to usage incentives (deprecated)", func() {
					res, err := s.handler.GetBalance(addr, denomMint)
					Expect(err).To(BeNil())
					actual := res.Balance

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.UsageIncentives //nolint:staticcheck
					expected := (provision.Mul(distribution)).TruncateInt()

					Expect(actual.IsZero()).To(BeTrue())
					Expect(actual.Amount).To(Equal(expected))
				})

				It("should allocate funds to the community pool", func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					balanceCommunityPoolAmt := res.Pool.AmountOf(denomMint)

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.CommunityPool
					expected := provision.Mul(distribution)

					allocatedAmt := balanceCommunityPoolAmt.Sub(prevCommPoolBalanceAmt)
					Expect(allocatedAmt.IsZero()).ToNot(BeTrue())
					Expect(allocatedAmt.GT(expected)).To(BeTrue())
				})
			})
		})

		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
				params.EnableInflation = true
				err := integrationutils.UpdateInflationParams(
					integrationutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      s.keyring.GetPrivKey(0),
						Params:  params,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "error while setting params")
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					prevCommPoolBalanceAmt = res.Pool.AmountOf(denomMint)

					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil())    // Start Epoch
					Expect(s.network.NextBlockAfter(time.Hour * 23)).To(BeNil()) // End Epoch
				})

				It("should not allocate funds to usage incentives", func() {
					res, err := s.handler.GetBalance(addr, denomMint)
					Expect(err).To(BeNil())
					balance := res.Balance
					Expect(balance.IsZero()).To(BeTrue())
				})
				It("should not allocate funds to the community pool", func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					finalAmt := res.Pool.AmountOf(denomMint)
					Expect(finalAmt.Sub(prevCommPoolBalanceAmt).IsZero()).To(BeTrue())
				})
			})

			Context("after an epoch ends", func() { //nolint:dupl
				BeforeEach(func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					prevCommPoolBalanceAmt = res.Pool.AmountOf(denomMint)

					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil())    // Start Epoch
					Expect(s.network.NextBlockAfter(time.Hour * 25)).To(BeNil()) // End Epoch
				})

				It("should not allocate funds to usage incentives (deprecated)", func() {
					res, err := s.handler.GetBalance(addr, denomMint)
					Expect(err).To(BeNil())
					actual := res.Balance

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.UsageIncentives //nolint:staticcheck
					expected := (provision.Mul(distribution)).TruncateInt()

					Expect(actual.IsZero()).To(BeTrue())
					Expect(actual.Amount).To(Equal(expected))
				})
				It("should allocate funds to the community pool", func() {
					res, err := s.handler.GetCommunityPool()
					Expect(err).To(BeNil())
					balanceCommunityPoolAmt := res.Pool.AmountOf(denomMint)

					provision := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
					params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
					distribution := params.InflationDistribution.CommunityPool
					expected := provision.Mul(distribution)

					allocatedAmt := balanceCommunityPoolAmt.Sub(prevCommPoolBalanceAmt)
					Expect(allocatedAmt.IsZero()).ToNot(BeTrue())
					Expect(allocatedAmt.GT(expected)).To(BeTrue())
				})
			})
		})

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.network.App.InflationKeeper.GetParams(s.network.GetContext())
				params.EnableInflation = false
				err := integrationutils.UpdateInflationParams(
					integrationutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      s.keyring.GetPrivKey(0),
						Params:  params,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "error while setting params")
			})

			Context("after the network was offline for several days/epochs", func() {
				BeforeEach(func() {
					Expect(s.network.NextBlockAfter(time.Minute)).To(BeNil()) // Start Epoch
					s.network.NextBlockAfter(time.Hour * 24 * 5)              // end epoch after several days
				})
				When("the epoch start time has not caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 3 blocks to trigger afterEpochEnd let EpochStartTime
						// catch up with BlockTime
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())

						epochInfo, found := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
						Expect(found).To(BeTrue())
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())

						// commit next block
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
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
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())

						epochInfo, found := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
						Expect(found).To(BeTrue())
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())

						// commit next block
						Expect(s.network.NextBlockAfter(time.Second * 6)).To(BeNil())
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
							err := integrationutils.UpdateInflationParams(
								integrationutils.UpdateParamsInput{
									Tf:      s.factory,
									Network: s.network,
									Pk:      s.keyring.GetPrivKey(0),
									Params:  params,
								},
							)
							Expect(err).ToNot(HaveOccurred(), "error while setting params")

							skipped := s.network.App.InflationKeeper.GetSkippedEpochs(s.network.GetContext())
							Expect(skipped > uint64(0)).To(BeTrue())

							epochsPerPeriod := s.network.App.InflationKeeper.GetEpochsPerPeriod(s.network.GetContext())
							provision = s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())

							// commit before next full epoch
							Expect(s.network.NextBlockAfter(time.Hour * 23)).To(BeNil())
							provisionAfter := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
							Expect(provisionAfter).To(Equal(provision))

							// commit after next full epoch (next period)
							for i := int64(0); i < epochsPerPeriod; i++ {
								Expect(s.network.NextBlockAfter(time.Hour * 24)).To(BeNil())
							}

							epochInfo, _ := s.network.App.EpochsKeeper.GetEpochInfo(s.network.GetContext(), epochstypes.DayEpochID)
							epochNumber := epochInfo.CurrentEpoch

							Expect(epochNumber > epochsPerPeriod).To(BeTrue())
						})

						It("should recalculate the EpochMintProvision", func() {
							provisionAfter := s.network.App.InflationKeeper.GetEpochMintProvision(s.network.GetContext())
							Expect(provisionAfter).ToNot(Equal(provision))
							Expect(provisionAfter).To(Equal(math.LegacyMustNewDecFromStr("436643835616438356164384").Quo(math.LegacyNewDec(inflationkeeper.ReductionFactor))))
						})
					})
				})
			})
		})
	})
})
