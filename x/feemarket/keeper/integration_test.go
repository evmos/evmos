package keeper_test

import (
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/math"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v19/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v19/testutil"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const chainID = utils.TestnetChainID + "-1"

var _ = Describe("Feemarket", func() {
	var (
		privKey *ethsecp256k1.PrivKey
		msg     banktypes.MsgSend
	)

	testSetup := []struct {
		name          string
		denomDecimals uint32
	}{
		{
			name:          "6 decimals denom",
			denomDecimals: evmtypes.Denom6Dec,
		},
		{
			name:          "18 decimals denom",
			denomDecimals: evmtypes.Denom18Dec,
		},
	}

	for _, setup := range testSetup {
		Describe(fmt.Sprintf("Performing Cosmos transactions - %s", setup.name), func() {
			Context("with min-gas-prices (local) < MinGasPrices (feemarket param)", func() {
				BeforeEach(func() {
					privKey, msg = setupTestWithContext(chainID, "1", math.LegacyNewDec(3), math.LegacyZeroDec(), setup.denomDecimals)
				})

				Context("during CheckTx", func() {
					It("should reject transactions with gasPrice < MinGasPrices", func() {
						gasPrice := math.NewInt(2)
						_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
					})

					It("should accept transactions with gasPrice >= MinGasPrices", func() {
						gasPrice := math.NewInt(3)
						res, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					})
				})

				Context("during DeliverTx", func() {
					It("should reject transactions with gasPrice < MinGasPrices", func() {
						gasPrice := math.NewInt(2)
						_, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).NotTo(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
					})

					It("should accept transactions with gasPrice >= MinGasPrices", func() {
						gasPrice := math.NewInt(3)
						res, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						s.Require().NoError(err)
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					})
				})
			})

			Context("with min-gas-prices (local) == MinGasPrices (feemarket param)", func() {
				BeforeEach(func() {
					privKey, msg = setupTestWithContext(chainID, "3", math.LegacyNewDec(3), math.LegacyZeroDec(), setup.denomDecimals)
				})

				Context("during CheckTx", func() {
					It("should reject transactions with gasPrice < min-gas-prices", func() {
						gasPrice := math.NewInt(2)
						_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"insufficient fee"),
						).To(BeTrue(), err.Error())
					})

					It("should accept transactions with gasPrice >= MinGasPrices", func() {
						gasPrice := math.NewInt(3)
						res, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					})
				})

				Context("during DeliverTx", func() {
					It("should reject transactions with gasPrice < MinGasPrices", func() {
						gasPrice := math.NewInt(2)
						_, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).NotTo(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
					})

					It("should accept transactions with gasPrice >= MinGasPrices", func() {
						gasPrice := math.NewInt(3)
						res, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					})
				})
			})

			Context("with MinGasPrices (feemarket param) < min-gas-prices (local)", func() {
				BeforeEach(func() {
					privKey, msg = setupTestWithContext(chainID, "5", math.LegacyNewDec(3), math.LegacyNewDec(5), setup.denomDecimals)
				})

				//nolint
				Context("during CheckTx", func() {
					It("should reject transactions with gasPrice < MinGasPrices", func() {
						gasPrice := math.NewInt(2)
						_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"insufficient fee"),
						).To(BeTrue(), err.Error())
					})

					It("should reject transactions with MinGasPrices < gasPrice < baseFee", func() {
						gasPrice := math.NewInt(4)
						_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"insufficient fee"),
						).To(BeTrue(), err.Error())
					})

					It("should accept transactions with gasPrice >= baseFee", func() {
						gasPrice := math.NewInt(5)
						res, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).To(BeNil())
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					})
				})

				//nolint
				Context("during DeliverTx", func() {
					It("should reject transactions with gasPrice < MinGasPrices", func() {
						gasPrice := math.NewInt(2)
						_, err := testutil.DeliverTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).NotTo(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), err.Error())
					})

					It("should reject transactions with MinGasPrices < gasPrice < baseFee", func() {
						gasPrice := math.NewInt(4)
						_, err := testutil.CheckTx(s.ctx, s.app, privKey, &gasPrice, &msg)
						Expect(err).ToNot(BeNil(), "transaction should have failed")
						Expect(
							strings.Contains(err.Error(),
								"insufficient fee"),
						).To(BeTrue(), err.Error())
					})
					It("should accept transactions with gasPrice >= baseFee", func() {
						gasPrice := math.NewInt(5)
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

			Context(fmt.Sprintf("%s - with BaseFee (feemarket) < MinGasPrices (feemarket param)", setup.name), func() {
				var (
					baseFee      int64
					minGasPrices int64
				)

				BeforeEach(func() {
					// These are on the evm denom
					// keep this in mind when submitting an eth tx and
					// the evm denom has 6 decimals
					baseFee = 10_000_000_000
					minGasPrices = baseFee + 30_000_000_000

					// Note that the tests run the same transactions with `gasLimit =
					// 100000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
					// a `minGasPrices = 40_000_000_000` results in `minGlobalFee =
					// 4000000000000000`
					privKey, _ = setupTestWithContext(chainID, "1", math.LegacyNewDec(minGasPrices), math.LegacyNewDec(baseFee), setup.denomDecimals)
				})

				Context("during CheckTx", func() {
					DescribeTable("should reject transactions with EffectivePrice < MinGasPrices",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							_, err := testutil.CheckEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).ToNot(BeNil(), "transaction should have failed")
							Expect(
								strings.Contains(err.Error(),
									"provided fee < minimum global fee"),
							).To(BeTrue(), err.Error())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrices := big.NewInt(minGasPrices - 10_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrices = evmtypes.Convert6To18DecimalsBigInt(gasPrices)
							}
							return txParams{gasPrices, nil, nil, nil}
						}),
						Entry("dynamic tx with GasFeeCap < MinGasPrices, no gasTipCap", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 10_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
						Entry("dynamic tx with GasFeeCap < MinGasPrices, max gasTipCap", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 10_000_000_000)
							gasTipCap := big.NewInt(30_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
								gasTipCap = evmtypes.Convert6To18DecimalsBigInt(gasTipCap)
							}
							// Note that max priority fee per gas can't be higher than the max fee per gas (gasFeeCap), i.e. 30_000_000_000)
							return txParams{nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{}}
						}),
						Entry("dynamic tx with GasFeeCap > MinGasPrices, EffectivePrice < MinGasPrices", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 10_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
					)

					DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							res, err := testutil.CheckEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).To(BeNil())
							Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(minGasPrices)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						// Note that this tx is not rejected on CheckTx, but not on DeliverTx,
						// as the baseFee is set to minGasPrices during DeliverTx when baseFee
						// < minGasPrices
						Entry("dynamic tx with GasFeeCap > MinGasPrices, EffectivePrice > MinGasPrices", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices)
							gasTipCap := big.NewInt(30_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
								gasTipCap = evmtypes.Convert6To18DecimalsBigInt(gasTipCap)
							}
							return txParams{nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{}}
						}),
					)
				})

				Context("during DeliverTx", func() {
					DescribeTable("should reject transactions with gasPrice < MinGasPrices",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							_, err := testutil.DeliverEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).ToNot(BeNil(), "transaction should have failed")
							Expect(
								strings.Contains(err.Error(),
									"provided fee < minimum global fee"),
							).To(BeTrue(), err.Error())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(minGasPrices - 10_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx with GasFeeCap < MinGasPrices, no gasTipCap", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 10_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
						Entry("dynamic tx with GasFeeCap < MinGasPrices, max gasTipCap", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 10_000_000_000)
							gasTipCap := big.NewInt(30_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
								gasTipCap = evmtypes.Convert6To18DecimalsBigInt(gasTipCap)
							}
							// Note that max priority fee per gas can't be higher than the max fee per gas (gasFeeCap), i.e. 30_000_000_000)
							return txParams{nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{}}
						}),
					)

					DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							res, err := testutil.DeliverEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).To(BeNil(), "transaction should have succeeded")
							Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(minGasPrices + 1)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx, EffectivePrice > MinGasPrices", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices + 10_000_000_000)
							gasTipCap := big.NewInt(30_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
								gasTipCap = evmtypes.Convert6To18DecimalsBigInt(gasTipCap)
							}
							return txParams{nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{}}
						}),
					)
				})
			})

			Context(fmt.Sprintf("%s - with MinGasPrices (feemarket param) < BaseFee (feemarket)", setup.name), func() {
				var (
					baseFee      int64
					minGasPrices int64
				)

				BeforeEach(func() {
					// These are on the evm denom
					// keep this in mind when submitting an eth tx and
					// the evm denom has 6 decimals
					baseFee = 10_000_000_000
					minGasPrices = baseFee - 5_000_000_000

					// Note that the tests run the same transactions with `gasLimit =
					// 100_000`. With the fee calculation `Fee = (baseFee + tip) * gasLimit`,
					// a `minGasPrices = 5_000_000_000` results in `minGlobalFee =
					// 500_000_000_000_000`
					privKey, _ = setupTestWithContext(chainID, "1", math.LegacyNewDec(minGasPrices), math.LegacyNewDec(baseFee), setup.denomDecimals)

					// setup evm params
					evmParams := s.app.EvmKeeper.GetParams(s.ctx)
					evmParams.DenomDecimals = setup.denomDecimals
					s.app.EvmKeeper.SetParams(s.ctx, evmParams)
					// s.Commit()
				})

				Context("during CheckTx", func() {
					DescribeTable("should reject transactions with gasPrice < MinGasPrices",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							_, err := testutil.CheckEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).ToNot(BeNil(), "transaction should have failed")
							Expect(
								strings.Contains(err.Error(),
									"provided fee < minimum global fee"),
							).To(BeTrue(), err.Error())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(minGasPrices - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx with GasFeeCap < MinGasPrices, no gasTipCap", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
						Entry("dynamic tx with GasFeeCap < MinGasPrices, max gasTipCap", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 1_000_000_000)
							gasTipCap := big.NewInt(minGasPrices - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
								gasTipCap = evmtypes.Convert6To18DecimalsBigInt(gasTipCap)
							}
							return txParams{nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{}}
						}),
					)

					DescribeTable("should reject transactions with MinGasPrices < tx gasPrice < EffectivePrice",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							_, err := testutil.CheckEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).ToNot(BeNil(), "transaction should have failed")
							Expect(
								strings.Contains(err.Error(),
									"insufficient fee"),
							).To(BeTrue(), err.Error())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(baseFee - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(baseFee - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
					)

					DescribeTable("should accept transactions with gasPrice >= EffectivePrice",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							res, err := testutil.CheckEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).To(BeNil(), "transaction should have succeeded")
							Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(baseFee)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(baseFee)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
					)
				})

				Context("during DeliverTx", func() {
					DescribeTable("should reject transactions with gasPrice < MinGasPrices",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							_, err := testutil.DeliverEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).ToNot(BeNil(), "transaction should have failed")
							Expect(
								strings.Contains(err.Error(),
									"provided fee < minimum global fee"),
							).To(BeTrue(), err.Error())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(minGasPrices - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(minGasPrices - 1_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, nil, &ethtypes.AccessList{}}
						}),
					)

					DescribeTable("should reject transactions with MinGasPrices < gasPrice < EffectivePrice",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							_, err := testutil.DeliverEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).NotTo(BeNil(), "transaction should have failed")
							Expect(
								strings.Contains(err.Error(),
									"insufficient fee"),
							).To(BeTrue(), err.Error())
						},
						// Note that the baseFee is not 10_000_000_000 anymore but updates to 8_750_000_000 because of the s.Commit
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(baseFee - 2_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(baseFee - 2_000_000_000)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
					)

					DescribeTable("should accept transactions with gasPrice >= EffectivePrice",
						func(malleate getprices) {
							p := malleate()
							to := utiltx.GenerateAddress()
							msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
							res, err := testutil.DeliverEthTx(s.ctx, s.app, privKey, msgEthereumTx)
							Expect(err).To(BeNil())
							Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
						},
						Entry("legacy tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasPrice := big.NewInt(baseFee)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
							}
							return txParams{gasPrice, nil, nil, nil}
						}),
						Entry("dynamic tx", func() txParams {
							evmParams := s.app.EvmKeeper.GetParams(s.ctx)
							gasFeeCap := big.NewInt(baseFee)
							// if evm denom has 6 decimals, need to scale to 18 decimals
							if evmParams.DenomDecimals == evmtypes.Denom6Dec {
								gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
							}
							return txParams{nil, gasFeeCap, big.NewInt(0), &ethtypes.AccessList{}}
						}),
					)
				})
			})
		})
	}
})
