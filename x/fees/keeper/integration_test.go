package keeper_test

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/v3/app"
	"github.com/tharsis/evmos/v3/testutil"
	"github.com/tharsis/evmos/v3/x/fees/types"
	inflationtypes "github.com/tharsis/evmos/v3/x/inflation/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	// authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	abci "github.com/tendermint/tendermint/abci/types"
	claimstypes "github.com/tharsis/evmos/v3/x/claims/types"
)

var contractCode = "600661000e60003960066000f300612222600055"

// Uses CREATE to deploy the above contract and emits log1(0, 0, contractAddress)
var factoryCode = "603061000e60003960306000f3007f600661000e60003960066000f300612222600055000000000000000000000000600052601460006000f060006000a1"

var _ = Describe("While", Ordered, func() {
	// feeCollectorAddr := s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	// distrAddr := s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	claimsDenom := claimstypes.DefaultClaimsDenom
	evmDenom := evmtypes.DefaultEVMDenom
	accountCount := 4

	// account initial balances
	initAmount := sdk.NewInt(int64(math.Pow10(18) * 2))
	initBalance := sdk.NewCoins(
		sdk.NewCoin(claimsDenom, initAmount),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, initAmount),
	)
	totalAmount := sdk.NewCoin(claimsDenom, initAmount.MulRaw(int64(accountCount)))

	var (
		deployerKey      *ethsecp256k1.PrivKey
		userKey          *ethsecp256k1.PrivKey
		deployerAddress  sdk.AccAddress
		userAddress      sdk.AccAddress
		params           types.Params
		contractAddress  common.Address
		contractAddress1 common.Address
		contractAddress2 common.Address
		factoryAddress   common.Address
		factoryNonce     uint64
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.FeesKeeper.GetParams(s.ctx)
		params.EnableFees = true
		s.app.FeesKeeper.SetParams(s.ctx, params)

		// mint coins for claiming and send them to the claims module
		coins := sdk.NewCoins(totalAmount)
		err := testutil.FundModuleAccount(s.app.BankKeeper, s.ctx, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)

		// setup accounts
		deployerKey, _ = ethsecp256k1.GenerateKey()
		deployerAddress = getAddr(deployerKey)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress, initBalance)

		userKey, _ = ethsecp256k1.GenerateKey()
		userAddress = getAddr(userKey)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, userAddress, initBalance)
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, userAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Commit()

		// deploy contracts
		contractAddress = deployContract(deployerKey, contractCode)
		contractAddress1 = deployContract(deployerKey, contractCode)
		contractAddress2 = deployContract(deployerKey, contractCode)

		// deploy a factory
		factoryNonce = s.app.EvmKeeper.GetNonce(s.ctx, common.BytesToAddress(deployerAddress.Bytes()))
		factoryAddress = deployContract(deployerKey, factoryCode)

		// register a contract with default withdraw address
		registerDevFeeInfo(deployerKey, &contractAddress, 0)
		fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress)
		Expect(isRegistered).To(Equal(true))
		Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
		Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
		Expect(fee.WithdrawAddress).To(Equal(""))
		s.Commit()
	})

	Context("fee distribution is disabled", func() {
		BeforeAll(func() {
			params = s.app.FeesKeeper.GetParams(s.ctx)
			params.EnableFees = false
			s.app.FeesKeeper.SetParams(s.ctx, params)
		})

		It("we cannot register contracts for receiving tx fees", func() {
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			msg := types.NewMsgRegisterDevFeeInfo(contractAddress1, deployerAddress, deployerAddress, []uint64{1})

			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(false), "registration should have failed")
			s.Commit()

			_, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress1)
			Expect(isRegistered).To(Equal(false))
		})

		It("no tx fees go to developers", func() {
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)
			contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)
			s.Commit()

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance))
		})
	})

	Context("fee distribution is enabled", func() {
		BeforeEach(func() {
			params = types.DefaultParams()
			params.EnableFees = true
			s.app.FeesKeeper.SetParams(s.ctx, params)
		})

		It("we can register contracts for receiving tx fees, with default withdrawal address", func() {
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			msg := types.NewMsgRegisterDevFeeInfo(contractAddress2, deployerAddress, nil, []uint64{2})

			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
			s.Commit()

			fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
			Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
			Expect(fee.WithdrawAddress).To(Equal(""))
			s.Commit()
		})

		It("we can register contracts for receiving tx fees, with withdrawal address", func() {
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			withdrawAddress := sdk.AccAddress(tests.GenerateAddress().Bytes())
			msg := types.NewMsgRegisterDevFeeInfo(contractAddress1, deployerAddress, withdrawAddress, []uint64{1})

			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
			s.Commit()

			fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress1)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.ContractAddress).To(Equal(contractAddress1.Hex()))
			Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
			Expect(fee.WithdrawAddress).To(Equal(withdrawAddress.String()))

			preBalance := s.app.BankKeeper.GetBalance(s.ctx, withdrawAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)
			res = contractInteract(userKey, &contractAddress1, gasPrice, nil, nil, nil)
			s.Commit()

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())

			balance := s.app.BankKeeper.GetBalance(s.ctx, withdrawAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
		})

		It("legacy tx fees are split 50-50 validators-developers", func() {
			// preDistrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			// preFeeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, evmDenom)
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)
			res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())
			// validatorCoins := developerCoins

			// distrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			// feeColectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddr, evmDenom)
			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
			// Expect(distrBalance).To(Equal(preDistrBalance.Add(validatorCoins)))
			// Expect(preFeeColectorBalance).To(Equal(feeColectorBalance.Add(validatorCoins)))
			s.Commit()
		})

		It("dynamic tx fees are split 50-50 validators-developers", func() {
			// preDistrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasTipCap := big.NewInt(10000)
			gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
			res := contractInteract(userKey, &contractAddress, nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{})

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasFeeCap))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())
			// validatorCoins := developerCoins

			// distrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
			// Expect(distrBalance).To(Equal(preDistrBalance.Add(validatorCoins)))
			s.Commit()
		})

		It("tx fees are split 100-0 validators-developers", func() {
			params = s.app.FeesKeeper.GetParams(s.ctx)
			params.DeveloperShares = sdk.NewDec(0)
			params.ValidatorShares = sdk.NewDec(1)
			s.app.FeesKeeper.SetParams(s.ctx, params)

			// preDistrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasTipCap := big.NewInt(10000)
			gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
			res := contractInteract(userKey, &contractAddress, nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{})

			gasUsed := getGasUsedFromResponse(res, 10)
			fmt.Println("gasUsed", gasUsed)
			// feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasFeeCap))
			// validatorCoins := sdk.NewCoin(evmDenom, feeDistribution)

			// distrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance))
			// Expect(distrBalance).To(Equal(preDistrBalance.Add(validatorCoins)))
			s.Commit()
		})

		It("tx fees are split 0-100 validators-developers", func() {
			params = s.app.FeesKeeper.GetParams(s.ctx)
			params.DeveloperShares = sdk.NewDec(1)
			params.ValidatorShares = sdk.NewDec(0)
			s.app.FeesKeeper.SetParams(s.ctx, params)

			// preDistrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasTipCap := big.NewInt(10000)
			gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
			res := contractInteract(userKey, &contractAddress, nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{})

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasFeeCap))
			developerCoins := sdk.NewCoin(evmDenom, feeDistribution)

			// distrBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, evmDenom)
			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
			// Expect(distrBalance).To(Equal(preDistrBalance))
			s.Commit()
		})

		It("update of withdrawal address: withdraw address is different than the deployer address", func() {
			params = s.app.FeesKeeper.GetParams(s.ctx)
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			withdrawAddress := sdk.AccAddress(tests.GenerateAddress().Bytes())
			msg := types.NewMsgUpdateDevFeeInfo(contractAddress2, deployerAddress, withdrawAddress)

			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(true), "withdraw update failed: "+res.GetLog())
			s.Commit()

			fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress2)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.ContractAddress).To(Equal(contractAddress2.Hex()))
			Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
			Expect(fee.WithdrawAddress).To(Equal(withdrawAddress.String()))
			s.Commit()

			preBalanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			preBalanceW := s.app.BankKeeper.GetBalance(s.ctx, withdrawAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)
			res = contractInteract(userKey, &contractAddress2, gasPrice, nil, nil, nil)
			s.Commit()

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())

			balanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			balanceW := s.app.BankKeeper.GetBalance(s.ctx, withdrawAddress, evmDenom)
			Expect(balanceW).To(Equal(preBalanceW.Add(developerCoins)))
			Expect(balanceD).To(Equal(preBalanceD))
		})

		It("canceled developer fees: check that no fees are distributed", func() {
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			withdrawAddress, found := s.app.FeesKeeper.GetWithdrawal(s.ctx, contractAddress2)
			if !found {
				withdrawAddress = deployerAddress
			}
			msg := types.NewMsgCancelDevFeeInfo(contractAddress2, deployerAddress)

			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(true), "withdraw update failed: "+res.GetLog())
			s.Commit()

			fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress2)
			Expect(isRegistered).To(Equal(false))
			Expect(fee.ContractAddress).To(Equal(""))
			Expect(fee.DeployerAddress).To(Equal(""))
			Expect(fee.WithdrawAddress).To(Equal(""))
			s.Commit()

			preBalanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			preBalanceW := s.app.BankKeeper.GetBalance(s.ctx, withdrawAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)

			res = contractInteract(userKey, &contractAddress2, gasPrice, nil, nil, nil)
			s.Commit()

			balanceD := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			balanceW := s.app.BankKeeper.GetBalance(s.ctx, withdrawAddress, evmDenom)
			Expect(balanceW).To(Equal(preBalanceW))
			Expect(balanceD).To(Equal(preBalanceD))
		})

		It("factory generated contracts with legacy tx", func() {
			// Create contract through factory
			contractNonce := s.app.EvmKeeper.GetNonce(s.ctx, common.BytesToAddress(factoryAddress.Bytes()))
			contractAddress := deployContractWithFactory(deployerKey, &factoryAddress)
			s.Commit()

			// Register contract for receiving tx fees
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			msg := types.NewMsgRegisterDevFeeInfo(contractAddress, deployerAddress, nil, []uint64{factoryNonce, contractNonce})
			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
			s.Commit()

			// Check contract has been correctly registered
			fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
			Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
			Expect(fee.WithdrawAddress).To(Equal(""))

			// Get deployer balance before user interaction
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)

			// User interaction with registered contract
			gasPrice := big.NewInt(2000000000)
			res = contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)

			// Calculate fees amount
			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
			s.Commit()
		})

		It("factory generated contracts with dynamic fee txs", func() {
			// Create contract through factory
			contractNonce := s.app.EvmKeeper.GetNonce(s.ctx, common.BytesToAddress(factoryAddress.Bytes()))
			contractAddress := deployContractWithFactory(deployerKey, &factoryAddress)
			s.Commit()

			// Register contract for receiving tx fees
			deployerAddress := sdk.AccAddress(deployerKey.PubKey().Address().Bytes())
			msg := types.NewMsgRegisterDevFeeInfo(contractAddress, deployerAddress, nil, []uint64{factoryNonce, contractNonce})
			res := deliverTx(deployerKey, msg)
			Expect(res.IsOK()).To(Equal(true), "contract registration failed: "+res.GetLog())
			s.Commit()

			// Check contract has been correctly registered
			fee, isRegistered := s.app.FeesKeeper.GetFeeInfo(s.ctx, contractAddress)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
			Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
			Expect(fee.WithdrawAddress).To(Equal(""))

			// Get deployer balance before user interaction
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)

			// User interaction with registered contract
			gasTipCap := big.NewInt(10000)
			gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
			res = contractInteract(userKey, &contractAddress, nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{})

			// Calculate fees amount
			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasFeeCap))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
			s.Commit()
		})
	})
})

func getGasUsedFromResponse(res abci.ResponseDeliverTx, index int64) int64 {
	registerEvent := res.GetEvents()[index]
	Expect(registerEvent.Type).To(Equal("ethereum_tx"))
	Expect(string(registerEvent.Attributes[3].Key)).To(Equal("txGasUsed"))
	gasUsed, err := strconv.ParseInt(string(registerEvent.Attributes[3].Value), 10, 64)
	s.Require().NoError(err)
	return gasUsed
}

func registerDevFeeInfo(priv *ethsecp256k1.PrivKey, contractAddress *common.Address, nonce uint64) {
	deployerAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	msg := types.NewMsgRegisterDevFeeInfo(*contractAddress, deployerAddress, deployerAddress, []uint64{nonce})

	res := deliverTx(priv, msg)
	Expect(res.IsOK()).To(Equal(true), res.GetLog())

	s.Commit()
	registerEvent := res.GetEvents()[4]
	Expect(registerEvent.Type).To(Equal(types.EventTypeRegisterDevFeeInfo))
	Expect(string(registerEvent.Attributes[0].Key)).To(Equal(sdk.AttributeKeySender))
	Expect(string(registerEvent.Attributes[1].Key)).To(Equal(types.AttributeKeyContract))
}

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return sdk.AccAddress(priv.PubKey().Address().Bytes())
}

func deployContractWithFactory(priv *ethsecp256k1.PrivKey, factoryAddress *common.Address) common.Address {
	factoryNonce := s.app.EvmKeeper.GetNonce(s.ctx, *factoryAddress)
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)
	data := make([]byte, 0)
	msgEthereumTx := evmtypes.NewTx(chainID, nonce, factoryAddress, nil, uint64(100000), big.NewInt(1000000000), nil, nil, data, nil)
	msgEthereumTx.From = from.String()

	res := performEthTx(priv, msgEthereumTx)
	s.Commit()

	ethereumTx := res.GetEvents()[11]
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
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	data := common.Hex2Bytes(contractCode)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTxContract(chainID, nonce, nil, gasLimit, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	res := performEthTx(priv, msgEthereumTx)
	s.Commit()

	ethereumTx := res.GetEvents()[10]
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
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTx(chainID, nonce, contractAddr, nil, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
	msgEthereumTx.From = from.String()

	return performEthTx(priv, msgEthereumTx)
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	s.Require().NoError(err)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	s.Require().True(ok)
	builder.SetExtensionOptions(option)

	err = msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	err = txBuilder.SetMsgs(msgEthereumTx)
	s.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msgEthereumTx.Data)
	s.Require().NoError(err)

	fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewIntFromBigInt(txData.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msgEthereumTx.GetGas())

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	return res
}

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
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

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}
