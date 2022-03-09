package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"

	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/ibctesting"
	claimstypes "github.com/tharsis/evmos/v2/x/claims/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibcgotesting.TestChain
	chainB *ibcgotesting.TestChain

	path *ibcgotesting.Path
}

var s *IBCTestingSuite

func TestIBCTestingSuite(t *testing.T) {
	s = new(IBCTestingSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "IBC Testing Suite")
}

func (suite *IBCTestingSuite) SetupTest() {
	ibcgotesting.DefaultTestingAppInit = app.SetupTestingApp

	ibcgotesting.ChainIDPrefix = "evmos_9000-"
	suite.coordinator = ibctesting.NewMixedCoordinator(suite.T(), 1, 1)   // initializes 2 test chains
	suite.chainA = suite.coordinator.GetChain(ibcgotesting.GetChainID(2)) // convenience and readability
	suite.chainB = suite.coordinator.GetChain(ibcgotesting.GetChainID(1)) // convenience and readability
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)

	params := claimstypes.DefaultParams()
	params.AirdropStartTime = suite.chainA.GetContext().BlockTime()
	params.EnableClaims = true
	suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)

	suite.path = NewTransferPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(suite.path)                      // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", suite.path.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.path.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.path.EndpointA.ChannelID)
}

var timeoutHeight = clienttypes.NewHeight(1000, 1000)

func NewTransferPath(chainA, chainB *ibcgotesting.TestChain) *ibcgotesting.Path {
	path := ibcgotesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibcgotesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibcgotesting.TransferPort

	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = "ics20-1"
	path.EndpointB.ChannelConfig.Version = "ics20-1"

	return path
}

// func (suite *IBCTestingSuite) TestOnReceiveWithdraw() {

// 	testCases := []struct {
// 		name    string
// 		expPass bool
// 	}{
// 		{
// 			"correct execution",
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
// 			suite.SetupTest()
// 			path := suite.path

// 			coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.senderAcc, "aevmos")
// 			suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(10000)))

// 			transfer := transfertypes.NewFungibleTokenPacketData("testcoin", "10", suite.sender, suite.sender)
// 			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
// 			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

// 			// send on endpointA
// 			suite.path.EndpointA.SendPacket(packet)

// 			err := suite.path.RelayPacket(packet)
// 			suite.Require().NoError(err)

// 			// Recreate packets that were sent in the ibc_callback
// 			transfer2 := transfertypes.FungibleTokenPacketData{
// 				Amount:   "10000",
// 				Denom:    "aevmos",
// 				Receiver: suite.sender,
// 				Sender:   suite.sender,
// 			}
// 			packet2 := channeltypes.NewPacket(
// 				transfer2.GetBytes(),
// 				1,
// 				"transfer",
// 				"channel-0",
// 				"transfer",
// 				"channel-0",
// 				clienttypes.ZeroHeight(), // timeout height disabled
// 				1677926229000000000,      // timeout timestamp disabled
// 			)

// 			transfer3 := transfertypes.FungibleTokenPacketData{
// 				Amount:   "10",
// 				Denom:    "transfer/channel-0/testcoin",
// 				Receiver: suite.sender,
// 				Sender:   suite.sender,
// 			}
// 			packet3 := channeltypes.NewPacket(
// 				transfer3.GetBytes(),
// 				2,
// 				"transfer",
// 				"channel-0",
// 				"transfer",
// 				"channel-0",
// 				clienttypes.ZeroHeight(), // timeout height disabled
// 				1677926229000000000,      // timeout timestamp disabled
// 			)

// 			// Relay both packets that were sent in the ibc_callback
// 			err = suite.path.RelayPacket(packet2)
// 			suite.Require().NoError(err)
// 			err = suite.path.RelayPacket(packet3)
// 			suite.Require().NoError(err)

// 			if tc.expPass {
// 				coin = suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), suite.senderAcc, "aevmos")
// 				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))
// 				coins := suite.chainA.App.(*app.Evmos).BankKeeper.GetAllBalances(suite.chainA.GetContext(), suite.senderAcc)
// 				suite.Require().Equal(coins[0].Amount, sdk.NewInt(10000))
// 				suite.Require().Equal(coins[1].Amount, sdk.NewInt(10))
// 			}
// 		})
// 	}
// }
