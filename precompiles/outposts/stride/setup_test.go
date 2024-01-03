// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride_test

import (
	"testing"

	"github.com/evmos/evmos/v16/precompiles/erc20"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/precompiles/outposts/stride"
	"github.com/evmos/evmos/v16/testutil/integration/common/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
)

var _ *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *stride.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	network := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	precompile, err := stride.NewPrecompile(
		common.HexToAddress(erc20.WEVMOSContractTestnet),
		network.App.TransferKeeper,
		network.App.Erc20Keeper,
		network.App.AuthzKeeper,
		network.App.StakingKeeper,
	)
	s.Require().NoError(err, "expected no error during precompile creation")
	s.precompile = precompile

	grpcHandler := grpc.NewIntegrationHandler(network)

	s.network = network
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.precompile = precompile

	// Register stEvmos Coin as an ERC20 token
	s.registerStrideCoinERC20()
}
