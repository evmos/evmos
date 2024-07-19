package bank_test

import (
	"math/big"
	"testing"

	"github.com/evmos/evmos/v19/precompiles/bank/testdata"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	integrationutils "github.com/evmos/evmos/v19/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v19/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v19/x/inflation/v1/types"

	evmosutiltx "github.com/evmos/evmos/v19/testutil/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/precompiles/bank"

	"github.com/evmos/evmos/v19/precompiles/testutil"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var is *IntegrationTestSuite

// IntegrationTestSuite is the implementation of the TestSuite interface for Bank precompile
// unit testis.
type IntegrationTestSuite struct {
	bondDenom, tokenDenom string
	evmosAddr, xmplAddr   common.Address

	// tokenDenom is the specific token denomination used in testing the Bank precompile.
	// This denomination is used to instantiate the precompile.
	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     keyring.Keyring

	precompile *bank.Precompile
}

func (is *IntegrationTestSuite) SetupTest() {
	keyring := keyring.New(2)
	genesis := integrationutils.CreateGenesisWithTokenPairs(keyring)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(genesis),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	ctx := integrationNetwork.GetContext()
	sk := integrationNetwork.App.StakingKeeper
	bondDenom := sk.BondDenom(ctx)
	Expect(bondDenom).ToNot(BeEmpty(), "bond denom cannot be empty")

	is.bondDenom = bondDenom
	is.tokenDenom = "xmpl"
	is.factory = txFactory
	is.grpcHandler = grpcHandler
	is.keyring = keyring
	is.network = integrationNetwork

	tokenPairID := is.network.App.Erc20Keeper.GetTokenPairID(is.network.GetContext(), is.bondDenom)
	tokenPair, found := is.network.App.Erc20Keeper.GetTokenPair(is.network.GetContext(), tokenPairID)
	Expect(found).To(BeTrue(), "failed to register token erc20 extension")
	is.evmosAddr = common.HexToAddress(tokenPair.Erc20Address)

	// Mint and register a second coin for testing purposes
	err := is.network.App.BankKeeper.MintCoins(is.network.GetContext(), inflationtypes.ModuleName, sdk.Coins{{Denom: is.tokenDenom, Amount: sdk.NewInt(1e18)}})
	Expect(err).ToNot(HaveOccurred(), "failed to mint coin")

	tokenPairID = is.network.App.Erc20Keeper.GetTokenPairID(is.network.GetContext(), is.tokenDenom)
	tokenPair, found = is.network.App.Erc20Keeper.GetTokenPair(is.network.GetContext(), tokenPairID)
	Expect(found).To(BeTrue(), "failed to register token erc20 extension")
	is.xmplAddr = common.HexToAddress(tokenPair.Erc20Address)
	is.precompile = is.setupBankPrecompile()
}

func TestIntegrationSuite(t *testing.T) {
	is = new(IntegrationTestSuite)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bank Extension Suite")
}

var _ = Describe("Bank Extension -", func() {
	var (
		bankCallerContractAddr common.Address
		bankCallerContract     evmtypes.CompiledContract

		err    error
		sender keyring.Key
		amount *big.Int

		// contractData is a helper struct to hold the addresses and ABIs for the
		// different contract instances that are subject to testing here.
		contractData ContractData
		passCheck    testutil.LogCheckArgs
	)

	BeforeEach(func() {
		is.SetupTest()

		// Default sender, amount
		sender = is.keyring.GetKey(0)
		amount = big.NewInt(1e18)

		bankCallerContract, err = testdata.LoadBankCallerContract()
		Expect(err).ToNot(HaveOccurred(), "failed to load BankCaller contract")

		bankCallerContractAddr, err = is.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: bankCallerContract,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter burner contract")

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			precompileAddr: is.precompile.Address(),
			precompileABI:  is.precompile.ABI,
			contractAddr:   bankCallerContractAddr,
			contractABI:    bankCallerContract.ABI,
		}

		passCheck = testutil.LogCheckArgs{}.WithExpPass(true)

		err = is.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")
	})

	Context("Direct precompile queries", func() {
		Context("balances query", func() {
			It("should return the correct balance", func() {
				balanceBefore, err := is.grpcHandler.GetBalance(sender.AccAddr, is.tokenDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")
				Expect(balanceBefore.Balance.Amount).To(Equal(sdk.NewInt(0)))
				Expect(balanceBefore.Balance.Denom).To(Equal(is.tokenDenom))

				is.mintAndSendXMPLCoin(is.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(sender.AccAddr, is.tokenDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(sdk.NewInt(balances[1].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
			})

			It("should return a single token balance", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(sender.AccAddr, utils.BaseDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(sdk.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
			})

			It("should return no balance for new account", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, evmosutiltx.GenerateAddress())
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(balances).To(BeEmpty())
			})

			It("should consume the correct amount of gas", func() {
				is.mintAndSendXMPLCoin(is.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				res, err := is.factory.ExecuteContractCall(sender.Priv, queryArgs, balancesArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				gasUsed := Max(bank.GasBalanceOf, len(balances)*bank.GasBalanceOf)
				// Here increasing the GasBalanceOf will increase the use of gas so they will never be equal
				Expect(gasUsed).To(BeNumerically("<=", ethRes.GasUsed))
			})
		})

		Context("totalSupply query", func() {
			It("should return the correct total supply", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.TotalSupplyMethod)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.TotalSupplyMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse evmos total supply")
				xmplTotalSupply := amount

				Expect(balances[0].Amount).To(Equal(evmosTotalSupply))
				Expect(balances[1].Amount).To(Equal(xmplTotalSupply))
			})
		})

		Context("supplyOf query", func() {
			It("should return the supply of Evmos", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, is.evmosAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse evmos total supply")

				Expect(out[0].(*big.Int)).To(Equal(evmosTotalSupply))
			})

			It("should return the supply of XMPL", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, is.xmplAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int)).To(Equal(amount))
			})

			It("should return a supply of 0 for a non existing token", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, evmosutiltx.GenerateAddress())
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int).Int64()).To(Equal(big.NewInt(0).Int64()))
			})

			It("should consume the correct amount of gas", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, is.xmplAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Here increasing the GasSupplyOf will increase the use of gas so they will never be equal
				Expect(bank.GasSupplyOf).To(BeNumerically("<=", ethRes.GasUsed))
			})
		})
	})

	Context("Calls from a contract", func() {
		const (
			BalancesFunction = "callBalances"
			TotalSupplyOf    = "callTotalSupply"
			SupplyOfFunction = "callSupplyOf"
		)

		Context("balances query", func() {
			It("should return the correct balance", func() {
				balanceBefore, err := is.grpcHandler.GetBalance(sender.AccAddr, is.tokenDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")
				Expect(balanceBefore.Balance.Amount).To(Equal(sdk.NewInt(0)))
				Expect(balanceBefore.Balance.Denom).To(Equal(is.tokenDenom))

				is.mintAndSendXMPLCoin(is.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, sender.Addr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(sender.AccAddr, is.tokenDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(sdk.NewInt(balances[1].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
			})

			It("should return a single token balance", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, sender.Addr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(sender.AccAddr, utils.BaseDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(sdk.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
			})

			It("should return no balance for new account", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, evmosutiltx.GenerateAddress())
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(balances).To(BeEmpty())
			})

			It("should consume the correct amount of gas", func() {
				is.mintAndSendXMPLCoin(is.keyring.GetAccAddr(0), sdk.NewInt(amount.Int64()))

				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, sender.Addr)
				res, err := is.factory.ExecuteContractCall(sender.Priv, queryArgs, balancesArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				gasUsed := Max(bank.GasBalanceOf, len(balances)*bank.GasBalanceOf)
				// Here increasing the GasBalanceOf will increase the use of gas so they will never be equal
				Expect(gasUsed).To(BeNumerically("<=", ethRes.GasUsed))
			})
		})

		Context("totalSupply query", func() {
			It("should return the correct total supply", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, TotalSupplyOf)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.TotalSupplyMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse evmos total supply")
				xmplTotalSupply := amount

				Expect(balances[0].Amount).To(Equal(evmosTotalSupply))
				Expect(balances[1].Amount).To(Equal(xmplTotalSupply))
			})
		})

		Context("supplyOf query", func() {
			It("should return the supply of Evmos", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, SupplyOfFunction, is.evmosAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				evmosTotalSupply, ok := new(big.Int).SetString("11000000000000000000", 10)
				Expect(ok).To(BeTrue(), "failed to parse evmos total supply")

				Expect(out[0].(*big.Int)).To(Equal(evmosTotalSupply))
			})

			It("should return the supply of XMPL", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, SupplyOfFunction, is.xmplAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int)).To(Equal(amount))
			})

			It("should return a supply of 0 for a non existing token", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, SupplyOfFunction, evmosutiltx.GenerateAddress())
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int).Int64()).To(Equal(big.NewInt(0).Int64()))
			})

			It("should consume the correct amount of gas", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, SupplyOfFunction, is.xmplAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// Here increasing the GasSupplyOf will increase the use of gas so they will never be equal
				Expect(bank.GasSupplyOf).To(BeNumerically("<=", ethRes.GasUsed))
			})
		})
	})
})
