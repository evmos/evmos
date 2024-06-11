package accesscontrol_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	accesscontrol "github.com/evmos/evmos/v18/precompiles/access_control"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	"testing"

	tokenfactory "github.com/evmos/evmos/v18/precompiles/token_factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

// PrecompileTestSuite is the implementation of the TestSuite interface for Token Factory precompile
// unit tests.
type PrecompileTestSuite struct {
	suite.Suite

	bondDenom   string
	tokenPair   erc20types.TokenPair
	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *accesscontrol.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	ctx := integrationNetwork.GetContext()
	sk := integrationNetwork.App.StakingKeeper
	bondDenom := sk.BondDenom(ctx)
	s.Require().NotEmpty(bondDenom, "bond denom cannot be empty")

	s.bondDenom = bondDenom
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = integrationNetwork

	// Create a token factory precompile
	tokenFactoryPrecompile, err := tokenfactory.NewPrecompile(
		s.network.App.AuthzKeeper,
		s.network.App.AccountKeeper,
		s.network.App.BankKeeper,
		s.network.App.EvmKeeper,
		s.network.App.Erc20Keeper,
		s.network.App.TransferKeeper,
		s.network.App.AccessControlKeeper,
	)
	s.Require().NoError(err)

	// Generate the address for the new ERC20 access control token
	precompileAddr := tokenFactoryPrecompile.Address()
	account := s.network.App.AccountKeeper.GetAccount(ctx, precompileAddr.Bytes())
	if account == nil {
		account = s.network.App.AccountKeeper.NewAccountWithAddress(ctx, precompileAddr.Bytes())
	}

	address := crypto.CreateAddress(tokenFactoryPrecompile.Address(), account.GetSequence())

	addrHex := address.String()
	denom := erc20types.CreateDenom(addrHex)

	tokenPair := erc20types.TokenPair{
		Erc20Address:  addrHex,
		Denom:         denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_EXTERNAL,
	}

	s.tokenPair = tokenPair

	// Save the token pair in the store
	s.network.App.Erc20Keeper.SetTokenPair(ctx, tokenPair)
	s.network.App.Erc20Keeper.SetDenomMap(ctx, tokenPair.Denom, tokenPair.GetID())
	s.network.App.Erc20Keeper.SetERC20Map(ctx, common.HexToAddress(tokenPair.Erc20Address), tokenPair.GetID())

	precompile, err := accesscontrol.NewPrecompile(
		tokenPair,
		s.network.App.BankKeeper,
		s.network.App.AuthzKeeper,
		s.network.App.TransferKeeper,
		s.network.App.AccessControlKeeper,
	)

	s.Require().NoError(err)
	s.precompile = precompile
}
