// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	commonnetwork "github.com/evmos/evmos/v15/testutil/integration/common/network"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/testutil/integration/ibc/chain"
	"github.com/evmos/evmos/v15/testutil/integration/ibc/coordinator"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/types"
	"github.com/stretchr/testify/suite"
)

const (
	portID    = "transfer"
	channelID = "channel-0"
)

const (
	TokenToMint = 1e18
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	stateDB *statedb.StateDB

	precompile *osmosis.Precompile

	coordinator *coordinator.IntegrationCoordinator
	chainA      chain.Chain
	chainB      chain.Chain
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	txFactory := factory.New(unitNetwork, grpcHandler)

	headerHash := unitNetwork.GetContext().HeaderHash()
	stateDB := statedb.New(
		unitNetwork.GetContext(),
		unitNetwork.App.EvmKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash.Bytes())),
	)

	precompile, err := osmosis.NewPrecompile(
		portID,
		channelID,
		osmosis.XCSContract,
		unitNetwork.App.BankKeeper,
		unitNetwork.App.TransferKeeper,
		unitNetwork.App.StakingKeeper,
		unitNetwork.App.Erc20Keeper,
	)
	s.Require().NoError(err)

	s.stateDB = stateDB
	s.network = unitNetwork
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile

	coordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{unitNetwork},
	)

	chainA := coordinator.GetChain(unitNetwork.GetChainID()).(*ibctesting.TestChain)
	chainB := coordinator.GetChain(ibctesting.GetChainID(2)).(*ibctesting.TestChain)
	path := coordinator.NewTransferPath(chainA, chainB)
	coordinator.Setup(path)

	s.chainA = chainA
	s.chainB = chainB

	s.registerEvmosERC20Coins()
	s.registerOsmoERC20Coins()
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) registerEvmosERC20Coins() {
	bondDenom := s.network.App.StakingKeeper.BondDenom(s.network.GetContext())
	evmosMetadata := banktypes.Metadata{
		Name:        "Evmos token",
		Symbol:      "EVMOS",
		Description: "The native token of Evmos",
		Base:        bondDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    bondDenom,
				Exponent: 0,
				Aliases:  []string{"aevmos"},
			},
			{
				Denom:    "aevmos",
				Exponent: 18,
			},
		},
		Display: "evmos",
	}

	s.T().Log("Before minting evmos...")
	amount := sdk.NewInt(TokenToMint)
	coin := sdk.NewCoin(evmosMetadata.Base, amount)
	err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, sdk.NewCoins(coin))
	s.Require().NoError(err)

	s.T().Log("Before registering evmos...")
	// Register Evmos Token Pair.
	_, err = s.network.App.Erc20Keeper.RegisterCoin(s.network.GetContext(), evmosMetadata)
	s.Require().NoError(err)

	s.T().Log("Before sending evmos...")
	sender := s.keyring.GetAccAddr(0)
	err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.network.GetContext(),
		inflationtypes.ModuleName,
		sender,
		sdk.NewCoins(coin),
	)
	s.Require().NoError(err)

	s.T().Log("Before getting denom evmos...")
	// Check that token has been registered correctly.
	evmosDenomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), bondDenom)
	_, ok := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), evmosDenomID)
	s.Require().True(ok, "expected evmos token pair to be found")
}

// registerERC20Coins registers Evmos and IBC OSMO coin as an ERC20 token
func (s *PrecompileTestSuite) registerOsmoERC20Coins() {
	// Register EVMOS ERC20 equivalent Register IBC OSMO Token Pair
	denomTrace := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", portID, channelID),
		BaseDenom: osmosis.OsmosisDenom,
	}
	s.network.App.TransferKeeper.SetDenomTrace(s.network.GetContext(), denomTrace)
	osmoMetadata := banktypes.Metadata{
		Name:        "Evmos Osmo Token",
		Symbol:      "OSMO",
		Description: "The IBC representation of OSMO on Evmos chain",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denomTrace.IBCDenom(),
				Exponent: 0,
				Aliases:  []string{"uosmo"},
			},
			{
				Denom:    "uosmo",
				Exponent: 18,
			},
		},
		Display: "osmo",
		Base:    denomTrace.IBCDenom(),
	}

	s.T().Log("Before minting osmo...")
	amount := sdk.NewInt(TokenToMint)
	coin := sdk.NewCoin(osmoMetadata.Base, amount)
	err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, sdk.NewCoins(coin))
	s.Require().NoError(err)

	s.T().Log("Before registering osmo...")
	// Register Evmos Token Pair.
	_, err = s.network.App.Erc20Keeper.RegisterCoin(s.network.GetContext(), osmoMetadata)
	s.Require().NoError(err)

	s.T().Log("Before sending osmo...")
	sender := s.keyring.GetAccAddr(0)
	err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.network.GetContext(),
		inflationtypes.ModuleName,
		sender,
		sdk.NewCoins(coin),
	)
	s.Require().NoError(err)

	s.T().Log("Before getting denom osmo...")
	// Retrieve Osmo token information useful for the testing
	osmoDenomID := s.network.App.Erc20Keeper.GetDenomMap(
		s.network.GetContext(),
		denomTrace.IBCDenom(),
	)
	_, ok := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), osmoDenomID)
	s.Require().True(ok, "expected osmo token pair to be found")
}
