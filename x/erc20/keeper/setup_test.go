package keeper_test

import (
	"testing"

	//nolint:revive // dot imports are fine for Ginkgo
	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

type KeeperTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory

	queryClient types.QueryClient

	mintFeeCollector bool
}

func TestKeeperUnitTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	keys := keyring.New(2)
	// Set custom balance based on test params
	customGenesis := network.CustomGenesisState{}

	if s.mintFeeCollector {
		// mint some coin to fee collector
		coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(int64(params.TxGas)-1)))
		balances := []banktypes.Balance{
			{
				Address: authtypes.NewModuleAddress(authtypes.FeeCollectorName).String(),
				Coins:   coins,
			},
		}
		bankGenesis := banktypes.DefaultGenesisState()
		bankGenesis.Balances = balances
		customGenesis[banktypes.ModuleName] = bankGenesis
	}

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGenesis),
	)
	gh := grpc.NewIntegrationHandler(nw)
	tf := factory.New(nw, gh)

	s.network = nw
	s.factory = tf
	s.handler = gh
	s.keyring = keys
	s.queryClient = nw.GetERC20Client()
}
