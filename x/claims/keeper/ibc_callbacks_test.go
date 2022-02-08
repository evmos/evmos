package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	path *ibctesting.Path
}

func (suite *IBCTestingSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = app.SetupTestingApp

	ibctesting.ChainIDPrefix = "evmos_9000-"
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)         // initializes 2 test chains
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1)) // convenience and readability
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2)) // convenience and readability

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10000)))
	err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.chainB.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainB.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)

	err = suite.chainA.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainA.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.chainA.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainA.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)

	params := types.DefaultParams()
	params.AirdropStartTime = suite.chainA.GetContext().BlockTime()
	params.EnableClaims = true
	suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainA.GetContext(), params)
	suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)

	suite.path = NewTransferPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(suite.path)                      // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", suite.path.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.path.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.path.EndpointA.ChannelID)
}

func TestIBCTestingSuite(t *testing.T) {
	suite.Run(t, new(IBCTestingSuite))
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

func (suite *IBCTestingSuite) TestOnReceiveClaim() {
	senderstr := "cosmos1adjs2y3gchg28k7zup8wwmyjv3rrnylc0hufk3"
	receiverstr := "cosmos1s06n8al83537v5nrlxf6v94v4jaug50cd42xlx"
	senderaddr, _ := sdk.AccAddressFromBech32(senderstr)
	receiveraddr, _ := sdk.AccAddressFromBech32(receiverstr)

	testCases := []struct {
		name            string
		malleate        func(int64)
		claimableAmount int64
		expectedBalance int64
		expPass         bool
	}{
		{
			"correct execution - Claimable Transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderaddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderaddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{true, true, true, true}})
			},
			4,
			0,
			true,
		},
		{
			"correct execution - Recipient Claimable transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiveraddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Recipient Claimed transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiveraddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{true, true, true, true}})
			},
			4,
			0,
			true,
		},
		{
			"Disabled by params",
			func(_ int64) {
				params := types.DefaultParams()
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainA.GetContext(), params)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)
			},
			0,
			0,
			false,
		},
		{
			"No claim record",
			func(claimableAmount int64) {
			},
			0,
			0,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.path

			tc.malleate(tc.claimableAmount)

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderstr, receiverstr)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			suite.path.EndpointA.SendPacket(packet)

			// receive on endpointB
			err := path.EndpointB.RecvPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiveraddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				claim, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiveraddr)
				suite.Require().True(found)
				suite.Require().Equal(claim.InitialClaimableAmount, sdk.NewInt(4))
			} else {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiveraddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiveraddr)
				suite.Require().True(!found)

			}
		})
	}
}

func (suite *IBCTestingSuite) TestOnAckClaim() {
	senderstr := "cosmos1adjs2y3gchg28k7zup8wwmyjv3rrnylc0hufk3"
	receiverstr := "cosmos1s06n8al83537v5nrlxf6v94v4jaug50cd42xlx"
	senderaddr, _ := sdk.AccAddressFromBech32(senderstr)

	testCases := []struct {
		name            string
		malleate        func(int64)
		claimableAmount int64
		expectedBalance int64
		expPass         bool
	}{
		{
			"correct execution - Claimable Transfer",
			func(claimableAmount int64) {
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderaddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderaddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{true, true, true, true}})
			},
			4,
			0,
			true,
		},
		{
			"Disabled by params",
			func(_ int64) {
				params := types.DefaultParams()
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainA.GetContext(), params)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)
			},
			0,
			0,
			false,
		},
		{
			"No claim record",
			func(claimableAmount int64) {
			},
			0,
			0,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.path

			tc.malleate(tc.claimableAmount)

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderstr, receiverstr)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			suite.path.EndpointA.SendPacket(packet)

			// receive on endpointB
			err := path.EndpointB.RecvPacket(packet)
			suite.Require().NoError(err)

			// TODO: should use testing method path.EndpointA.AcknowledgePacket(packet, ack)
			err = suite.chainA.App.(*app.Evmos).ClaimsKeeper.OnAcknowledgementPacket(suite.chainA.GetContext(), packet, ibctesting.MockAcknowledgement)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderaddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				claim, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderaddr)
				suite.Require().True(found)
				suite.Require().Equal(claim.InitialClaimableAmount, sdk.NewInt(4))
			} else {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderaddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				_, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderaddr)
				suite.Require().True(!found)
			}
		})
	}
}
