package v17_test

import (
	"fmt"
	"os"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	v17 "github.com/evmos/evmos/v16/app/upgrades/v17"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

type Unmarshaler interface {
	Unmarshal(dAtA []byte) error
}

func getCustomGenesisState[K Unmarshaler](genState K, path string) (K, error) {
	out, err := os.ReadFile(path)
	if err != nil {
		return genState, err
	}
	err = genState.Unmarshal(out)
	if err != nil {
		return genState, err
	}
	return genState, nil
}

func generateCustomGenesisState() (network.CustomGenesisState, error) {
	evmGenState := &evmtypes.GenesisState{}
	evmGenState, err := getCustomGenesisState(evmGenState, "./genesis_files/evm_gen_state.json")
	if err != nil {
		panic(err)
	}

	authGenState := &authtypes.GenesisState{}
	authGenState, err = getCustomGenesisState(authGenState, "./genesis_files/auth_gen_state.json")
	if err != nil {
		panic(err)
	}

	bankGenState := &banktypes.GenesisState{}
	bankGenState, err = getCustomGenesisState(bankGenState, "./genesis_files/bank_gen_state.json")
	if err != nil {
		panic(err)
	}

	erc20State := &erc20types.GenesisState{}
	erc20State, err = getCustomGenesisState(erc20State, "./genesis_files/erc20_gen_state.json")
	if err != nil {
		panic(err)
	}

	return network.CustomGenesisState{
		evmtypes.ModuleName:   evmGenState,
		authtypes.ModuleName:  authGenState,
		erc20types.ModuleName: erc20State,
		banktypes.ModuleName:  bankGenState,
	}, nil
}

type benchmarkSuite struct {
	network     *network.UnitTestNetwork
	grpcHandler grpc.Handler
	txFactory   factory.TxFactory
	keyring     testkeyring.Keyring
}

func BenchmarkShittyMigration(b *testing.B) {
	// Reset chain on every tx type to have a clean state
	// and a fair benchmark
	b.StopTimer()
	keyring := testkeyring.New(3)

	// Custom genesis state to add erc20 token pairs for dynamic precompiles
	customGenesisState, err := generateCustomGenesisState()
	if err != nil {
		panic(err)
	}

	// Because we are not going thorugh the ante handler,
	// we need to configure the context to execution mode
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesisState),
	)

	b.Run(fmt.Sprintf("killer_benchmark"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			// FUNCTION CALL
			err := v17.RunSTRv2Migration(
				unitNetwork.GetContext(),
				unitNetwork.GetContext().Logger(),
				unitNetwork.App.AccountKeeper,
				unitNetwork.App.BankKeeper,
				unitNetwork.App.Erc20Keeper,
				unitNetwork.App.EvmKeeper,
				common.HexToAddress("0xD0E2fE4eBBB7ECd3Acc719Ac45B70ade6bad024d"),
				unitNetwork.GetDenom(),
			)
			b.StopTimer()

			if err != nil {
				panic(err)
			}
		}
	})
}
