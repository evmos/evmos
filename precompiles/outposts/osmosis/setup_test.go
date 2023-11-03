// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
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
	"github.com/evmos/evmos/v15/utils"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
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
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
	txFactory := factory.New(integrationNetwork, grpcHandler)

	headerHash := integrationNetwork.GetContext().HeaderHash()
	stateDB := statedb.New(
		integrationNetwork.GetContext(),
		integrationNetwork.App.EvmKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash.Bytes())),
	)

	precompile, err := osmosis.NewPrecompile(
		portID,
		channelID,
		osmosis.XCSContract,
		integrationNetwork.App.BankKeeper,
		integrationNetwork.App.TransferKeeper,
		integrationNetwork.App.StakingKeeper,
		integrationNetwork.App.Erc20Keeper,
	)
	s.Require().NoError(err)

	s.stateDB = stateDB
	s.network = integrationNetwork
	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile

	network := network.New()
	coordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{network},
	)

	chainA := coordinator.GetChain(ibctesting.GetChainID(1)).(*ibctesting.TestChain)
	chainB := coordinator.GetChain(ibctesting.GetChainID(3)).(*ibctesting.TestChain)

	path := ibctesting.NewPath(chainA, chainB)

	s.registerERC20Coins()
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

// registerERC20Coins registers Evmos and IBC OSMO coin as an ERC20 token
func (s *PrecompileTestSuite) registerERC20Coins() {
	// Register EVMOS ERC20 equivalent
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

	coin := sdk.NewCoin(evmosMetadata.Base, sdk.NewInt(TokenToMint))
	err := s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, sdk.NewCoins(coin))
	s.Require().NoError(err)

	// Register Evmos Token Pair
	_, err = s.network.App.Erc20Keeper.RegisterCoin(s.network.GetContext(), evmosMetadata)
	s.Require().NoError(err)

	// Register IBC OSMO Token Pair
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

	osmo := sdk.NewCoin(osmoMetadata.Base, sdk.NewInt(TokenToMint))
	err = s.network.App.BankKeeper.MintCoins(s.network.GetContext(), inflationtypes.ModuleName, sdk.NewCoins(osmo))
	s.Require().NoError(err)
	err = s.network.App.BankKeeper.SendCoinsFromModuleToAccount(
		s.network.GetContext(),
		inflationtypes.ModuleName,
		s.keyring.GetAddr(0).Bytes(),
		sdk.NewCoins(osmo),
	)
	s.Require().NoError(err)

	_, err = s.network.App.Erc20Keeper.RegisterCoin(s.network.GetContext(), osmoMetadata)
	s.Require().NoError(err)

	convertCoin := erc20types.NewMsgConvertCoin(
		osmo,
		s.keyring.GetAddr(0),
		s.keyring.GetAddr(0).Bytes(),
	)

	_, err = s.network.App.Erc20Keeper.ConvertCoin(s.network.GetContext(), convertCoin)
	s.Require().NoError(err)

	// Retrieve Evmos token information useful for the testing
	evmosDenomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), bondDenom)
	_, ok := s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), evmosDenomID)
	s.Require().True(ok, "expected evmos token pair to be found")

	// Retrieve Osmo token information useful for the testing
	osmoIBCDenom := utils.ComputeIBCDenom(portID, channelID, osmosis.OsmosisDenom)
	osmoDenomID := s.network.App.Erc20Keeper.GetDenomMap(s.network.GetContext(), osmoIBCDenom)
	_, ok = s.network.App.Erc20Keeper.GetTokenPair(s.network.GetContext(), osmoDenomID)
	s.Require().True(ok, "expected osmo token pair to be found")
}
