package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"

	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/ibctesting"
	"github.com/tharsis/evmos/v2/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibcgotesting.TestChain
	chainB *ibcgotesting.TestChain

	path *ibcgotesting.Path
}

func (suite *IBCTestingSuite) SetupTest() {
	ibcgotesting.DefaultTestingAppInit = app.SetupTestingApp

	ibcgotesting.ChainIDPrefix = "evmos_9000-"
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)           // initializes 2 test chains
	suite.chainA = suite.coordinator.GetChain(ibcgotesting.GetChainID(1)) // convenience and readability
	suite.chainB = suite.coordinator.GetChain(ibcgotesting.GetChainID(2)) // convenience and readability
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)

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

func (suite *IBCTestingSuite) TestOnReceiveClaim() {
	sender := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"
	receiver := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

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
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Merge Transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{false, false, true, false}})
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{false, true, false, false}})
			},
			4,
			4,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{true, true, true, true}})
			},
			4,
			0,
			true,
		},
		{
			"correct execution - Recipient Claimable transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Recipient Claimed transfer",
			func(claimableAmount int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{true, true, true, true}})
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

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			suite.path.EndpointA.SendPacket(packet)

			// receive on endpointB
			err := path.EndpointB.RecvPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiverAddr)
				suite.Require().True(found)
			} else {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiverAddr)
				suite.Require().False(found)
			}
		})
	}
}

func (suite *IBCTestingSuite) TestOnAckClaim() {
	sender := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	receiver := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)

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
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimable Transfer",
			func(claimableAmount int64) {
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{}})
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(claimableAmount), ActionsCompleted: []bool{true, true, true, true}})
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

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// // send on endpointA
			err := suite.path.EndpointA.SendPacket(packet)
			suite.Require().NoError(err)

			err = suite.path.RelayPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderAddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				claim, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderAddr)
				suite.Require().True(found)
				suite.Require().Equal(claim.InitialClaimableAmount, sdk.NewInt(4))
			} else {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderAddr, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)))
				_, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderAddr)
				suite.Require().False(found)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestReceive() {
	pk := secp256k1.GenPrivKey()
	secpAddr := sdk.AccAddress(pk.PubKey().Address())
	secpAddrEvmos := secpAddr.String()
	secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)
	sender := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	receiver := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"

	disabledTimeoutTimestamp := uint64(0)
	timeoutHeight = clienttypes.NewHeight(0, 100)
	mockpacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	ack := ibcmock.MockAcknowledgement

	testCases := []struct {
		name string
		test func()
	}{
		{
			"params disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, mockpacket, ack)
				suite.Require().Equal(ack, resAck)
			},
		},
		{
			"params, channel not authorized",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-100", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().Equal(ack, resAck)
			},
		},
		{
			"non ics20 packet",
			func() {
				err := sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
				expectedAck := channeltypes.NewErrorAcknowledgement(err.Error())
				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, mockpacket, ack)
				suite.Require().Equal(expectedAck, resAck)
			},
		},
		{
			"invalid sender",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "evmos", receiver)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"invalid sender 2",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", receiver)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"invalid recipient",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", receiver, "badbadhf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"fail - sender and receiver address is the same (no claim record)",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"fail - sender and receiver address is the same (with claim record)",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, secpAddr, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"correct",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultAuthorizedChannels[0], timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())
			},
		},
		{
			"correct, same sender with EVM channel",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultEVMChannels[0], timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupClaimTest() // reset

			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestAck() {
	disabledTimeoutTimestamp := uint64(0)
	timeoutHeight = clienttypes.NewHeight(0, 100)
	mockpacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	ack := ibcmock.MockAcknowledgement

	testCases := []struct {
		name string
		test func()
	}{
		{
			"params disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				err := suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().NoError(err)
			},
		},
		{
			"non ics20 packet",
			func() {
				err := suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, []byte{3})
				suite.Require().Error(err)
			},
		},
		{
			"error Ack",
			func() {
				err := sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
				ack := transfertypes.NewErrorAcknowledgement(err)
				err = suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupClaimTest() // reset

			tc.test()
		})
	}
}
