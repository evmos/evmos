package bank_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v20/precompiles/bank"
	"github.com/evmos/evmos/v20/precompiles/bank/testdata"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	testutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v20/x/inflation/v1/types"

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

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     keyring.Keyring

	precompile *bank.Precompile
}

func (is *IntegrationTestSuite) SetupTest() {
	// Mint and register a second coin for testing purposes
	// FIXME the RegisterCoin logic will need to be refactored
	// once logic is integrated
	// with the protocol via genesis and/or a transaction
	is.tokenDenom = xmplDenom
	keyring := keyring.New(2)
	genesis := testutils.CreateGenesisWithTokenPairs(keyring)

	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithOtherDenoms([]string{is.tokenDenom}), // set some funds of other denom to the prefunded accounts
		network.WithCustomGenesis(genesis),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	ctx := integrationNetwork.GetContext()
	sk := integrationNetwork.App.StakingKeeper
	bondDenom, err := sk.BondDenom(ctx)
	Expect(err).ToNot(HaveOccurred())
	Expect(bondDenom).ToNot(BeEmpty(), "bond denom cannot be empty")

	is.bondDenom = bondDenom
	is.factory = txFactory
	is.grpcHandler = grpcHandler
	is.keyring = keyring
	is.network = integrationNetwork

	tokenPairID := is.network.App.Erc20Keeper.GetTokenPairID(is.network.GetContext(), is.bondDenom)
	tokenPair, found := is.network.App.Erc20Keeper.GetTokenPair(is.network.GetContext(), tokenPairID)
	Expect(found).To(BeTrue(), "failed to register token erc20 extension")
	is.evmosAddr = common.HexToAddress(tokenPair.Erc20Address)

	// Mint and register a second coin for testing purposes
	err = is.network.App.BankKeeper.MintCoins(is.network.GetContext(), inflationtypes.ModuleName, sdk.Coins{{Denom: is.tokenDenom, Amount: math.NewInt(1e18)}})
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

		evmosTotalSupply, _ = new(big.Int).SetString("200003000000000000000000", 10)
		xmplTotalSupply, _  = new(big.Int).SetString("200000000000000000000000", 10)
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
				// New account with 0 balances (does not exist on the chain yet)
				receiver := utiltx.GenerateAddress()

				err := is.factory.FundAccount(sender, receiver.Bytes(), sdk.NewCoins(sdk.NewCoin(is.tokenDenom, math.NewIntFromBigInt(amount))))
				Expect(err).ToNot(HaveOccurred(), "error while funding account")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, receiver)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(receiver.Bytes(), is.tokenDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(math.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
				Expect(*balances[0].Amount).To(Equal(*amount))
			})

			It("should return a single token balance", func() {
				// New account with 0 balances (does not exist on the chain yet)
				receiver := utiltx.GenerateAddress()

				err := testutils.FundAccountWithBaseDenom(is.factory, is.network, sender, receiver.Bytes(), math.NewIntFromBigInt(amount))
				Expect(err).ToNot(HaveOccurred(), "error while funding account")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, receiver)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(receiver.Bytes(), is.network.GetDenom())
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(math.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
				Expect(*balances[0].Amount).To(Equal(*amount))
			})

			It("should return no balance for new account", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, utiltx.GenerateAddress())
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(balances).To(BeEmpty())
			})

			It("should consume the correct amount of gas", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(directCall, contractData, bank.BalancesMethod, sender.Addr)
				res, err := is.factory.ExecuteContractCall(sender.Priv, queryArgs, balancesArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				gasUsed := Max(bank.GasBalances, len(balances)*bank.GasBalances)
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

				Expect(out[0].(*big.Int)).To(Equal(evmosTotalSupply))
			})

			It("should return the supply of XMPL", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, is.xmplAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int)).To(Equal(xmplTotalSupply))
			})

			It("should return a supply of 0 for a non existing token", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(directCall, contractData, bank.SupplyOfMethod, utiltx.GenerateAddress())
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
				receiver := utiltx.GenerateAddress()

				err := is.factory.FundAccount(sender, receiver.Bytes(), sdk.NewCoins(sdk.NewCoin(is.tokenDenom, math.NewIntFromBigInt(amount))))
				Expect(err).ToNot(HaveOccurred(), "error while funding account")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, receiver)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(receiver.Bytes(), is.tokenDenom)
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(math.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
				Expect(*balances[0].Amount).To(Equal(*amount))
			})

			It("should return a single token balance", func() {
				// New account with 0 balances (does not exist on the chain yet)
				receiver := utiltx.GenerateAddress()

				err := testutils.FundAccountWithBaseDenom(is.factory, is.network, sender, receiver.Bytes(), math.NewIntFromBigInt(amount))
				Expect(err).ToNot(HaveOccurred(), "error while funding account")
				Expect(is.network.NextBlock()).ToNot(HaveOccurred(), "error on NextBlock")

				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, receiver)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				balanceAfter, err := is.grpcHandler.GetBalance(receiver.Bytes(), is.network.GetDenom())
				Expect(err).ToNot(HaveOccurred(), "failed to get balance")

				Expect(math.NewInt(balances[0].Amount.Int64())).To(Equal(balanceAfter.Balance.Amount))
				Expect(*balances[0].Amount).To(Equal(*amount))
			})

			It("should return no balance for new account", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, utiltx.GenerateAddress())
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, balancesArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(balances).To(BeEmpty())
			})

			It("should consume the correct amount of gas", func() {
				queryArgs, balancesArgs := getTxAndCallArgs(contractCall, contractData, BalancesFunction, sender.Addr)
				res, err := is.factory.ExecuteContractCall(sender.Priv, queryArgs, balancesArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				ethRes, err := evmtypes.DecodeTxResponse(res.Data)
				Expect(err).ToNot(HaveOccurred(), "failed to decode tx response")

				var balances []bank.Balance
				err = is.precompile.UnpackIntoInterface(&balances, bank.BalancesMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				gasUsed := Max(bank.GasBalances, len(balances)*bank.GasBalances)
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

				Expect(out[0].(*big.Int)).To(Equal(evmosTotalSupply))
			})

			It("should return the supply of XMPL", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, SupplyOfFunction, is.xmplAddr)
				_, ethRes, err := is.factory.CallContractAndCheckLogs(sender.Priv, queryArgs, supplyArgs, passCheck)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				out, err := is.precompile.Unpack(bank.SupplyOfMethod, ethRes.Ret)
				Expect(err).ToNot(HaveOccurred(), "failed to unpack balances")

				Expect(out[0].(*big.Int)).To(Equal(xmplTotalSupply))
			})

			It("should return a supply of 0 for a non existing token", func() {
				queryArgs, supplyArgs := getTxAndCallArgs(contractCall, contractData, SupplyOfFunction, utiltx.GenerateAddress())
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
