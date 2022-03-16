package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/evmos/v2/testutil"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"
	ibcmock "github.com/cosmos/ibc-go/v3/testing/mock"
)

func (suite *KeeperTestSuite) TestReceive() {
	pk := secp256k1.GenPrivKey()
	secpAddr := sdk.AccAddress(pk.PubKey().Address())
	secpAddrEvmos := secpAddr.String()
	secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)
	senderStr := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	receiverStr := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"
	// sender, err := sdk.AccAddressFromBech32(senderStr)
	// suite.Require().NoError(err)
	// receiver, err := sdk.AccAddressFromBech32(receiverStr)
	// suite.Require().NoError(err)

	ethPk, err := ethsecp256k1.GenerateKey()
	suite.Require().Nil(err)
	ethsecpAddr := sdk.AccAddress(ethPk.PubKey().Address())
	ethsecpAddrEvmos := sdk.AccAddress(ethPk.PubKey().Address()).String()
	ethsecpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, ethsecpAddr)

	timeoutHeight := clienttypes.NewHeight(0, 100)
	disabledTimeoutTimestamp := uint64(0)
	mockPacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	packet := mockPacket
	expAck := ibcmock.MockAcknowledgement

	testCases := []struct {
		name        string
		malleate    func()
		ackSuccess  bool
		expWithdraw bool
	}{
		{
			"continue - params disabled",
			func() {
				params := suite.app.WithdrawKeeper.GetParams(suite.ctx)
				params.EnableWithdraw = false
				suite.app.WithdrawKeeper.SetParams(suite.ctx, params)
			},
			true,
			false,
		},
		{
			"continue - destination channel not authorized",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-100", timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"continue - destination channel is EVM",
			func() {
				EVMChannels := suite.app.ClaimsKeeper.GetParams(suite.ctx).EVMChannels
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, EVMChannels[0], timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"fail - non ics20 packet",
			func() {
				packet = mockPacket
			},
			false,
			false,
		},
		{
			"fail - invalid sender - missing '1' ",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "evmos", receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"fail - invalid sender - invalid bech32",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", "badba1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms", receiverStr)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"fail - invalid recipient",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", senderStr, "badbadhf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625")
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
			},
			false,
			false,
		},
		{
			"continue - case 1: sender and receiver address are not the same",
			func() {
				pk1 := secp256k1.GenPrivKey()
				otherSecpAddrEvmos := sdk.AccAddress(pk1.PubKey().Address()).String()

				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, otherSecpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"continue - case 2: receiver pubkey is a supported key",
			func() {
				// Set account to generate a pubkey
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(ethsecpAddr, ethPk.PubKey(), 0, 0))

				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", ethsecpAddrCosmos, ethsecpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
			},
			true,
			false,
		},
		{
			"withdraw - same sender with EVM channel, no claims record",
			func() {
				transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", secpAddrCosmos, secpAddrEvmos)
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, 0)
			},
			true, // TODO is this correct?
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

			testutil.FundAccount(suite.app.BankKeeper, suite.ctx, secpAddr, sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(1000))))
			fmt.Println(suite.app.BankKeeper.GetAllBalances(suite.ctx, secpAddr))

			ack := suite.app.WithdrawKeeper.OnRecvPacket(suite.ctx, packet, expAck)

			// Check ackknowledgement
			if tc.ackSuccess {
				suite.Require().Equal(expAck, ack)
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}

			// Check withdrawal
			if tc.expWithdraw {
				balances := suite.app.BankKeeper.GetAllBalances(suite.ctx, secpAddr)
				suite.Require().True(balances.IsZero())
			}
		})
	}
}
