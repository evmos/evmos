package erc20_test

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/contracts"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

type fuzzSuite struct {
	network     *network.IntegrationNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     keyring.Keyring
}

func newFuzzSuite() *fuzzSuite {
	keyring := keyring.New(2)
	fuzzNetwork := network.New(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(fuzzNetwork)
	factory := factory.New(fuzzNetwork, grpcHandler)

	return &fuzzSuite{
		network:     fuzzNetwork,
		factory:     factory,
		grpcHandler: grpcHandler,
		keyring:     keyring,
	}
}

func TransferERC20(
	suite *fuzzSuite,
	contractAddr common.Address,
	to common.Address,
	fromKey keyring.Key,
	amount uint64,
) (abcitypes.ResponseDeliverTx, error) {
	txArgs := evmtypes.EvmTxArgs{To: &contractAddr}
	transferArgs := factory.CallArgs{
		ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
		MethodName:  erc20.TransferMethod,
		Args:        []interface{}{to, big.NewInt(int64(amount))},
	}
	transferCheck := testutil.LogCheckArgs{
		ABIEvents: contracts.ERC20MinterBurnerDecimalsContract.ABI.Events,
		ExpEvents: []string{erc20.EventTypeTransfer},
		ExpPass:   true,
	}

	res, _, err := suite.factory.CallContractAndCheckLogs(
		fromKey.Priv,
		txArgs,
		transferArgs,
		transferCheck,
	)
	if err != nil {
		return abcitypes.ResponseDeliverTx{}, err
	}

	return res, nil
}

func FuzzERC20Transfer(f *testing.F) {
	suite := newFuzzSuite()

	erc20MinterBurnerAddr, err := suite.factory.DeployContract(
		suite.keyring.GetPrivKey(0),
		evmtypes.EvmTxArgs{},
		factory.ContractDeploymentData{
			Contract: contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{
				"Xmpl", "XMPL", uint8(18),
			},
		},
	)
	if err != nil {
		f.Fatal(err)
	}

	err = suite.network.NextBlock()
	if err != nil {
		f.Fatal(err)
	}

	// Mint some tokens to the first account
	txArgs := evmtypes.EvmTxArgs{To: &erc20MinterBurnerAddr}
	mintArgs := factory.CallArgs{
		ContractABI: contracts.ERC20MinterBurnerDecimalsContract.ABI,
		MethodName:  "mint",
		Args:        []interface{}{suite.keyring.GetAddr(0), abi.MaxUint256},
	}
	mintCheck := testutil.LogCheckArgs{
		ABIEvents: contracts.ERC20MinterBurnerDecimalsContract.ABI.Events,
		ExpEvents: []string{erc20.EventTypeTransfer},
		ExpPass:   true,
	}

	_, _, err = suite.factory.CallContractAndCheckLogs(
		suite.keyring.GetPrivKey(0),
		txArgs,
		mintArgs,
		mintCheck,
	)
	if err != nil {
		f.Fatal(err)
	}

	// Open file to write used gas to
	file, err := os.Create("gas_used.csv")
	if err != nil {
		f.Fatal(err)
	}
	defer file.Close()

	file.WriteString("amount, gas used\n")

	f.Add(uint64(1e6))
	f.Add(uint64(1e9))
	f.Add(uint64(1e12))
	f.Add(uint64(1e15))
	f.Add(uint64(1e18))

	f.Fuzz(func(t *testing.T, amount uint64) {
		res, err := TransferERC20(
			suite,
			erc20MinterBurnerAddr,
			suite.keyring.GetAddr(1),
			suite.keyring.GetKey(0),
			amount,
		)
		if err != nil {
			t.Fatal(err)
		}

		file.WriteString(fmt.Sprintf("| %d:%d  |\n", amount, res.GasUsed))
	})
}
