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

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/v2/app"
	ibctesting "github.com/tharsis/evmos/v2/ibc/testing"
	"github.com/tharsis/evmos/v2/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	chainA      *ibcgotesting.TestChain // Evmos chain A
	chainB      *ibcgotesting.TestChain // Evmos chain B
	chainCosmos *ibcgotesting.TestChain // Cosmos chain

	pathEVM    *ibcgotesting.Path // chainA (Evmos) <-->  chainB (Evmos)
	pathCosmos *ibcgotesting.Path // chainA (Evmos) <--> chainCosmos
}

func (suite *IBCTestingSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2, 1) // initializes 2 Evmos test chains and 1 Cosmos Chain
	suite.chainA = suite.coordinator.GetChain(ibcgotesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibcgotesting.GetChainID(2))
	suite.chainCosmos = suite.coordinator.GetChain(ibcgotesting.GetChainID(3))

	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)
	suite.coordinator.CommitNBlocks(suite.chainCosmos, 2)

	claimsRecord := types.NewClaimsRecord(sdk.NewInt(10000))
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10000)))

	err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.chainB.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainB.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)
	suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), addr, claimsRecord)

	err = suite.chainA.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainA.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.chainA.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainA.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)
	suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), addr, claimsRecord)

	params := types.DefaultParams()
	params.AirdropStartTime = suite.chainA.GetContext().BlockTime()
	params.EnableClaims = true
	suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainA.GetContext(), params)
	suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)

	suite.pathEVM = ibctesting.NewTransferPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(suite.pathEVM)                                 // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", suite.pathEVM.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.pathEVM.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.pathEVM.EndpointA.ChannelID)

	suite.pathCosmos = ibctesting.NewTransferPath(suite.chainA, suite.chainCosmos) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(suite.pathCosmos)                                      // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-1", suite.pathCosmos.EndpointA.ClientID)
	suite.Require().Equal("connection-1", suite.pathCosmos.EndpointA.ConnectionID)
	suite.Require().Equal("channel-1", suite.pathCosmos.EndpointA.ChannelID)
}

func TestIBCTestingSuite(t *testing.T) {
	suite.Run(t, new(IBCTestingSuite))
}

var timeoutHeight = clienttypes.NewHeight(1000, 1000)

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
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.NewClaimsRecord(amt))

				// update the escrowed account balance to maintain the invariant
				err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.chainB.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainB.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Merge Transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt.Add(amt.QuoRaw(2))))

				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, false, true, false}})
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, true, false, false}})

				// update the escrowed account balance to maintain the invariant
				err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.chainB.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainB.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			4,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, true}})
			},
			4,
			0,
			true,
		},
		{
			"correct execution - Recipient Claimable transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, false, false, false}})

				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))
				// update the escrowed account balance to maintain the invariant
				err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.chainB.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainB.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Recipient Claimed transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, true}})
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
			path := suite.pathEVM

			tc.malleate(tc.claimableAmount)

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			path.EndpointA.SendPacket(packet)

			// receive on endpointB
			err := path.EndpointB.RecvPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiverAddr)
				suite.Require().True(found)
			} else {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
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
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.NewClaimsRecord(amt))
				// update the escrowed account balance to maintain the invariant
				err := suite.chainA.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainA.GetContext(), inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.chainA.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainA.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimable Transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.NewClaimsRecord(amt))
				// update the escrowed account balance to maintain the invariant
				err := suite.chainA.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainA.GetContext(), inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.chainA.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainA.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, true}})

				// update the escrowed account balance to maintain the invariant
				err := suite.chainA.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainA.GetContext(), inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.chainA.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToModule(suite.chainA.GetContext(), inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
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
			path := suite.pathEVM

			tc.malleate(tc.claimableAmount)

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			err := path.EndpointA.SendPacket(packet)
			suite.Require().NoError(err)

			err = path.RelayPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
				claim, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderAddr)
				suite.Require().True(found)
				suite.Require().Equal(claim.InitialClaimableAmount, sdk.NewInt(4))
			} else {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
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
	senderStr := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	receiverStr := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"
	sender, err := sdk.AccAddressFromBech32(senderStr)
	suite.Require().NoError(err)
	receiver, err := sdk.AccAddressFromBech32(receiverStr)
	suite.Require().NoError(err)

	disabledTimeoutTimestamp := uint64(0)
	timeoutHeight = clienttypes.NewHeight(0, 100)
	mockpacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	ack := ibcmock.MockAcknowledgement

	testCases := []struct {
		name string
		test func()
	}{
		{
			"fail - params disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, mockpacket, ack)
				suite.Require().Equal(ack, resAck)
			},
		},
		{
			"fail - params, channel not authorized",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-100", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().Equal(ack, resAck)
			},
		},
		{
			"fail - non ics20 packet",
			func() {
				err := sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
				expectedAck := channeltypes.NewErrorAcknowledgement(err.Error())
				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, mockpacket, ack)
				suite.Require().Equal(expectedAck, resAck)
			},
		},
		{
			"fail - invalid sender",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "evmos", receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"fail - invalid sender 2",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"fail - invalid recipient",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", receiverStr, "badbadhf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625")
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
			"fail - sender and receiver address are the same (with claim record)",
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
			"case 1: sender ≠ recipient",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(100)))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is merged to the recipient
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 2: same sender ≠ recipient, sender claims record found",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is migrated
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 3: same sender ≠ recipient, recipient claims record found",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is not deleted
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 3: same sender with EVM channel, with claims record",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultEVMChannels[0], timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, secpAddr, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is not deleted
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, secpAddr))
			},
		},
		{
			"case 4: sender different than recipient, no claims records",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultAuthorizedChannels[0], timeoutHeight, 0)

				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())
			},
		},
		{
			"case 4: same sender with EVM channel, no claims record",
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
