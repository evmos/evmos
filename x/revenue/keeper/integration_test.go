package keeper_test

import (
	"math"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/evmos/v9/app"
	"github.com/evmos/evmos/v9/testutil"
	"github.com/evmos/evmos/v9/x/revenue/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

var contractCode = "600661000e60003960066000f300612222600055"

// Uses CREATE opcode to deploy the above contract and emits
// log1(0, 0, contractAddress)
var factoryCode = "603061000e60003960306000f3007f600661000e60003960066000f300612222600055000000000000000000000000600052601460006000f060006000a1"

// Creates the above factory
var doubleFactoryCode = "605461000e60003960546000f3007f603061000e60003960306000f3007f600661000e60003960066000f3006122226000527f600055000000000000000000000000600052601460006000f060006000a10000602052603e60006000f060006000a1"

var _ = Describe("Fee distribution:", Ordered, func() {
	feeCollectorAddr := s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	denom := s.denom

	// account initial balances
	initAmount := sdk.NewInt(int64(math.Pow10(18) * 4))
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))

	var (
		deployerKey     *ethsecp256k1.PrivKey
		userKey         *ethsecp256k1.PrivKey
		deployerAddress sdk.AccAddress
		userAddress     sdk.AccAddress
		params          types.Params
		factoryAddress  common.Address
		factoryNonce    uint64
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.RevenueKeeper.GetParams(s.ctx)
		params.EnableRevenue = true
		s.app.RevenueKeeper.SetParams(s.ctx, params)

		// setup deployer account
		deployerKey, deployerAddress = generateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress, initBalance)

		// setup account interacting with registered contracts
		userKey, userAddress = generateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, userAddress, initBalance)
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, userAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Commit()

		// deploy a factory
		factoryNonce = getNonce(deployerAddress.Bytes())
		factoryAddress = deployContract(deployerKey, factoryCode)
		s.Commit()
	})

	Context("with revenue param disabled", func() {
		var registeredContract common.Address

		BeforeAll(func() {
			// revenue registered before disabling params
			nonce := getNonce(deployerAddress.Bytes())
			registeredContract = deployContract(deployerKey, contractCode)
			res := registerFee(deployerKey, &registeredContract, nil, []uint64{nonce})
			Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

			fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, registeredContract)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.ContractAddress).To(Equal(registeredContract.Hex()))
			Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
			Expect(fee.WithdrawerAddress).To(Equal(""))
			s.Commit()

			// Disable revenue module
			params = s.app.RevenueKeeper.GetParams(s.ctx)
			params.EnableRevenue = false
			s.app.RevenueKeeper.SetParams(s.ctx, params)
		})

		It("should not allow new contract registrations", func() {
			contractAddress := deployContract(deployerKey, contractCode)
			msg := types.NewMsgRegisterRevenue(
				contractAddress,
				deployerAddress,
				withdraw,
				[]uint64{1},
			)

			res := deliverTx(deployerKey, nil, msg)
			Expect(res.IsOK()).To(Equal(false), "registration should have failed")
			Expect(
				strings.Contains(res.GetLog(),
					"revenue module is disabled by governance"),
			).To(BeTrue())
			s.Commit()

			_, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
			Expect(isRegistered).To(Equal(false))
		})

		It("should not distribute tx fees for previously registered contracts", func() {
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
			gasPrice := big.NewInt(2000000000)
			contractInteract(userKey, &registeredContract, gasPrice, nil, nil, nil)
			s.Commit()

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
			Expect(balance).To(Equal(preBalance))
		})

		It("should not allow revenue updates for previously registered contracts", func() {
			withdrawerAddress := sdk.AccAddress(tests.GenerateAddress().Bytes())
			msg := types.NewMsgUpdateRevenue(
				registeredContract,
				deployerAddress,
				withdrawerAddress,
			)
			res := deliverTx(deployerKey, nil, msg)
			Expect(res.IsOK()).To(Equal(false), "update should have failed")
			Expect(
				strings.Contains(res.GetLog(),
					"revenue module is disabled by governance"),
			).To(BeTrue())
			s.Commit()
		})

		It("should not allow cancellations of previously registered contracts", func() {
			msg := types.NewMsgCancelRevenue(registeredContract, deployerAddress)
			res := deliverTx(deployerKey, nil, msg)
			Expect(res.IsOK()).To(Equal(false), "cancel should have failed")
			Expect(
				strings.Contains(res.GetLog(),
					"revenue module is disabled by governance"),
			).To(BeTrue())
			s.Commit()
		})
	})

	Context("with revenue param enabled", func() {
		BeforeEach(func() {
			params = types.DefaultParams()
			params.EnableRevenue = true
			s.app.RevenueKeeper.SetParams(s.ctx, params)
		})

		Describe("Registering a contract for receiving tx fees", func() {
			Context("with an empty withdrawer address", Ordered, func() {
				var contractAddress common.Address
				var nonce uint64

				BeforeAll(func() {
					nonce = getNonce(deployerAddress.Bytes())
					contractAddress = deployContract(deployerKey, contractCode)
				})

				It("should be possible", func() {
					res := registerFee(deployerKey, &contractAddress, nil, []uint64{nonce})
					Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(""))
					s.Commit()
				})

				It("should result in sending the tx fees to the deployer address", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					gasPrice := big.NewInt(2000000000)
					res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)
					s.Commit()

					developerCoins, _ := calculateFees(denom, params, res, gasPrice, 14)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(developerCoins.IsPositive()).To(BeTrue())
					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
				})
			})

			Context("with a withdrawer address equal to the deployer address", func() {
				It("should be possible", func() {
					nonce := getNonce(deployerAddress.Bytes())
					contractAddress := deployContract(deployerKey, contractCode)
					res := registerFee(deployerKey, &contractAddress, deployerAddress, []uint64{nonce})
					Expect(res.IsOK()).To(BeTrue())

					_, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(BeTrue())
					s.Commit()
				})
			})

			Context("with an empty withdrawer address", func() {
				It("should be possible", func() {
					nonce := getNonce(deployerAddress.Bytes())
					contractAddress := deployContract(deployerKey, contractCode)
					res := registerFee(deployerKey, &contractAddress, nil, []uint64{nonce})
					Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(""))
					s.Commit()
				})
			})

			Context("with a withdrawer address different than deployer", Ordered, func() {
				var withdrawerAddress sdk.AccAddress
				var contractAddress common.Address
				var nonce uint64

				BeforeAll(func() {
					nonce = getNonce(deployerAddress.Bytes())
					contractAddress = deployContract(deployerKey, contractCode)
					withdrawerAddress = sdk.AccAddress(tests.GenerateAddress().Bytes())
				})

				It("should be possible", func() {
					res := registerFee(deployerKey, &contractAddress, withdrawerAddress, []uint64{nonce})
					Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

					registerEvent := res.GetEvents()[8]
					Expect(string(registerEvent.Attributes[2].Value)).ToNot(Equal(deployerAddress.String()))
					Expect(string(registerEvent.Attributes[2].Value)).To(Equal(withdrawerAddress.String()))

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(withdrawerAddress.String()))
				})

				It("should send the fees to the withdraw address", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, withdrawerAddress, denom)
					gasPrice := big.NewInt(2000000000)
					res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)
					s.Commit()

					developerCoins, _ := calculateFees(denom, params, res, gasPrice, 14)
					balance := s.app.BankKeeper.GetBalance(s.ctx, withdrawerAddress, denom)
					Expect(developerCoins.IsPositive()).To(BeTrue())
					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
				})
			})
		})

		Describe("Interacting with a registered revenue contract", func() {
			var contractAddress common.Address
			var nonce uint64

			BeforeAll(func() {
				nonce = getNonce(deployerAddress.Bytes())
				contractAddress = deployContract(deployerKey, contractCode)
				res := registerFee(deployerKey, &contractAddress, nil, []uint64{nonce})
				Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
			})

			Context("with a 50/50 validators-developers revenue", func() {
				BeforeEach(func() {
					params = s.app.RevenueKeeper.GetParams(s.ctx)
					params.DeveloperShares = sdk.NewDecWithPrec(50, 2)
					s.app.RevenueKeeper.SetParams(s.ctx, params)
				})

				It("should transfer legacy tx fees to validators and contract developer evenly", func() {
					preFeeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					gasPrice := big.NewInt(2000000000)
					res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)

					developerCoins, validatorCoins := calculateFees(denom, params, res, gasPrice, 14)
					feeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)

					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
					Expect(feeColectorBalance).To(Equal(
						preFeeColectorBalance.Add(validatorCoins),
					))
					s.Commit()
				})

				It("should transfer dynamic tx fees to validators and contract developer evenly", func() {
					preFeeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					gasTipCap := big.NewInt(10000)
					gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
					res := contractInteract(
						userKey,
						&contractAddress,
						nil,
						gasFeeCap,
						gasTipCap,
						&ethtypes.AccessList{},
					)

					developerCoins, validatorCoins := calculateFees(denom, params, res, gasFeeCap, 14)
					feeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
					Expect(feeColectorBalance).To(Equal(preFeeColectorBalance.Add(validatorCoins)))
					s.Commit()
				})
			})

			Context("with a 100/0 validators-developers revenue", func() {
				BeforeEach(func() {
					params = s.app.RevenueKeeper.GetParams(s.ctx)
					params.DeveloperShares = sdk.NewDec(0)
					s.app.RevenueKeeper.SetParams(s.ctx, params)
				})

				It("should transfer all tx fees to validators", func() {
					preFeeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					gasTipCap := big.NewInt(10000)
					gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
					res := contractInteract(
						userKey,
						&contractAddress,
						nil,
						gasFeeCap,
						gasTipCap,
						&ethtypes.AccessList{},
					)

					_, validatorCoins := calculateFees(denom, params, res, gasFeeCap, 10)
					feeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(balance).To(Equal(preBalance))
					Expect(feeColectorBalance).To(Equal(preFeeColectorBalance.Add(validatorCoins)))
					s.Commit()
				})
			})

			Context("with a 0/100 validators-developers revenue", func() {
				BeforeEach(func() {
					params = s.app.RevenueKeeper.GetParams(s.ctx)
					params.DeveloperShares = sdk.NewDec(1)
					s.app.RevenueKeeper.SetParams(s.ctx, params)
				})

				It("should transfer all tx fees to developers", func() {
					preFeeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					gasTipCap := big.NewInt(10000)
					gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
					res := contractInteract(
						userKey,
						&contractAddress,
						nil,
						gasFeeCap,
						gasTipCap,
						&ethtypes.AccessList{},
					)

					developerCoins, _ := calculateFees(denom, params, res, gasFeeCap, 14)
					feeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, denom)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
					Expect(feeColectorBalance).To(Equal(preFeeColectorBalance))
					s.Commit()
				})
			})
		})

		Describe("Updating registered revenue", func() {
			Context("with a withdraw address that is different from the deployer address", Ordered, func() {
				var withdrawerAddress sdk.AccAddress
				var contractAddress common.Address
				var nonce uint64

				BeforeAll(func() {
					nonce = getNonce(deployerAddress.Bytes())
					withdrawerAddress = sdk.AccAddress(tests.GenerateAddress().Bytes())
					contractAddress = deployContract(deployerKey, contractCode)
					res := registerFee(deployerKey, &contractAddress, nil, []uint64{nonce})
					Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(""))
				})

				It("should update revenue successfully", func() {
					msg := types.NewMsgUpdateRevenue(
						contractAddress,
						deployerAddress,
						withdrawerAddress,
					)

					res := deliverTx(deployerKey, nil, msg)
					Expect(res.IsOK()).To(
						Equal(true),
						"withdraw update failed: "+res.GetLog(),
					)
					s.Commit()

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(withdrawerAddress.String()))
					s.Commit()
				})

				It("should send tx fees to the new withdraw address", func() {
					preBalanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					preBalanceW := s.app.BankKeeper.GetBalance(s.ctx, withdrawerAddress, denom)
					gasPrice := big.NewInt(2000000000)
					res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)
					s.Commit()

					developerCoins, _ := calculateFees(denom, params, res, gasPrice, 14)
					balanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					balanceW := s.app.BankKeeper.GetBalance(s.ctx, withdrawerAddress, denom)
					Expect(balanceW).To(Equal(preBalanceW.Add(developerCoins)))
					Expect(balanceD).To(Equal(preBalanceD))
				})
			})

			Context("with a withdraw address equal to the deployer's address", func() {
				var contractAddress common.Address
				var nonce uint64

				BeforeAll(func() {
					nonce = getNonce(deployerAddress.Bytes())
					contractAddress = deployContract(deployerKey, contractCode)
					res := registerFee(deployerKey, &contractAddress, nil, []uint64{nonce})
					Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(""))
				})

				It("should not update revenue", func() {
					msg := types.NewMsgUpdateRevenue(
						contractAddress,
						deployerAddress,
						deployerAddress,
					)

					res := deliverTx(deployerKey, nil, msg)
					Expect(res.IsOK()).To(
						Equal(false),
						"withdraw update failed: "+res.GetLog(),
					)
					Expect(
						strings.Contains(res.GetLog(),
							"revenue already exists for given contract"),
					).To(BeTrue(), res.GetLog())
					s.Commit()
				})
			})

			Context("for a contract that was not registered", func() {
				It("should fail", func() {
					contractAddress := tests.GenerateAddress()
					withdrawerAddress := sdk.AccAddress(tests.GenerateAddress().Bytes())
					msg := types.NewMsgUpdateRevenue(
						contractAddress,
						deployerAddress,
						withdrawerAddress,
					)

					res := deliverTx(deployerKey, nil, msg)
					Expect(res.IsOK()).To(
						Equal(false),
						"withdraw update failed: "+res.GetLog(),
					)
					Expect(
						strings.Contains(res.GetLog(),
							"is not registered"),
					).To(BeTrue(), res.GetLog())
					s.Commit()
				})
			})
		})

		Describe("Canceling a revenue registration", func() {
			When("the registered revenue exists", Ordered, func() {
				var contractAddress common.Address
				var nonce uint64

				BeforeAll(func() {
					nonce = getNonce(deployerAddress.Bytes())
					contractAddress = deployContract(deployerKey, contractCode)
					registerFee(deployerKey, &contractAddress, nil, []uint64{nonce})
					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)

					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(""))
				})

				It("should be possible", func() {
					msg := types.NewMsgCancelRevenue(contractAddress, deployerAddress)
					res := deliverTx(deployerKey, nil, msg)
					Expect(res.IsOK()).To(Equal(true), "withdraw update failed: "+res.GetLog())
					s.Commit()

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(false))
					Expect(fee.ContractAddress).To(Equal(""))
					Expect(fee.DeployerAddress).To(Equal(""))
					Expect(fee.WithdrawerAddress).To(Equal(""))
					s.Commit()
				})

				It("should no longer distribute fees to the contract deployer", func() {
					preBalanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					gasPrice := big.NewInt(2000000000)

					contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)
					s.Commit()

					balanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(balanceD).To(Equal(preBalanceD))
				})
			})

			When("the revenue does not exist", func() {
				It("should not be possible", func() {
					contractAddress := tests.GenerateAddress()
					msg := types.NewMsgCancelRevenue(contractAddress, deployerAddress)
					res := deliverTx(deployerKey, nil, msg)
					Expect(res.IsOK()).To(
						Equal(false),
						"cancelling failed: "+res.GetLog(),
					)
					Expect(
						strings.Contains(res.GetLog(),
							"is not registered"),
					).To(BeTrue(), res.GetLog())
					s.Commit()
				})
			})
		})

		Describe("Registering contracts created by a factory contract with CREATE opcode", func() {
			Context("with one factory", Ordered, func() {
				var contractNonce uint64
				var contractAddress common.Address

				BeforeAll(func() {
					contractNonce = getNonce(factoryAddress.Bytes())
					contractAddress = deployContractWithFactory(deployerKey, &factoryAddress)
					s.Commit()
				})

				It("should be possible", func() {
					msg := types.NewMsgRegisterRevenue(
						contractAddress,
						deployerAddress,
						nil,
						[]uint64{factoryNonce, contractNonce},
					)
					res := deliverTx(deployerKey, nil, msg)
					Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
					s.Commit()

					fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
					Expect(isRegistered).To(Equal(true))
					Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
					Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
					Expect(fee.WithdrawerAddress).To(Equal(""))
				})

				It("should transfer legacy tx fees evenly to validator and deployer", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)

					// User interaction with registered contract
					gasPrice := big.NewInt(2000000000)
					res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)

					developerCoins, _ := calculateFees(denom, params, res, gasPrice, 14)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
					s.Commit()
				})

				It("should transfer dynamic tx fees evenly to validator and deployer", func() {
					preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)

					// User interaction with registered contract
					gasTipCap := big.NewInt(10000)
					gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
					res := contractInteract(
						userKey,
						&contractAddress,
						nil,
						gasFeeCap,
						gasTipCap,
						&ethtypes.AccessList{},
					)

					developerCoins, _ := calculateFees(denom, params, res, gasFeeCap, 14)
					balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, denom)
					Expect(balance).To(Equal(preBalance.Add(developerCoins)))
					s.Commit()
				})
			})

			Context("With factory-created factory contract", func() {
				var (
					gasUsedOneDerivation int64
					factory1Nonce        uint64
					factory2Nonce        uint64
					contractNonce        uint64
					factory1Address      common.Address
					factory2Address      common.Address
					contractAddress      common.Address
				)
				deployerKey1, deployerAddress1 := generateKey()
				deployerKey2, deployerAddress2 := generateKey()

				BeforeEach(func() {
					testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress1, initBalance)
					testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress2, initBalance)

					// Create contract: deployerKey1 -> factory1 -> factory2 -> contract
					// Create factory1
					factory1Nonce = getNonce(deployerAddress1.Bytes())
					factory1Address = deployContract(deployerKey1, doubleFactoryCode)

					// Create factory2
					factory2Nonce = getNonce(factory1Address.Bytes())
					factory2Address = deployContractWithFactory(deployerKey1, &factory1Address)

					// Create contract
					contractNonce = getNonce(factory2Address.Bytes())
					contractAddress = deployContractWithFactory(deployerKey1, &factory2Address)
				})

				DescribeTable("should consume gas for three address derivation iterations",
					func(gasCost int) {
						params = s.app.RevenueKeeper.GetParams(s.ctx)
						params.AddrDerivationCostCreate = uint64(gasCost)
						s.app.RevenueKeeper.SetParams(s.ctx, params)

						// Cost for registration with one address derivation
						// We use another deployer, to have the same storage cost for
						// SetDeployerFees
						factory1Nonce2 := getNonce(deployerAddress2.Bytes())
						factory1Address2 := deployContract(deployerKey2, doubleFactoryCode)
						res := registerFee(
							deployerKey2,
							&factory1Address2,
							nil,
							[]uint64{factory1Nonce2},
						)
						gasUsedOneDerivation = res.GetGasUsed()
						Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())

						s.Commit()

						// Registering contract for receiving fees
						// Use a new deployer, to pay the same storage costs for SetDeployerFees
						res = registerFee(
							deployerKey1,
							&contractAddress,
							nil,
							[]uint64{factory1Nonce, factory2Nonce, contractNonce},
						)
						Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
						s.Commit()

						fee, isRegistered := s.app.RevenueKeeper.GetRevenue(s.ctx, contractAddress)
						Expect(isRegistered).To(Equal(true))
						Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
						Expect(fee.DeployerAddress).To(Equal(deployerAddress1.String()))
						Expect(fee.WithdrawerAddress).To(Equal(""))

						// Check addressDerivationCostCreate is subtracted 3 times
						setFeeInverseCost := int64(20)
						Expect(res.GetGasUsed()).To(Equal(
							gasUsedOneDerivation + int64(gasCost)*2 + setFeeInverseCost,
						))
					},
					// FIXME: make both test Entries pass
					Entry("with address derivation cost of 50", 50),
					Entry("with address derivation cost of 500", 500),
				)
			})
		})
	})
})

func calculateFees(
	denom string,
	params types.Params,
	res abci.ResponseDeliverTx,
	gasPrice *big.Int,
	logIndex int64,
) (sdk.Coin, sdk.Coin) {
	feeDistribution := sdk.NewInt(res.GasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
	developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
	developerCoins := sdk.NewCoin(denom, developerFee.TruncateInt())
	validatorShares := sdk.OneDec().Sub(params.DeveloperShares)
	validatorFee := sdk.NewDecFromInt(feeDistribution).Mul(validatorShares)
	validatorCoins := sdk.NewCoin(denom, validatorFee.TruncateInt())
	return developerCoins, validatorCoins
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
}

func registerFee(
	priv *ethsecp256k1.PrivKey,
	contractAddress *common.Address,
	withdrawerAddress sdk.AccAddress,
	nonces []uint64,
) abci.ResponseDeliverTx {
	deployerAddress := sdk.AccAddress(priv.PubKey().Address())
	msg := types.NewMsgRegisterRevenue(*contractAddress, deployerAddress, withdrawerAddress, nonces)

	res := deliverTx(priv, nil, msg)
	s.Commit()

	if res.IsOK() {
		registerEvent := res.GetEvents()[8]
		Expect(registerEvent.Type).To(Equal(types.EventTypeRegisterRevenue))
		Expect(string(registerEvent.Attributes[0].Key)).To(Equal(sdk.AttributeKeySender))
		Expect(string(registerEvent.Attributes[1].Key)).To(Equal(types.AttributeKeyContract))
		Expect(string(registerEvent.Attributes[2].Key)).To(Equal(types.AttributeKeyWithdrawerAddress))
	}
	return res
}

func generateKey() (*ethsecp256k1.PrivKey, sdk.AccAddress) {
	address, priv := tests.NewAddrKey()
	return priv.(*ethsecp256k1.PrivKey), sdk.AccAddress(address.Bytes())
}

func deployContractWithFactory(priv *ethsecp256k1.PrivKey, factoryAddress *common.Address) common.Address {
	factoryNonce := getNonce(factoryAddress.Bytes())
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)
	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		factoryAddress,
		nil,
		uint64(100000),
		big.NewInt(1000000000),
		nil,
		nil,
		data,
		nil,
	)
	msgEthereumTx.From = from.String()

	res := deliverEthTx(priv, msgEthereumTx)
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	s.Commit()

	ethereumTx := res.GetEvents()[12]
	Expect(ethereumTx.Type).To(Equal("tx_log"))
	Expect(string(ethereumTx.Attributes[0].Key)).To(Equal("txLog"))
	txLog := string(ethereumTx.Attributes[0].Value)

	contractAddress := crypto.CreateAddress(*factoryAddress, factoryNonce)
	Expect(
		strings.Contains(txLog, strings.ToLower(contractAddress.String()[2:])),
	).To(BeTrue(), "log topic does not match created contract address")

	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc, "contract not created")
	s.Require().True(acc.IsContract(), "not a contract")
	return contractAddress
}

func deployContract(priv *ethsecp256k1.PrivKey, contractCode string) common.Address {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())

	data := common.Hex2Bytes(contractCode)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTxContract(
		chainID,
		nonce,
		nil,
		gasLimit,
		nil,
		s.app.FeeMarketKeeper.GetBaseFee(s.ctx),
		big.NewInt(1),
		data,
		&ethtypes.AccessList{},
	)
	msgEthereumTx.From = from.String()

	res := deliverEthTx(priv, msgEthereumTx)
	s.Commit()

	ethereumTx := res.GetEvents()[11]
	Expect(ethereumTx.Type).To(Equal("ethereum_tx"))
	Expect(string(ethereumTx.Attributes[1].Key)).To(Equal("ethereumTxHash"))

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func contractInteract(
	priv *ethsecp256k1.PrivKey,
	contractAddr *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) abci.ResponseDeliverTx {
	msgEthereumTx := buildEthTx(priv, contractAddr, gasPrice, gasFeeCap, gasTipCap, accesses)
	res := deliverEthTx(priv, msgEthereumTx)
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	return res
}

func buildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		to,
		nil,
		gasLimit,
		gasPrice,
		gasFeeCap,
		gasTipCap,
		data,
		accesses,
	)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}

func prepareEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) []byte {
	// Sign transaction
	err := msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	// Assemble transaction from fields
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	tx, err := msgEthereumTx.BuildTx(txBuilder, s.app.EvmKeeper.GetParams(s.ctx).EvmDenom)
	s.Require().NoError(err)

	// Encode transaction by default Tx encoder and broadcasted over the network
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(tx)
	s.Require().NoError(err)

	return bz
}

func deliverEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func checkEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseCheckTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestCheckTx{Tx: bz}
	res := s.app.BaseApp.CheckTx(req)
	return res
}

func prepareCosmosTx(priv *ethsecp256k1.PrivKey, gasPrice *sdkmath.Int, msgs ...sdk.Msg) []byte {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	if gasPrice == nil {
		_gasPrice := sdk.NewInt(1)
		gasPrice = &_gasPrice
	}
	fees := &sdk.Coins{{Denom: denom, Amount: gasPrice.MulRaw(1000000)}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.ctx, accountAddress)
	s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	s.Require().NoError(err)

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}

func deliverTx(priv *ethsecp256k1.PrivKey, gasPrice *sdkmath.Int, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := prepareCosmosTx(priv, gasPrice, msgs...)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func checkTx(priv *ethsecp256k1.PrivKey, gasPrice *sdkmath.Int, msgs ...sdk.Msg) abci.ResponseCheckTx {
	bz := prepareCosmosTx(priv, gasPrice, msgs...)
	req := abci.RequestCheckTx{Tx: bz}
	res := s.app.BaseApp.CheckTx(req)
	return res
}
