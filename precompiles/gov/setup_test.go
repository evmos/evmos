package gov_test

import (
	"testing"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/evmos/evmos/v20/precompiles/gov"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"

	"github.com/stretchr/testify/suite"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *gov.Precompile
}

func TestPrecompileUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)

	// seed the db with one proposal
	customGen := network.CustomGenesisState{}
	now := time.Now().UTC()
	inOneHour := now.Add(time.Hour)
	prop := &govv1.Proposal{
		Id:              1,
		Status:          govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:      &now,
		DepositEndTime:  &inOneHour,
		VotingStartTime: &now,
		VotingEndTime:   &inOneHour,
		Metadata:        "ipfs://CID",
		Title:           "test prop",
		Summary:         "test prop",
		Proposer:        keyring.GetAccAddr(0).String(),
	}
	govGen := govv1.DefaultGenesisState()
	govGen.Proposals = append(govGen.Proposals, prop)
	customGen[govtypes.ModuleName] = govGen

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

	var err error
	if s.precompile, err = gov.NewPrecompile(
		s.network.App.GovKeeper,
		s.network.App.AuthzKeeper,
	); err != nil {
		panic(err)
	}
}
