package erc20_test

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/erc20/testdata"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v15/testutil/tx"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

// gs is the instance of the integration test suite used for testing the gas consumption
var gs *IntegrationTestSuite

func TestGasSuite(t *testing.T) {
	gs = new(IntegrationTestSuite)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "ERC20 Extension Gas Usage Test Suite")
}

// SetupGasTest runs the specific setup used for the gas usage tests.
func (is *IntegrationTestSuite) SetupGasTest() {
	keys := keyring.New(2)
	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	// Set up min deposit in Evmos
	params, err := gh.GetGovParams("deposit")
	Expect(err).ToNot(HaveOccurred(), "failed to get gov params")
	Expect(params).ToNot(BeNil(), "returned gov params are nil")

	updatedParams := params.Params
	updatedParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(nw.GetDenom(), sdk.NewInt(1e18)))
	err = nw.UpdateGovParams(*updatedParams)
	Expect(err).ToNot(HaveOccurred(), "failed to update the min deposit")

	is.network = nw
	is.factory = tf
	is.handler = gh
	is.keyring = keys

	is.tokenDenom = "xmpl"

	is.network.NextBlock()

	// NOTE: here we deploy an ERC20 contract and register it, which is then used to deploy the ERC20 precompile
	// for the registered token pair.
	regERC20Addr, err := is.factory.DeployContract(
		is.keyring.GetPrivKey(0),
		evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
		factory.ContractDeploymentData{
			Contract: contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{
				"Evmos", "EVMOS", uint8(18),
			},
		},
	)
	Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter burner contract")

	is.network.NextBlock()

	// Register ERC20 token pair for this test
	tokenPair, err := utils.RegisterERC20(is.factory, is.network, utils.ERC20RegistrationData{
		Address:      regERC20Addr,
		Denom:        "aevmos",
		ProposerPriv: is.keyring.GetPrivKey(0),
	})
	Expect(err).ToNot(HaveOccurred(), "failed to register ERC20 token")

	is.network.NextBlock()

	// Create precompile for registered token pair
	is.precompile, err = setupERC20PrecompileForTokenPair(is.network, tokenPair)
	Expect(err).ToNot(HaveOccurred(), "failed to setup ERC20 precompile")

	is.network.NextBlock()
}

var _ = Describe("ERC20 Extension Gas Tests - ", Ordered, func() {
	var contractsData ContractsData

	minVal := common.Big0
	maxVal := new(big.Int).Div(abi.MaxUint256, big.NewInt(10))
	tokenAmounts := getNExponentValuesBetween(minVal, maxVal, 10)
	usedGasTable := map[string]map[string]map[CallType]int64{}

	BeforeAll(func() {
		// FIXME: this is breaking the tests somehow??
		//gs.SetupGasTest()
		gs.SetupTest()

		err := gs.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to produce block")

		deployer := gs.keyring.GetKey(0)

		extCallerAddr, err := gs.factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20AllowanceCallerContract,
				// NOTE: we're passing the precompile address to the constructor because that initiates the contract
				// to make calls to the correct ERC20 precompile.
				ConstructorArgs: []interface{}{gs.precompile.Address()},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		erc20MinterBurnerAddr, err := gs.factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: contracts.ERC20MinterBurnerDecimalsContract,
				ConstructorArgs: []interface{}{
					"Xmpl", "XMPL", uint8(6),
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter burner contract")

		ERC20MinterV5Addr, err := gs.factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20MinterV5Contract,
				ConstructorArgs: []interface{}{
					"Xmpl", "XMPL",
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter contract")

		erc20MinterV5CallerAddr, err := gs.factory.DeployContract(
			deployer.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract: testdata.ERC20AllowanceCallerContract,
				ConstructorArgs: []interface{}{
					ERC20MinterV5Addr,
				},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy ERC20 minter caller contract")

		// Store the data of the deployed contracts
		contractsData = ContractsData{
			ownerPriv: deployer.Priv,
			contractData: map[CallType]ContractData{
				directCall: {
					Address: gs.precompile.Address(),
					ABI:     gs.precompile.ABI,
				},
				contractCall: {
					Address: extCallerAddr,
					ABI:     testdata.ERC20AllowanceCallerContract.ABI,
				},
				erc20Call: {
					Address: erc20MinterBurnerAddr,
					ABI:     contracts.ERC20MinterBurnerDecimalsContract.ABI,
				},
				erc20V5Call: {
					Address: ERC20MinterV5Addr,
					ABI:     testdata.ERC20MinterV5Contract.ABI,
				},
				erc20V5CallerCall: {
					Address: erc20MinterV5CallerAddr,
					ABI:     testdata.ERC20AllowanceCallerContract.ABI,
				},
			},
		}

	})

	for _, tokens := range tokenAmounts {
		tokens := tokens
		if tokens.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		Context(erc20.TransferMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should transfer %s tokens", tokens.String()), func(callType CallType) {
				sender := gs.keyring.GetKey(0)
				receiver := gs.keyring.GetKey(1)

				fmt.Println("Sending tokens: ", tokens.String())
				fundCoins := sdk.Coins{sdk.NewCoin(gs.tokenDenom, sdk.NewIntFromBigInt(tokens))}
				transferCoins := fundCoins

				gs.fundWithTokens(callType, contractsData, sender.Addr, fundCoins)

				gs.network.NextBlock()

				//gs.ExpectBalancesForContract(
				//	callType, contractsData,
				//	[]ExpectedBalance{
				//		{sender.AccAddr, fundCoins},
				//	},
				//)

				txArgs, transferArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferMethod,
					receiver.Addr, transferCoins[0].Amount.BigInt(),
				)

				res, err := gs.factory.ExecuteContractCall(sender.Priv, txArgs, transferArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, erc20.TransferMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				Entry(" - ERC20 v5 contract", erc20V5Call),
			)
		})

		Context(erc20.TransferFromMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf(" - it should transfer %s tokens from other account", tokens.String()), func(callType CallType) {
				owner := gs.keyring.GetKey(0)
				spender := gs.keyring.GetKey(1)
				receiverAddr := utiltx.GenerateAddress()

				fundCoins := sdk.Coins{sdk.NewCoin(gs.tokenDenom, sdk.NewIntFromBigInt(tokens))}
				transferCoins := fundCoins

				gs.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				// approve transfer
				txArgs, approveArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.ApproveMethod,
					spender.Addr, transferCoins[0].Amount.BigInt(),
				)
				_, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, approveArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				// execute transfer
				txArgs, transferArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					erc20.TransferFromMethod,
					owner.Addr, receiverAddr, transferCoins[0].Amount.BigInt(),
				)
				res, err := gs.factory.ExecuteContractCall(spender.Priv, txArgs, transferArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, erc20.TransferFromMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				Entry(" - ERC20 v5 contract", erc20V5Call),
			)
		})

		Context(auth.ApproveMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should approve %s tokens", tokens.String()), func(callType CallType) {
				owner := gs.keyring.GetKey(0)
				spender := gs.keyring.GetKey(1)

				fundCoins := sdk.Coins{sdk.NewCoin(gs.tokenDenom, sdk.NewIntFromBigInt(tokens))}
				transferCoins := fundCoins

				gs.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				txArgs, transferArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.ApproveMethod,
					spender.Addr, transferCoins[0].Amount.BigInt(),
				)
				res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, transferArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, auth.ApproveMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				Entry(" - ERC20 v5 contract", erc20V5Call),
			)
		})

		Context(auth.AllowanceMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should return %s allowance", tokens.String()), func(callType CallType) {
				owner := gs.keyring.GetKey(0)
				spender := gs.keyring.GetKey(1)

				fundCoins := sdk.Coins{sdk.NewCoin(gs.tokenDenom, sdk.NewIntFromBigInt(tokens))}
				transferCoins := fundCoins

				gs.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				// approve transfer
				txArgs, approveArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.ApproveMethod,
					spender.Addr, transferCoins[0].Amount.BigInt(),
				)
				_, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, approveArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				txArgs, allowanceArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.AllowanceMethod,
					owner.Addr, spender.Addr,
				)
				res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, allowanceArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, auth.AllowanceMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				Entry(" - ERC20 v5 contract", erc20V5Call),
			)
		})

		// FIXME: This is still failing
		//Context(erc20.NameMethod, Ordered, func() {
		//	DescribeTable(fmt.Sprintf("should return the name of the token"), func(callType CallType) {
		//		owner := gs.keyring.GetKey(0)
		//
		//		txArgs, nameArgs := gs.getTxAndCallArgs(
		//			callType, contractsData,
		//			erc20.NameMethod,
		//		)
		//		res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, nameArgs)
		//		Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
		//
		//		fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
		//		insertIntoGasTable(usedGasTable, callType, erc20.NameMethod, tokens.String(), res.GasUsed)
		//	},
		//		Entry(" - EVM extension", directCall),
		//		Entry(" - ERC20 contract", erc20Call),
		//		Entry(" - ERC20 v5 contract", erc20V5Call),
		//	)
		//})
		//
		//Context(erc20.SymbolMethod, Ordered, func() {
		//	DescribeTable(fmt.Sprintf("should return the symbol of the token"), func(callType CallType) {
		//		owner := gs.keyring.GetKey(0)
		//
		//		txArgs, symbolArgs := gs.getTxAndCallArgs(
		//			callType, contractsData,
		//			erc20.SymbolMethod,
		//		)
		//		res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, symbolArgs)
		//		Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
		//
		//		fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
		//		insertIntoGasTable(usedGasTable, callType, erc20.SymbolMethod, tokens.String(), res.GasUsed)
		//	},
		//		Entry(" - EVM extension", directCall),
		//		Entry(" - ERC20 contract", erc20Call),
		//		Entry(" - ERC20 v5 contract", erc20V5Call),
		//	)
		//})
		//
		//Context(erc20.DecimalsMethod, Ordered, func() {
		//	DescribeTable(fmt.Sprintf("should return the decimals of the token"), func(callType CallType) {
		//		owner := gs.keyring.GetKey(0)
		//
		//		txArgs, decimalsArgs := gs.getTxAndCallArgs(
		//			callType, contractsData,
		//			erc20.DecimalsMethod,
		//		)
		//		res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, decimalsArgs)
		//		Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")
		//
		//		fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
		//		insertIntoGasTable(usedGasTable, callType, erc20.DecimalsMethod, tokens.String(), res.GasUsed)
		//	},
		//		Entry(" - EVM extension", directCall),
		//		Entry(" - ERC20 contract", erc20Call),
		//		Entry(" - ERC20 v5 contract", erc20V5Call),
		//	)
		//})

		Context(erc20.TotalSupplyMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should return the total supply of the token"), func(callType CallType) {
				owner := gs.keyring.GetKey(0)

				txArgs, totalSupplyArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					erc20.TotalSupplyMethod,
				)
				res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, totalSupplyArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, erc20.TotalSupplyMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				Entry(" - ERC20 v5 contract", erc20V5Call),
			)
		})

		Context(erc20.BalanceOfMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should return %s token", tokens.String()), func(callType CallType) {
				owner := gs.keyring.GetKey(0)

				txArgs, balanceOfArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					erc20.BalanceOfMethod,
					owner.Addr,
				)
				res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, balanceOfArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, erc20.BalanceOfMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				Entry(" - ERC20 v5 contract", erc20V5Call),
			)
		})

		Context(auth.IncreaseAllowanceMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should increase allowance by %s tokens", tokens.String()), func(callType CallType) {
				owner := gs.keyring.GetKey(0)
				spender := gs.keyring.GetKey(1)

				fundCoins := sdk.Coins{sdk.NewCoin(gs.tokenDenom, sdk.NewIntFromBigInt(tokens))}
				transferCoins := fundCoins

				gs.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				txArgs, increaseAllowanceArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.IncreaseAllowanceMethod,
					spender.Addr, transferCoins[0].Amount.BigInt(),
				)
				res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, increaseAllowanceArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, auth.IncreaseAllowanceMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				// NOTE: The OpenZeppelin v5 contracts don't include this
			)
		})

		Context(auth.DecreaseAllowanceMethod, Ordered, func() {
			DescribeTable(fmt.Sprintf("should decrease allowance by %s tokens", tokens.String()), func(callType CallType) {
				owner := gs.keyring.GetKey(0)
				spender := gs.keyring.GetKey(1)

				fundCoins := sdk.Coins{sdk.NewCoin(gs.tokenDenom, sdk.NewIntFromBigInt(tokens))}
				transferCoins := fundCoins

				gs.fundWithTokens(callType, contractsData, owner.Addr, fundCoins)

				// approve transfer with sufficient amount before decreasing it afterwards
				txArgs, approveArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.ApproveMethod,
					spender.Addr, transferCoins[0].Amount.BigInt(),
				)
				_, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, approveArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				txArgs, decreaseAllowanceArgs := gs.getTxAndCallArgs(
					callType, contractsData,
					auth.DecreaseAllowanceMethod,
					spender.Addr, transferCoins[0].Amount.BigInt(),
				)
				res, err := gs.factory.ExecuteContractCall(owner.Priv, txArgs, decreaseAllowanceArgs)
				Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

				fmt.Printf("Adding gas consumption for: callType: %d, tokens: %s, gas used: %d", callType, tokens.String(), res.GasUsed)
				insertIntoGasTable(usedGasTable, callType, auth.DecreaseAllowanceMethod, tokens.String(), res.GasUsed)
			},
				Entry(" - EVM extension", directCall),
				Entry(" - ERC20 contract", erc20Call),
				// NOTE: The OpenZeppelin v5 contracts don't include this
			)
		})
	}

	AfterAll(func() {
		for transaction, entries := range usedGasTable {
			exportToFile(entries, fmt.Sprintf("erc20_%s.csv", transaction))
		}
	})
})

func exportToFile(entries map[string]map[CallType]int64, filename string) {
	// Create file
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write header
	//if _, err := f.WriteString("callType,tokens,gas\n"); err != nil {
	//	panic(err)
	//}

	callTypes := []CallType{directCall, erc20Call, erc20V5Call}

	// Write entries
	for sentTokens, gasConsumptions := range entries {
		line := fmt.Sprintf("%s", sentTokens)
		for _, ct := range callTypes {
			line += fmt.Sprintf(",%d", gasConsumptions[ct])
		}
		line += "\n"

		if _, err := f.WriteString(line); err != nil {
			panic(err)
		}
	}
}

func getNPointsBetween(min, max *big.Int, n int) []*big.Int {
	if min.Cmp(max) == 1 {
		panic("min cannot be greater than max")
	}

	if n < 2 {
		panic("n must be greater than 1")
	}

	points := make([]*big.Int, n)
	step := new(big.Int).Sub(max, min)
	step.Div(step, big.NewInt(int64(n-1)))

	for i := 0; i < n; i++ {
		points[i] = new(big.Int).Add(min, new(big.Int).Mul(step, big.NewInt(int64(i))))
	}

	return points
}

func getNExponentValuesBetween(min, max *big.Int, n int) []*big.Int {
	if min.Cmp(max) == 1 {
		panic("min cannot be greater than max")
	}

	if n < 2 {
		panic("n must be greater than 1")
	}

	lenDigitsMax := len(max.String())
	lenDigitsMin := len(min.String())

	// Get n points between min and max lengths
	points := getNPointsBetween(big.NewInt(int64(lenDigitsMin)), big.NewInt(int64(lenDigitsMax)), n)

	var numbers []*big.Int
	for _, point := range points {
		numberString := fmt.Sprintf("1%0*s", int(point.Int64()), "0")
		number, ok := new(big.Int).SetString(numberString, 10)
		if !ok {
			panic("could not convert string to big.Int")
		}

		numbers = append(numbers, number)
	}

	return numbers
}

func insertIntoGasTable(
	usedGas map[string]map[string]map[CallType]int64,
	callType CallType,
	transaction,
	tokens string,
	gas int64,
) {
	if _, ok := usedGas[transaction]; !ok {
		usedGas[transaction] = map[string]map[CallType]int64{}
	}

	if _, ok := usedGas[transaction][tokens]; !ok {
		usedGas[transaction][tokens] = map[CallType]int64{}
	}

	usedGas[transaction][tokens][callType] = gas
}
