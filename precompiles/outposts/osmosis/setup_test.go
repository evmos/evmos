// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
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

	unitNetwork *network.UnitTestNetwork
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	stateDB *statedb.StateDB

	precompile *osmosis.Precompile

	coordinator *coordinator.IntegrationCoordinator
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	precompile, err := osmosis.NewPrecompile(
		portID,
		channelID,
		osmosis.XCSContract,
		unitNetwork.App.BankKeeper,
		unitNetwork.App.TransferKeeper,
		unitNetwork.App.StakingKeeper,
		unitNetwork.App.Erc20Keeper,
		unitNetwork.App.AuthzKeeper,
	)
	s.Require().NoError(err, "expected no error during precompile creation")

	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)

	s.unitNetwork = unitNetwork
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile

	s.registerEvmosERC20Coins()
	s.registerOsmoERC20Coins()

	coordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{unitNetwork},
	)
	// Account to sign IBC txs
	acc, err := s.grpcHandler.GetAccount(s.keyring.GetAccAddr(0).String())
	coordinator.SetDefaultSignerForChain(s.network.GetChainID(), s.keyring.GetPrivKey(0), acc)

	dummyChains := coordinator.GetDummyChainsIds()
	_ = coordinator.Setup(s.network.GetChainID(), dummyChains[0])
	s.coordinator = coordinator
}
