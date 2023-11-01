// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis_test

import (
	"testing"
	"time"

	"github.com/evmos/evmos/v15/precompiles/outposts/osmosis"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmosapp "github.com/evmos/evmos/v15/app"
	evmosibc "github.com/evmos/evmos/v15/ibc/testing"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

type PrecompileTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	app           *evmosapp.Evmos
	address       common.Address
	differentAddr common.Address
	validators    []stakingtypes.Validator
	valSet        *tmtypes.ValidatorSet
	ethSigner     ethtypes.Signer
	privKey       cryptotypes.PrivKey
	signer        keyring.Signer
	bondDenom     string

	precompile *osmosis.Precompile
	stateDB    *statedb.StateDB

	coordinator    *ibctesting.Coordinator
	chainA         *ibctesting.TestChain
	chainB         *ibctesting.TestChain
	transferPath   *evmosibc.Path
	queryClientEVM evmtypes.QueryClient

	defaultExpirationDuration time.Time

	suiteIBCTesting bool
}

type PrecompileTestSuiteV2 struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	stateDB    *statedb.StateDB
	precompile *osmosis.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	s.DoSetupTest()
}

func TestPrecompileTestSuiteV2(t *testing.T) {
	s2 := new(PrecompileTestSuiteV2)
	suite.Run(t, s2)
}

func (s *PrecompileTestSuiteV2) SetupTest() {
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
		portId,
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
}
