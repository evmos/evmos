package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"

	"github.com/tharsis/evmos/v5/x/claims/types"
)

var timeoutHeight = clienttypes.NewHeight(1000, 1000)

func (suite *KeeperTestSuite) TestAckknowledgementPacket() {
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
			"invalid ACK",
			func() {
				err := suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, []byte{})
				suite.Require().Error(err)
			},
		},
		{
			"non ics20 packet",
			func() {
				err := suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().Error(err)
			},
		},
		{
			"no-op: error Ack",
			func() {
				err := sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
				ack := transfertypes.NewErrorAcknowledgement(err)
				err = suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().NoError(err)
			},
		},
		{
			"error - no escrowed funds",
			func() {
				addr, err := sdk.AccAddressFromBech32("evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v")
				suite.Require().NoError(err)

				mockpacket.Data = transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v",
						Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
					},
				)

				cr := types.NewClaimsRecord(sdk.NewInt(100))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)
				err = suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().Error(err)
			},
		},
		{
			"noop - claims record not found ",
			func() {
				suite.SetupTestWithEscrow()

				addr, err := sdk.AccAddressFromBech32("evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v")
				suite.Require().NoError(err)

				mockpacket.Data = transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v",
						Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
					},
				)

				err = suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().NoError(err)

				_, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().False(found)
			},
		},
		{
			"pass - claim IBC action ",
			func() {
				suite.SetupTestWithEscrow()

				addr, err := sdk.AccAddressFromBech32("evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v")
				suite.Require().NoError(err)

				mockpacket.Data = transfertypes.ModuleCdc.MustMarshalJSON(
					&transfertypes.FungibleTokenPacketData{
						Sender:   "evmos1x2w87cvt5mqjncav4lxy8yfreynn273xn5335v",
						Receiver: "cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc",
					},
				)

				cr := types.NewClaimsRecord(sdk.NewInt(100))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, cr)

				err = suite.app.ClaimsKeeper.OnAcknowledgementPacket(suite.ctx, mockpacket, ack.Acknowledgement())
				suite.Require().NoError(err)

				expCR := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{false, false, false, true},
				}
				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
				suite.Require().True(found)
				suite.Require().Equal(expCR, cr)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.test()
		})
	}
}

func (suite *KeeperTestSuite) TestOnRecvPacket() {
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
			"no-op - params disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, mockpacket, ack)
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
			"fail - blocked recipient (deny list)",
			func() {
				blockedAddr := authtypes.NewModuleAddress(transfertypes.ModuleName)
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, blockedAddr.String())
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())
			},
		},
		{
			"no-op - sender record found with ibc action already claimed",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				expCR := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{false, false, false, true},
				}
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, expCR)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// claims record not changed
				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, sender)
				suite.Require().True(found)
				suite.Require().Equal(expCR, cr)
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
			"no-op - channel not authorized",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-100", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().Equal(ack, resAck)
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
			"pass - sender and receiver address is the same (no claim record) - attempt recovery",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())
			},
		},
		{
			"case 1: no-op - sender ≠ recipient, but wrong trigger amount",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(100)))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 1: fail - not enough funds on escrow account",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", types.IBCTriggerAmt, senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(10000000000)))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, types.NewClaimsRecord(sdk.NewInt(10000000000)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success())

				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 1: pass/merge - sender ≠ recipient",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", types.IBCTriggerAmt, senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(100)))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is merged to the recipient
				expCR := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(200),
					ActionsCompleted:       []bool{false, false, false, true},
				}

				// check that the record is migrated and action is completed
				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(s.ctx, receiver)
				suite.Require().True(found)
				suite.Require().Equal(expCR, cr)
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
			},
		},
		{
			"case 2: no-op - same sender ≠ recipient, sender claims record found, but wrong types.IBCTriggerAmt",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is not migrated
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 2: no-op - same sender ≠ recipient, sender claims record found, not enough escowed funds",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", types.IBCTriggerAmt, senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(9000000000000000000)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().False(resAck.Success(), ack.String())

				// check that the record is not migrated
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 2: pass/migrate - same sender ≠ recipient, sender claims record found",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", types.IBCTriggerAmt, senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, sender, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				// check that the record is migrated
				suite.Require().False(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, sender))
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))

				expCR := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{false, false, false, true},
				}

				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, receiver)
				suite.Require().True(found)
				suite.Require().Equal(expCR, cr)
			},
		},
		{
			"case 3: fail - not enough funds on escrow account",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, types.NewClaimsRecord(sdk.NewInt(1000000000000000)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)

				var ack channeltypes.Acknowledgement
				transfertypes.ModuleCdc.MustUnmarshalJSON(resAck.Acknowledgement(), &ack)
				suite.Require().False(resAck.Success(), ack.String())

				// check that the record is not deleted
				suite.Require().True(suite.app.ClaimsKeeper.HasClaimsRecord(suite.ctx, receiver))
			},
		},
		{
			"case 3: pass/claim - same sender ≠ recipient, recipient claims record found",
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
			"case 3: claim - same Address with authorized EVM channel, with claims record",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.AuthorizedChannels = []string{
					"channel-2", // Injective
				}
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultEVMChannels[0], timeoutHeight, 0)

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, secpAddr, types.NewClaimsRecord(sdk.NewInt(100)))

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				expCR := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{false, false, false, true},
				}

				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(s.ctx, secpAddr)
				// check that the record is not deleted and action is completed
				suite.Require().True(found)
				suite.Require().Equal(expCR, cr)
			},
		},
		{
			"case 3: claim - same Address with unauthorized EVM channel, with claims record",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.AuthorizedChannels = []string{
					"channel-0", // Osmosis
					"channel-3", // Cosmos Hub
				}
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultEVMChannels[0], timeoutHeight, 0)

				cr := types.NewClaimsRecord(sdk.NewInt(100))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, secpAddr, cr)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				crAfter, found := suite.app.ClaimsKeeper.GetClaimsRecord(s.ctx, secpAddr)
				// check that the record is not deleted and action is completed
				suite.Require().True(found)
				suite.Require().Equal(crAfter, cr)
			},
		},
		{
			"case 3: claim - sender ≠ recipient, recipient claims record found, where ibc is last action",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet := channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)

				cr := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{true, true, true, false},
				}

				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, receiver, cr)

				resAck := suite.app.ClaimsKeeper.OnRecvPacket(suite.ctx, packet, ack)
				suite.Require().True(resAck.Success())

				expCR := types.ClaimsRecord{
					InitialClaimableAmount: sdk.NewInt(100),
					ActionsCompleted:       []bool{true, true, true, true},
				}

				cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(s.ctx, receiver)
				// check that the record is not deleted and action is completed
				suite.Require().True(found)
				suite.Require().Equal(expCR, cr)
			},
		},
		{
			"case 4: no-op - sender different than recipient, no claims records",
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
			"case 4: no-op - same sender with EVM channel, no claims record",
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
			suite.SetupTestWithEscrow() // reset

			tc.test()
		})
	}
}
