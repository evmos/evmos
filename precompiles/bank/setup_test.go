package bank_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/bank"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	integrationUtils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

// PrecompileTestSuite is the implementation of the TestSuite interface for ERC20 precompile
// unit tests.
type PrecompileTestSuite struct {
	suite.Suite

	bondDenom, tokenDenom string
	evmosAddr, xmplAddr   common.Address

	// tokenDenom is the specific token denomination used in testing the ERC20 precompile.
	// This denomination is used to instantiate the precompile.
	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *bank.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := keyring.New(2)
	genesis := integrationUtils.CreateGenesisWithTokenPairs(keyring)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(genesis),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	ctx := integrationNetwork.GetContext()
	sk := integrationNetwork.App.StakingKeeper
	bondDenom := sk.BondDenom(ctx)
	s.Require().NotEmpty(bondDenom, "bond denom cannot be empty")

	s.bondDenom = bondDenom
	s.tokenDenom = "xmpl"
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = integrationNetwork

	tokenPairId := s.network.App.Erc20Keeper.GetTokenPairID(s.network.GetContext(), s.bondDenom)
	tokenPair, found := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), tokenPairId)
	s.Require().True(found)
	s.evmosAddr = common.HexToAddress(tokenPair.Erc20Address)

	s.evmosAddr = tokenPair.GetERC20Contract()

	// Mint and register a second coin for testing purposes
	err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, sdk.Coins{{Denom: "xmpl", Amount: math.NewInt(1e18)}})
	s.Require().NoError(err)

	tokenPairId = s.network.App.Erc20Keeper.GetTokenPairID(s.network.GetContext(), s.tokenDenom)
	tokenPair, found = s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), tokenPairId)
	s.Require().True(found)
	s.xmplAddr = common.HexToAddress(tokenPair.Erc20Address)

	s.xmplAddr = tokenPair.GetERC20Contract()

	s.precompile = s.setupBankPrecompile()
}
