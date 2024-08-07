package keeper_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/evmos/evmos/v19/testutil/integration/common/factory"
	testutils "github.com/evmos/evmos/v19/testutil/integration/evmos/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	fmkttypes "github.com/evmos/evmos/v19/x/feemarket/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

type txParams struct {
	gasPrice  *big.Int
	gasFeeCap *big.Int
	gasTipCap *big.Int
	accesses  *ethtypes.AccessList
}
type getprices func() txParams

func TestKeeperIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

var _ = Describe("Feemarket", func() {
	var (
		s       *KeeperTestSuite
		privKey cryptotypes.PrivKey
	)

	BeforeEach(func() {
		s = new(KeeperTestSuite)
		s.SetupTest()
		privKey = s.keyring.GetPrivKey(0)
	})

	Describe("Performing Cosmos transactions", func() {
		var (
			txArgs    factory.CosmosTxArgs
			gasWanted uint64 = 200_000
		)

		BeforeEach(func() {
			msg := banktypes.MsgSend{
				FromAddress: s.keyring.GetAccAddr(0).String(),
				ToAddress:   s.keyring.GetAccAddr(1).String(),
				Amount: sdk.Coins{sdk.Coin{
					Denom:  s.denom,
					Amount: math.NewInt(10000),
				}},
			}
			txArgs = factory.CosmosTxArgs{
				ChainID: s.network.GetChainID(),
				Msgs:    []sdk.Msg{&msg},
				Gas:     &gasWanted,
			}
		})

		Context("with min-gas-prices (local) < MinGasPrices (feemarket param)", func() {
			// minGasPrices is the feemarket MinGasPrices
			const minGasPrices int64 = 15

			BeforeEach(func() {
				// local min-gas-prices is 10aevmos
				params := fmkttypes.DefaultParams()
				params.MinGasPrice = math.LegacyNewDec(minGasPrices)
				params.BaseFee = math.ZeroInt()
				err := testutils.UpdateFeeMarketParams(
					testutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      privKey,
						Params:  params,
					},
				)
				Expect(err).To(BeNil())
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices - 3)
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices)
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeTrue(), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices - 2)
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices)
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})

		Context("with min-gas-prices (local) == MinGasPrices (feemarket param)", func() {
			// minGasPrices is the feemarket MinGasPrices
			const minGasPrices int64 = 10
			BeforeEach(func() {
				// local min-gas-prices is 10aevmos
				params := fmkttypes.DefaultParams()
				params.MinGasPrice = math.LegacyNewDec(minGasPrices)
				params.BaseFee = math.ZeroInt()

				err := testutils.UpdateFeeMarketParams(
					testutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      privKey,
						Params:  params,
					},
				)
				Expect(err).To(BeNil())
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < min-gas-prices", func() {
					gasPrice := math.NewInt(minGasPrices - 3)
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("insufficient fee"))
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices)
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices - 2)
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices)
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})

		Context("with MinGasPrices (feemarket param) < min-gas-prices (local)", func() {
			// minGasPrices is the feemarket MinGasPrices
			const minGasPrices int64 = 7
			baseFee := math.NewInt(15)

			BeforeEach(func() {
				// local min-gas-prices is 10aevmos
				params := fmkttypes.DefaultParams()
				params.MinGasPrice = math.LegacyNewDec(minGasPrices)
				params.BaseFee = baseFee

				err := testutils.UpdateFeeMarketParams(
					testutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      privKey,
						Params:  params,
					},
				)
				Expect(err).To(BeNil())
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices - 3)
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("insufficient fee"))
				})

				It("should reject transactions with MinGasPrices < gasPrice < baseFee", func() {
					gasPrice := math.NewInt(minGasPrices + 1)
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("insufficient fee"))
				})

				It("should accept transactions with gasPrice >= baseFee", func() {
					gasPrice := baseFee
					txArgs.GasPrice = &gasPrice
					tx, err := s.factory.BuildCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					bz, err := s.factory.EncodeTx(tx)
					Expect(err).To(BeNil())

					res, err := s.network.CheckTx(bz)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := math.NewInt(minGasPrices - 2)
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
				})

				It("should reject transactions with MinGasPrices < gasPrice < baseFee", func() {
					gasPrice := math.NewInt(minGasPrices + 1)
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeFalse())
					Expect(res.Log).To(ContainSubstring("insufficient fee"))
				})
				It("should accept transactions with gasPrice >= baseFee", func() {
					gasPrice := baseFee
					txArgs.GasPrice = &gasPrice
					res, err := s.factory.ExecuteCosmosTx(privKey, txArgs)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})
	})

	Describe("Performing EVM transactions", func() {
		var (
			txArgs    evmtypes.EvmTxArgs
			gasWanted uint64 = 200_000
		)

		BeforeEach(func() {
			toAddr := s.keyring.GetAddr(1)
			txArgs = evmtypes.EvmTxArgs{
				ChainID:  s.network.GetEIP155ChainID(),
				GasLimit: gasWanted,
				To:       &toAddr,
				Amount:   big.NewInt(10000),
			}
		})

		Context("with MinGasPrices (feemarket param) < BaseFee (feemarket)", func() {
			var (
				baseFee      int64
				minGasPrices int64
			)

			BeforeEach(func() {
				baseFee = 10_000_000_000
				minGasPrices = baseFee - 5_000_000_000

				params := fmkttypes.DefaultParams()
				params.MinGasPrice = math.LegacyNewDec(minGasPrices)
				params.BaseFee = math.NewInt(baseFee)

				// Note that the tests run the same transactions with `gasLimit =
				// 200_000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
				// a `minGasPrices = 5_000_000_000` results in `minGlobalFee =
				// 1_000_000_000_000_000`
				err := testutils.UpdateFeeMarketParams(
					testutils.UpdateParamsInput{
						Tf:      s.factory,
						Network: s.network,
						Pk:      privKey,
						Params:  params,
					},
				)
				Expect(err).To(BeNil())
			})

			Context("during CheckTx", func() {
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						tx, err := s.factory.GenerateSignedEthTx(privKey, txArgs)
						Expect(err).To(BeNil())

						Expect(err).To(BeNil())
						bz, err := s.factory.EncodeTx(tx)
						Expect(err).To(BeNil())

						res, err := s.network.CheckTx(bz)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(BeFalse())
						Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(minGasPrices - 1_000_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices, no gasTipCap", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices - 1_000_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices, max gasTipCap", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices - 1_000_000_000), big.NewInt(minGasPrices - 1_000_000_000), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should reject transactions with MinGasPrices < tx gasPrice < EffectivePrice",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						tx, err := s.factory.GenerateSignedEthTx(privKey, txArgs)
						Expect(err).To(BeNil())

						Expect(err).To(BeNil())
						bz, err := s.factory.EncodeTx(tx)
						Expect(err).To(BeNil())

						res, err := s.network.CheckTx(bz)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(BeFalse())
						Expect(res.Log).To(ContainSubstring("insufficient fee"))
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee - 2_000_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee - 2_000_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= EffectivePrice",
					func(malleate getprices) {
						p := malleate()
						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						tx, err := s.factory.GenerateSignedEthTx(privKey, txArgs)
						Expect(err).To(BeNil())

						Expect(err).To(BeNil())
						bz, err := s.factory.EncodeTx(tx)
						Expect(err).To(BeNil())

						res, err := s.network.CheckTx(bz)
						Expect(err).To(BeNil(), "transaction should have succeeded")
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)
			})

			Context("during DeliverTx", func() {
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						res, err := s.factory.ExecuteEthTx(privKey, txArgs)
						Expect(err).NotTo(BeNil())
						Expect(res.IsOK()).To(BeFalse())
						Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(minGasPrices - 1_000_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices - 1_000_000_000), nil, &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should reject transactions with MinGasPrices < gasPrice < EffectivePrice",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						res, err := s.factory.ExecuteEthTx(privKey, txArgs)
						Expect(err).NotTo(BeNil())
						Expect(res.IsOK()).To(BeFalse())
						Expect(res.Log).To(ContainSubstring("insufficient fee"))
					},
					// Note that the baseFee is not 10_000_000_000 anymore but updates to 7_656_250_000 because of the s.Commit
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee - 2_500_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee - 2_500_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= EffectivePrice",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						res, err := s.factory.ExecuteEthTx(privKey, txArgs)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)
			})
		})

		Context("with BaseFee (feemarket) < MinGasPrices (feemarket param)", func() {
			var (
				baseFee      int64
				minGasPrices int64
			)

			Context("during CheckTx", func() {
				BeforeEach(func() {
					baseFee = 10_000_000_000
					minGasPrices = baseFee + 30_000_000_000

					// Note that the tests run the same transactions with `gasLimit =
					// 200000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
					// with `minGasPrices = 40_000_000_000` results in `minGlobalFee =
					// 8000000000000000`
					// local min-gas-prices is 10aevmos
					params := fmkttypes.DefaultParams()
					params.MinGasPrice = math.LegacyNewDec(minGasPrices)
					params.BaseFee = math.NewInt(baseFee)

					// Note that the tests run the same transactions with `gasLimit =
					// 200_000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
					// a `minGasPrices = 5_000_000_000` results in `minGlobalFee =
					// 1_000_000_000_000_000`
					err := testutils.UpdateFeeMarketParams(
						testutils.UpdateParamsInput{
							Tf:      s.factory,
							Network: s.network,
							Pk:      privKey,
							Params:  params,
						},
					)
					Expect(err).To(BeNil())
				})

				DescribeTable("should reject transactions with EffectivePrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						tx, err := s.factory.GenerateSignedEthTx(privKey, txArgs)
						Expect(err).To(BeNil())

						Expect(err).To(BeNil())
						bz, err := s.factory.EncodeTx(tx)
						Expect(err).To(BeNil())

						res, err := s.network.CheckTx(bz)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(BeFalse())
						Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(minGasPrices - 10_000_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices, no gasTipCap", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices - 10_000_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices, max gasTipCap", func() txParams {
						// Note that max priority fee per gas can't be higher than the max fee per gas (gasFeeCap), i.e. 30_000_000_000)
						return txParams{nil, big.NewInt(minGasPrices - 10_000_000_000), big.NewInt(30_000_000_000), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						tx, err := s.factory.GenerateSignedEthTx(privKey, txArgs)
						Expect(err).To(BeNil())

						Expect(err).To(BeNil())
						bz, err := s.factory.EncodeTx(tx)
						Expect(err).To(BeNil())

						res, err := s.network.CheckTx(bz)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(BeTrue(), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(minGasPrices), nil, nil, nil}
					}),
					// Note that this tx is not rejected on CheckTx, but not on DeliverTx,
					// as the baseFee is set to minGasPrices during DeliverTx when baseFee
					// < minGasPrices
					Entry("dynamic tx with GasFeeCap > MinGasPrices, EffectivePrice > MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices), big.NewInt(30_000_000_000), &ethtypes.AccessList{}}
					}),
				)
			})

			Context("during DeliverTx", func() {
				BeforeEach(func() {
					baseFee = 10_000_000_000
					minGasPrices = baseFee + 30_000_000_000

					// Note that the tests run the same transactions with `gasLimit =
					// 200000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
					// with `minGasPrices = 40_000_000_000` results in `minGlobalFee =
					// 8000000000000000`
					// local min-gas-prices is 10aevmos
					params := fmkttypes.DefaultParams()
					params.MinGasPrice = math.LegacyNewDec(minGasPrices)
					params.BaseFee = math.NewInt(baseFee)

					err := testutils.UpdateFeeMarketParams(
						testutils.UpdateParamsInput{
							Tf:      s.factory,
							Network: s.network,
							Pk:      privKey,
							Params:  params,
						},
					)
					Expect(err).To(BeNil())
				})
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						res, err := s.factory.ExecuteEthTx(privKey, txArgs)
						Expect(err).NotTo(BeNil())
						Expect(res.IsOK()).To(BeFalse())
						Expect(res.Log).To(ContainSubstring("provided fee < minimum global fee"))
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(minGasPrices - 10_000_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices, no gasTipCap", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices - 10_000_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices, max gasTipCap", func() txParams {
						// Note that max priority fee per gas can't be higher than the max fee per gas (gasFeeCap), i.e. 30_000_000_000)
						return txParams{nil, big.NewInt(minGasPrices - 10_000_000_000), big.NewInt(30_000_000_000), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
					func(malleate getprices) {
						p := malleate()

						txArgs.GasPrice = p.gasPrice
						txArgs.GasFeeCap = p.gasFeeCap
						txArgs.GasTipCap = p.gasTipCap
						txArgs.Accesses = p.accesses

						res, err := s.factory.ExecuteEthTx(privKey, txArgs)
						Expect(err).To(BeNil(), "transaction should have succeeded")
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(minGasPrices + 1), nil, nil, nil}
					}),
					Entry("dynamic tx, EffectivePrice > MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices + 10_000_000_000), big.NewInt(30_000_000_000), &ethtypes.AccessList{}}
					}),
				)
			})
		})
	})
})
