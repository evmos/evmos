package keeper_test

import (
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/testutil"
	utiltx "github.com/evmos/evmos/v14/testutil/tx"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ = Describe("Feemarket", func() {
	var (
		privKey *ethsecp256k1.PrivKey
		msg     banktypes.MsgSend
	)

	Describe("Performing Cosmos transactions", func() {
		Context("with min-gas-prices (local) < MinGasPrices (feemarket param)", func() {
			BeforeEach(func() {
				privKey, msg = setupTestWithContext("1", sdk.NewDec(3), sdk.ZeroInt())
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(2)
					_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).ToNot(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), err.Error())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(3)
					res, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(2)
					_, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).NotTo(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), err.Error())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(3)
					res, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					s.Require().NoError(err)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})

		Context("with min-gas-prices (local) == MinGasPrices (feemarket param)", func() {
			BeforeEach(func() {
				privKey, msg = setupTestWithContext("3", sdk.NewDec(3), sdk.ZeroInt())
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < min-gas-prices", func() {
					gasPrice := sdkmath.NewInt(2)
					_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).ToNot(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"insufficient fee"),
					).To(BeTrue(), err.Error())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(3)
					res, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(2)
					_, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).NotTo(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), err.Error())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(3)
					res, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})

		Context("with MinGasPrices (feemarket param) < min-gas-prices (local)", func() {
			BeforeEach(func() {
				privKey, msg = setupTestWithContext("5", sdk.NewDec(3), sdk.NewInt(5))
			})

			//nolint
			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(2)
					_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).ToNot(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"insufficient fee"),
					).To(BeTrue(), err.Error())
				})

				It("should reject transactions with MinGasPrices < gasPrice < baseFee", func() {
					gasPrice := sdkmath.NewInt(4)
					_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).ToNot(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"insufficient fee"),
					).To(BeTrue(), err.Error())
				})

				It("should accept transactions with gasPrice >= baseFee", func() {
					gasPrice := sdkmath.NewInt(5)
					res, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			//nolint
			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdkmath.NewInt(2)
					_, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).NotTo(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), err.Error())
				})

				It("should reject transactions with MinGasPrices < gasPrice < baseFee", func() {
					gasPrice := sdkmath.NewInt(4)
					_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).ToNot(BeNil(), "transaction should have failed")
					Expect(
						strings.Contains(err.Error(),
							"insufficient fee"),
					).To(BeTrue(), err.Error())
				})
				It("should accept transactions with gasPrice >= baseFee", func() {
					gasPrice := sdkmath.NewInt(5)
					res, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})
	})

	Describe("Performing EVM transactions", func() {
		type txParams struct {
			gasPrice  *big.Int
			gasFeeCap *big.Int
			gasTipCap *big.Int
			accesses  *ethtypes.AccessList
		}
		type getprices func() txParams

		Context("with BaseFee (feemarket) < MinGasPrices (feemarket param)", func() {
			var (
				baseFee      int64
				minGasPrices int64
			)

			BeforeEach(func() {
				baseFee = 10_000_000_000
				minGasPrices = baseFee + 30_000_000_000

				// Note that the tests run the same transactions with `gasLimit =
				// 100000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
				// a `minGasPrices = 40_000_000_000` results in `minGlobalFee =
				// 4000000000000000`
				privKey, _ = setupTestWithContext("1", sdk.NewDec(minGasPrices), sdkmath.NewInt(baseFee))
			})

			Context("during CheckTx", func() {
				DescribeTable("should reject transactions with EffectivePrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						_, err := testutil.CheckEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
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
					Entry("dynamic tx with GasFeeCap > MinGasPrices, EffectivePrice < MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(minGasPrices + 10_000_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res, err := testutil.CheckEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
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
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						_, err := testutil.DeliverEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
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
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res, err := testutil.DeliverEthTx(s.app, privKey, msgEthereumTx)
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

		Context("with MinGasPrices (feemarket param) < BaseFee (feemarket)", func() {
			var (
				baseFee      int64
				minGasPrices int64
			)

			BeforeEach(func() {
				baseFee = 10_000_000_000
				minGasPrices = baseFee - 5_000_000_000

				// Note that the tests run the same transactions with `gasLimit =
				// 100_000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
				// a `minGasPrices = 5_000_000_000` results in `minGlobalFee =
				// 500_000_000_000_000`
				privKey, _ = setupTestWithContext("1", sdk.NewDec(minGasPrices), sdkmath.NewInt(baseFee))
			})

			Context("during CheckTx", func() {
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						_, err := testutil.CheckEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
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
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						_, err := testutil.CheckEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"insufficient fee"),
						).To(BeTrue(), err.Error())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee - 1_000_000_000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee - 1_000_000_000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= EffectivePrice",
					func(malleate getprices) {
						p := malleate()
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res, err := testutil.CheckEthTx(s.app, privKey, msgEthereumTx)
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
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						_, err := testutil.DeliverEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
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
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						_, err := testutil.DeliverEthTx(s.app, privKey, msgEthereumTx)
						Expect(err).NotTo(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"insufficient fee"),
						).To(BeTrue(), err.Error())
					},
					// Note that the baseFee is not 10_000_000_000 anymore but updates to 8_750_000_000 because of the s.Commit
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
						to := utiltx.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res, err := testutil.DeliverEthTx(s.app, privKey, msgEthereumTx)
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
	})
})
