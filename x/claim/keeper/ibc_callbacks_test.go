package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/claim/types"
)

type CallbackTestSuite struct {
	suite.Suite
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
	timeoutHeight = clienttypes.NewHeight(2, 10000)
	maxSequence   = uint64(10)
)

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
}

func (suite *CallbackTestSuite) TestCallbacks() {

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10000)))
	err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.chainB.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainB.GetContext(), minttypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)

	params := types.DefaultParams()
	params.AirdropStartTime = suite.chainA.GetContext().BlockTime()
	params.EnableClaim = true
	suite.chainA.App.(*app.Evmos).ClaimKeeper.SetParams(suite.chainA.GetContext(), params)
	suite.chainB.App.(*app.Evmos).ClaimKeeper.SetParams(suite.chainB.GetContext(), params)

	path := NewTransferPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(path)                       // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", path.EndpointA.ClientID)
	suite.Require().Equal("connection-0", path.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", path.EndpointA.ChannelID)

	sendaddr, err := sdk.AccAddressFromBech32("cosmos1adjs2y3gchg28k7zup8wwmyjv3rrnylc0hufk3")
	suite.Require().NoError(err)
	suite.chainB.App.(*app.Evmos).ClaimKeeper.SetClaimRecord(suite.chainB.GetContext(), sendaddr, types.ClaimRecord{InitialClaimableAmount: sdk.NewInt(4), ActionsCompleted: []bool{true}})

	transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "cosmos1adjs2y3gchg28k7zup8wwmyjv3rrnylc0hufk3", "cosmos1s06n8al83537v5nrlxf6v94v4jaug50cd42xlx")
	bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
	packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// send on endpointA
	path.EndpointA.SendPacket(packet)

	// receive on endpointB
	path.EndpointB.RecvPacket(packet)

	recaddr, err := sdk.AccAddressFromBech32("cosmos1s06n8al83537v5nrlxf6v94v4jaug50cd42xlx")
	//claim, found := suite.chainB.App.(*app.Evmos).ClaimKeeper.GetClaimRecord(suite.chainB.GetContext(), recaddr)
	//suite.Require().True(found)
	//suite.Require().Equal(claim.InitialClaimableAmount, sdk.NewInt(4))

	coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), recaddr, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(1)))

	// acknowledge the receipt of the packet
	//path.EndpointA.AcknowledgePacket(packet, ack)
}
