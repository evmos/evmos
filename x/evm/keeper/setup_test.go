package keeper_test

import (
	"math"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/config"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	handler grpc.Handler
	keyring keyring.Keyring
	factory factory.TxFactory

	enableFeemarket  bool
	enableLondonHF   bool
	mintFeeCollector bool
}

type UnitTestSuite struct {
	suite.Suite
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	s.enableFeemarket = false
	s.enableLondonHF = true
	suite.Run(t, s)

	// Run UnitTestSuite
	unitTestSuite := new(UnitTestSuite)
	suite.Run(t, unitTestSuite)
}

func (suite *KeeperTestSuite) SetupTest() {
	keys := keyring.New(2)
	// Set custom balance based on test params
	customGenesis := network.CustomGenesisState{}
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	if s.enableFeemarket {
		feemarketGenesis.Params.EnableHeight = 1
		feemarketGenesis.Params.NoBaseFee = false
	} else {
		feemarketGenesis.Params.NoBaseFee = true
	}
	customGenesis[feemarkettypes.ModuleName] = feemarketGenesis

	if s.mintFeeCollector {
		// mint some coin to fee collector
		coins := sdk.NewCoins(sdk.NewCoin(config.GetDenom(), sdkmath.NewInt(int64(params.TxGas)-1)))
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

	if !s.enableLondonHF {
		eip155ChainID, err := evmostypes.ParseChainID(suite.network.GetChainID())
		suite.Require().NoError(err)
		chainConfig := config.DefaultChainConfig(eip155ChainID)
		maxInt := sdkmath.NewInt(math.MaxInt64)
		chainConfig.LondonBlock = maxInt.BigInt()
		chainConfig.ArrowGlacierBlock = maxInt.BigInt()
		chainConfig.GrayGlacierBlock = maxInt.BigInt()
		chainConfig.MergeNetsplitBlock = maxInt.BigInt()
		chainConfig.ShanghaiBlock = maxInt.BigInt()
		chainConfig.CancunBlock = maxInt.BigInt()
		err = config.NewEVMConfigurator().
			WithChainConfig(chainConfig).
			Configure(suite.network.GetChainID())
		suite.Require().NoError(err)
	} else {
		err := config.NewEVMConfigurator().
			Configure(suite.network.GetChainID())
		suite.Require().NoError(err)
	}
}
