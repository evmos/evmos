package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

func (suite *KeeperTestSuite) TestReceive() {
	var packet channeltypes.Packet

	pk := secp256k1.GenPrivKey()
	secpAddr := sdk.AccAddress(pk.PubKey().Address())
	secpAddrEvmos := secpAddr.String()
	secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)
	// senderStr := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	// receiverStr := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"
	// sender, err := sdk.AccAddressFromBech32(senderStr)
	// suite.Require().NoError(err)
	// receiver, err := sdk.AccAddressFromBech32(receiverStr)
	// suite.Require().NoError(err)

	timeoutHeight := clienttypes.NewHeight(0, 100)
	disabledTimeoutTimestamp := uint64(0)
	mockpacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	expAck := ibcmock.MockAcknowledgement

	testCases := []struct {
		name       string
		malleate   func()
		ackSuccess bool
	}{
		{
			"fail - params disabled",
			func() {
				params := suite.app.WithdrawKeeper.GetParams(suite.ctx)
				params.EnableWithdraw = false
				suite.app.WithdrawKeeper.SetParams(suite.ctx, params)

				packet = mockpacket
			},
			true,
		},
		{
			"fail - wrong packet",
			func() {
				packet = mockpacket
			},
			false,
		},
		{
			"TODO case 4: same sender with EVM channel, no claims record",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, types.DefaultEVMChannels[0], timeoutHeight, 0)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			params := suite.app.WithdrawKeeper.GetParams(suite.ctx)
			params.EnableWithdraw = true
			suite.app.WithdrawKeeper.SetParams(suite.ctx, params)

			tc.malleate()

			ack := suite.app.WithdrawKeeper.OnRecvPacket(suite.ctx, packet, expAck)
			if tc.ackSuccess {
				suite.Require().Equal(expAck, ack)
			} else {
				fmt.Println(ack.Success())
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}
		})
	}
}
