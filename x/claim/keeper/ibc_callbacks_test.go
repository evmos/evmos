package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/claim/types"
)

type CallbackTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app         *app.Evmos
	queryClient types.QueryClient

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *CallbackTestSuite) SetupTest() {

	ibctesting.DefaultTestingAppInit = app.SetupTestingApp

	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2) // initializes 2 test chains
	suite.chainA = suite.coordinator.GetChain("evmos_9000-1")   // convenience and readability
	suite.chainB = suite.coordinator.GetChain("evmos_9000-2")   // convenience and readability

}

func TestCallbackTestSuite(t *testing.T) {
	suite.Run(t, new(CallbackTestSuite))
}

var (
	timeoutHeight = clienttypes.NewHeight(0, 10000)
	maxSequence   = uint64(10)
)

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	return path
}

func (suite *CallbackTestSuite) TestCallbacks() {
	path := NewTransferPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(path)                       // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", path.EndpointA.ClientID)
	suite.Require().Equal("connection-0", path.EndpointA.ClientID)
	suite.Require().Equal("channel-0", path.EndpointA.ClientID)

	packet := channeltypes.NewPacket(ibctesting.MockPacketData, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// send on endpointA
	path.EndpointA.SendPacket(packet)

	// receive on endpointB
	path.EndpointB.RecvPacket(packet)

	// acknowledge the receipt of the packet
	//path.EndpointA.AcknowledgePacket(packet, ack)
}
