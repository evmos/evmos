package werc20_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v20/precompiles/erc20"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/precompiles/werc20"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"

	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"

	"github.com/ethereum/go-ethereum/common"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// -------------------------------------------------------------------------------------------------
// Integration test suite
// -------------------------------------------------------------------------------------------------

var is *PrecompileIntegrationTestSuite

type PrecompileIntegrationTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     keyring.Keyring

	wrappedCoinDenom string

	// WEVMOS related fields
	precompile        *werc20.Precompile
	precompileAddrHex string
}

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	is = new(PrecompileIntegrationTestSuite)
	suite.Run(t, is)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "WEVMOS precompile test suite")
}

func (is *PrecompileIntegrationTestSuite) SetupTest() {
	keyring := keyring.New(2)

	customGenesis := network.CustomGenesisState{}
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.NoBaseFee = true
	customGenesis[feemarkettypes.ModuleName] = feemarketGenesis

	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	is.network = integrationNetwork
	is.factory = txFactory
	is.grpcHandler = grpcHandler
	is.keyring = keyring

	is.wrappedCoinDenom = evmtypes.GetEVMCoinDenom()

	chainID := is.network.GetChainID()
	is.precompileAddrHex = erc20types.GetWEVMOSContractHex(chainID)

	ctx := integrationNetwork.GetContext()

	// Check that WEVMOS is part of the native precompiles.
	erc20Params := is.network.App.Erc20Keeper.GetParams(ctx)
	Expect(erc20Params.NativePrecompiles).To(
		ContainElement(is.precompileAddrHex),
		"expected wevmos to be in the native precompiles",
	)

	// Check that WEVMOS is registered in the token pairs map.
	tokenPairID := is.network.App.Erc20Keeper.GetTokenPairID(ctx, is.wrappedCoinDenom)
	tokenPair, found := is.network.App.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
	Expect(found).To(BeTrue(), "expected wevmos precompile to be registered in the tokens map")
	Expect(tokenPair.Erc20Address).To(Equal(is.precompileAddrHex))
}

// checkAndReturnBalance check that the balance of the address is the same in
// the smart contract and in the balance and returns the amount.
func (is *PrecompileIntegrationTestSuite) checkAndReturnBalance(
	balanceCheck testutil.LogCheckArgs,
	callsData CallsData,
	address common.Address,
) *big.Int {
	txArgs, balancesArgs := callsData.getTxAndCallArgs(directCall, erc20.BalanceOfMethod, address)

	_, ethRes, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, balancesArgs, balanceCheck)
	Expect(err).ToNot(HaveOccurred(), "failed to execute balanceOf")
	var erc20Balance *big.Int
	err = is.precompile.UnpackIntoInterface(&erc20Balance, erc20.BalanceOfMethod, ethRes.Ret)
	Expect(err).ToNot(HaveOccurred(), "failed to unpack result")

	addressAcc := sdk.AccAddress(address.Bytes())
	balanceAfter, err := is.grpcHandler.GetBalance(addressAcc, is.wrappedCoinDenom)
	Expect(err).ToNot(HaveOccurred(), "expected no error getting balance")

	Expect(erc20Balance.String()).To(Equal(balanceAfter.Balance.Amount.BigInt().String()), "expected return balance from contract equal to bank")
	return erc20Balance
}

// -------------------------------------------------------------------------------------------------
// Integration tests
// -------------------------------------------------------------------------------------------------

var _ = When("a user interact with the WEVMOS precompiled contract", func() {
	var (
		// callArgs factory.CallArgs
		// txArgs   evmtypes.EvmTxArgs

		passCheck, failCheck, depositCheck, withdrawCheck testutil.LogCheckArgs

		callsData CallsData

		txSender keyring.Key
		user     keyring.Key
	)

	depositAmount := big.NewInt(1e18)
	withdrawAmount := depositAmount

	BeforeEach(func() {
		is.SetupTest()

		precompileAddr := common.HexToAddress(is.precompileAddrHex)
		tokenPair := erc20types.NewTokenPair(
			precompileAddr,
			evmtypes.GetEVMCoinDenom(),
			erc20types.OWNER_MODULE,
		)

		precompile, err := werc20.NewPrecompile(
			tokenPair,
			is.network.App.BankKeeper,
			is.network.App.AuthzKeeper,
			is.network.App.TransferKeeper,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to instantiate the werc20 precompile")
		is.precompile = precompile

		failCheck = testutil.LogCheckArgs{ABIEvents: is.precompile.Events}
		passCheck = failCheck.WithExpPass(true)

		withdrawCheck = passCheck.WithExpEvents(werc20.EventTypeWithdrawal)
		depositCheck = passCheck.WithExpEvents(werc20.EventTypeDeposit)

		txSender = is.keyring.GetKey(0)
		// user = s.keyring.GetKey(1)

		callsData = CallsData{
			sender: txSender,

			precompileAddr: precompileAddr,
			precompileABI:  precompile.ABI,
		}
	})

	Context("and funds are part of the transaction", func() {
		When("the method is deposit", func() {
			It("it should return funds to sender and emit the event", func() {
				// Store initial balance to verify that sender
				// balance remains the same after the contract call.
				initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
				txArgs.Amount = depositAmount

				_, _, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(finalBalance).To(Equal(initBalance))
			})
			It("it should consume at least the deposit requested gas", func() {
				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
				txArgs.Amount = depositAmount

				_, ethRes, _ := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for deposit")
			})
		})
		When("no calldata is provided", func() {
			It("it should call the receive which behave like deposit", func() {
				// Store initial balance to verify withdraw is a no-op and sender
				// balance remains the same after the contract call.
				initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
				txArgs.Amount = depositAmount

				_, _, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(finalBalance).To(Equal(initBalance))
			})
			It("it should consume at least the deposit requested gas", func() {
				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.DepositMethod)
				txArgs.Amount = depositAmount

				_, ethRes, _ := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for receive")
			})
		})
		When("the specified method is too short", func() {
			It("it should call the fallback which behave like deposit", func() {
				// Store initial balance to verify withdraw is a no-op and sender
				// balance remains the same after the contract call.
				initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
				txArgs.Amount = depositAmount
				// Short method is directly set in the input to skip ABI validation
				txArgs.Input = []byte{1, 2, 3}

				_, _, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(finalBalance).To(Equal(initBalance))
			})
			It("it should consume at least the deposit requested gas", func() {
				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
				txArgs.Amount = depositAmount
				// Short method is directly set in the input to skip ABI validation
				txArgs.Input = []byte{1, 2, 3}

				_, ethRes, _ := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for fallback")
			})
		})
		When("the specified method does not exist", func() {
			It("it should call the fallback which behave like deposit", func() {
				// Store initial balance to verify withdraw is a no-op and sender
				// balance remains the same after the contract call.
				initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
				txArgs.Amount = depositAmount
				// Wrong method is directly set in the input to skip ABI validation
				txArgs.Input = []byte("nonExistingMethod")

				_, _, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(finalBalance).To(Equal(initBalance))
			})
			It("it should consume at least the deposit requested gas", func() {
				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, "")
				txArgs.Amount = depositAmount
				// Wrong method is directly set in the input to skip ABI validation
				txArgs.Input = []byte("nonExistingMethod")

				_, ethRes, _ := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, depositCheck)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.DepositRequiredGas), "expected different gas used for fallback")
			})
		})
	})
	Context("and funds are NOT part of the transaction", func() {
		When("the method is withdraw", func() {
			It("it should be a no-op and emit the event", func() {
				// Store initial balance to verify withdraw is a no-op and sender
				// balance remains the same after the contract call.
				initBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.WithdrawMethod, withdrawAmount)

				_, _, err := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, withdrawCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected error calling the precompile")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				finalBalance := is.checkAndReturnBalance(passCheck, callsData, user.Addr)
				Expect(finalBalance).To(Equal(initBalance))
			})
			It("it should consume at least the withdraw requested gas", func() {
				txArgs, callArgs := callsData.getTxAndCallArgs(directCall, werc20.WithdrawMethod, withdrawAmount)

				_, ethRes, _ := is.factory.CallContractAndCheckLogs(callsData.sender.Priv, txArgs, callArgs, withdrawCheck)
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				Expect(ethRes.GasUsed).To(BeNumerically(">=", werc20.WithdrawRequiredGas), "expected different gas used for withdraw")
			})
		})
	})
})
