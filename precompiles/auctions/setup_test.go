package auctions_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/precompiles/auctions"
	"github.com/evmos/evmos/v20/precompiles/erc20"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"

	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"

	"github.com/stretchr/testify/suite"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *auctions.Precompile
	tokenPair  erc20types.TokenPair
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)

	// Set up custom genesis state if needed
	customGen := network.CustomGenesisState{}
	// Add any auction-specific genesis setup here

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGen),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = nw

	// Create a token pair for uatom for the auction precompile
	tokenPair := erc20types.NewTokenPair(utiltx.GenerateAddress(), "uatom", erc20types.OWNER_MODULE)
	s.network.App.Erc20Keeper.SetTokenPair(s.network.GetContext(), tokenPair)
	s.network.App.Erc20Keeper.SetERC20Map(s.network.GetContext(), tokenPair.GetERC20Contract(), tokenPair.GetID())
	erc20Precompile, err := erc20.NewPrecompile(
		tokenPair,
		s.network.App.BankKeeper,
		s.network.App.AuthzKeeper,
		s.network.App.TransferKeeper,
	)
	s.Require().NoError(err, "failed to create %q erc20 precompile", tokenPair.Denom)
	err = s.network.App.Erc20Keeper.EnableDynamicPrecompiles(
		s.network.GetContext(),
		erc20Precompile.Address(),
	)
	s.Require().NoError(err, "failed to add %q erc20 precompile to EVM extensions", tokenPair.Denom)
	s.tokenPair = tokenPair

	err = s.network.FundAccount(s.keyring.GetAddr(0).Bytes(), types.NewCoins(types.NewCoin("uatom", sdkmath.NewInt(2e18))))
	s.Require().NoError(err)

	if s.precompile, err = auctions.NewPrecompile(
		s.network.App.AuctionsKeeper,
		s.network.App.Erc20Keeper,
		s.network.App.AuthzKeeper,
	); err != nil {
		panic(err)
	}
}
