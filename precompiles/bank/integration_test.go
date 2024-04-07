package bank_test

import (
	"github.com/ethereum/go-ethereum/core/vm"
	compiledcontracts "github.com/evmos/evmos/v16/contracts"
	"github.com/evmos/evmos/v16/precompiles/erc20"
	"math/big"
	"testing"

	"github.com/evmos/evmos/v16/precompiles/bank/testdata"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"

	evmosutiltx "github.com/evmos/evmos/v16/testutil/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v16/precompiles/bank"

	"github.com/evmos/evmos/v16/precompiles/testutil"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
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
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
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

	// Register EVMOS
	evmosMetadata, found := is.network.App.BankKeeper.GetDenomMetaData(is.network.GetContext(), is.bondDenom)
	Expect(found).To(BeTrue(), "failed to get denom metadata")

	tokenPair, err := is.network.App.Erc20Keeper.RegisterCoin(is.network.GetContext(), evmosMetadata)
	Expect(err).ToNot(HaveOccurred(), "failed to register coin")

	is.evmosAddr = common.HexToAddress(tokenPair.Erc20Address)

	// Mint and register a second coin for testing purposes
	err = is.network.App.BankKeeper.MintCoins(is.network.GetContext(), inflationtypes.ModuleName, sdk.Coins{{Denom: is.tokenDenom, Amount: sdk.NewInt(1e18)}})
	Expect(err).ToNot(HaveOccurred(), "failed to mint coin")

	xmplMetadata := banktypes.Metadata{
		Description: "An exemplary token",
		Base:        is.tokenDenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    is.tokenDenom,
				Exponent: 0,
				Aliases:  []string{is.tokenDenom},
			},
			{
				Denom:    is.tokenDenom,
				Exponent: 18,
			},
		},
		Name:    "Exemplary",
		Symbol:  "XMPL",
		Display: is.tokenDenom,
	}

	tokenPair, err = is.network.App.Erc20Keeper.RegisterCoin(is.network.GetContext(), xmplMetadata)
	Expect(err).ToNot(HaveOccurred(), "failed to register coin")

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
		err                    error
		sender                 keyring.Key
		amount                 *big.Int

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

		bankCallerContractAddr, err = is.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.BankCallerContract,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter burner contract")

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			precompileAddr: is.precompile.Address(),
			precompileABI:  is.precompile.ABI,
			contractAddr:   bankCallerContractAddr,
			contractABI:    testdata.BankCallerContract.ABI,
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

// The following tests are used to check that when batching multiple state changing transactions
// in one block, both states (Cosmos and EVM) are updated or reverted correctly.
//
// For this purpose, we are deploying an ERC20 contract and then calling a method
// on the BankCaller contract which transfers an ERC20 balance is sent between accounts as well as
// an interaction with the bank precompile is made.
//
// TODO: There are currently no state changes associated with the bank precompile, only queries
var _ = Describe("Bank - batching cosmos and eth interactions", func() {
	const (
		erc20Name     = "Test"
		erc20Token    = "TTT"
		erc20Decimals = uint8(18)
	)

	var (
		// bankCallerAddr is the address of the deployed BankCaller contract
		bankCallerAddr common.Address
		// erc20ContractAddr is the address of the deployed ERC20 contract
		erc20ContractAddr common.Address

		// mintAmount is the amount of ERC20 tokens minted to the BankCaller contract
		mintAmount = big.NewInt(1e18)
		// transferredAmount is the amount of ERC20 tokens to transfer during the tests
		transferredAmount = big.NewInt(1234e9)
	)

	BeforeEach(func() {
		is.SetupTest()
		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		deployer := is.keyring.GetKey(0)

		// Deploy BankCaller contract
		var err error
		bankCallerAddr, err = is.factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.BankCallerContract,
			},
		)
		Expect(err).To(BeNil(), "error while deploying the BankCaller contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		// Deploy ERC20 contract
		erc20ContractAddr, err = is.factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: compiledcontracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{
					erc20Name, erc20Token, erc20Decimals,
				},
			},
		)
		Expect(err).To(BeNil(), "error while deploying the ERC20 contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		// Mint tokens to the BankCaller contract
		_, _, err = is.factory.CallContractAndCheckLogs(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20ContractAddr},
			factory.CallArgs{
				ContractABI: compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "mint",
				Args:        []interface{}{bankCallerAddr, mintAmount},
			},
			testutil.LogCheckArgs{
				ABIEvents: compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI.Events,
				ExpEvents: []string{erc20.EventTypeTransfer}, // minting produces a Transfer event
				ExpPass:   true,
			},
		)
		Expect(err).To(BeNil(), "error while minting tokens to the BankCaller contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		// Check that the balance is zero before minting some tokens
		//
		// TODO: should be done with EthCall instead but haven't implemented it yet
		res, err := is.factory.ExecuteContractCall(
			deployer.Priv,
			evmtypes.EvmTxArgs{To: &erc20ContractAddr},
			factory.CallArgs{
				ContractABI: compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "balanceOf",
				Args:        []interface{}{bankCallerAddr},
			},
		)
		Expect(err).To(BeNil(), "error while getting ERC20 balance for the BankCaller contract")
		Expect(res.IsOK()).To(BeTrue(), "expected successful call to ERC20 contract")

		evmRes, err := evmtypes.DecodeTxResponse(res.Data)
		Expect(err).To(BeNil(), "error while decoding EVM response")
		Expect(evmRes.Ret).ToNot(HaveLen(0), "expected non-empty return data")

		var balance *big.Int
		err = compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI.UnpackIntoInterface(
			&balance, "balanceOf", evmRes.Ret,
		)
		Expect(err).To(BeNil(), "error while unpacking ERC20 balance before minting")
		Expect(balance.String()).To(Equal(mintAmount.String()), "expected different ERC20 balance for the BankCaller contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")
	})

	It("should revert the state correctly", func() {
		sender := is.keyring.GetKey(0)

		// NOTE: trying to transfer more than the balance of the contract should fail
		moreThanBalance := new(big.Int).Add(mintAmount, big.NewInt(1))

		_, _, err := is.factory.CallContractAndCheckLogs(
			sender.Priv,
			evmtypes.EvmTxArgs{To: &bankCallerAddr},
			factory.CallArgs{
				ContractABI: testdata.BankCallerContract.ABI,
				MethodName:  "callERC20AndRunQueries",
				Args:        []interface{}{erc20ContractAddr, sender.Addr, moreThanBalance},
			},
			testutil.LogCheckArgs{
				ExpPass:     false,
				ErrContains: vm.ErrExecutionReverted.Error(),
			},
		)
		Expect(err).ToNot(HaveOccurred(), "expected contract call to be reverted")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		// Check that the balance remains unchanged, as we are expecting the transaction to be reverted
		//
		// TODO: should be done with EthCall instead but haven't implemented it yet
		_, evmRes, err := is.factory.CallContractAndCheckLogs(
			sender.Priv,
			evmtypes.EvmTxArgs{To: &erc20ContractAddr},
			factory.CallArgs{
				ContractABI: compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "balanceOf",
				Args:        []interface{}{bankCallerAddr},
			},
			testutil.LogCheckArgs{
				ExpPass: true,
			},
		)
		Expect(err).To(BeNil(), "error while getting ERC20 balance for the BankCaller contract")

		var balance *big.Int
		err = compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI.UnpackIntoInterface(
			&balance, "balanceOf", evmRes.Ret,
		)
		Expect(err).To(BeNil(), "error while unpacking ERC20 balance before minting")
		Expect(balance.String()).To(Equal(mintAmount.String()), "expected different ERC20 balance for the BankCaller contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")
	})

	It("should persist the state changes correctly", func() {
		sender := is.keyring.GetKey(0)

		_, _, err := is.factory.CallContractAndCheckLogs(
			sender.Priv,
			evmtypes.EvmTxArgs{To: &bankCallerAddr},
			factory.CallArgs{
				ContractABI: testdata.BankCallerContract.ABI,
				MethodName:  "callERC20AndRunQueries",
				Args:        []interface{}{erc20ContractAddr, sender.Addr, transferredAmount},
			},
			testutil.LogCheckArgs{
				ABIEvents: compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI.Events,
				ExpEvents: []string{erc20.EventTypeTransfer},
				ExpPass:   true,
			},
		)
		Expect(err).ToNot(HaveOccurred(), "unexpected results calling the smart contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		// Check that the ERC20 balance of the BankCaller contract is updated
		//
		// TODO: should be done with EthCall instead but haven't implemented it yet
		_, evmRes, err := is.factory.CallContractAndCheckLogs(
			sender.Priv,
			evmtypes.EvmTxArgs{To: &erc20ContractAddr},
			factory.CallArgs{
				ContractABI: compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI,
				MethodName:  "balanceOf",
				Args:        []interface{}{bankCallerAddr},
			},
			testutil.LogCheckArgs{
				ExpPass: true,
			},
		)
		Expect(err).To(BeNil(), "error while getting ERC20 balance for the BankCaller contract")

		Expect(is.network.NextBlock()).To(BeNil(), "error while advancing block")

		var erc20Balance *big.Int
		err = compiledcontracts.ERC20MinterBurnerDecimalsContract.ABI.UnpackIntoInterface(
			&erc20Balance, "balanceOf", evmRes.Ret,
		)
		Expect(err).To(BeNil(), "error while unpacking ERC20 balance")
		expectedBalance := new(big.Int).Sub(mintAmount, transferredAmount)
		Expect(erc20Balance.String()).To(Equal(expectedBalance.String()),
			"expected different ERC20 balance for the BankCaller contract",
		)
	})
})
