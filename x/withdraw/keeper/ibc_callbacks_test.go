package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

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
				suite.Require().True(resAck.Success(), string(resAck.Acknowledgement()))
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
