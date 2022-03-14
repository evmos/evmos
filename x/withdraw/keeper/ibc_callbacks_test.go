package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"

	"github.com/tharsis/evmos/v2/ibctesting"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/tharsis/evmos/v2/app"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"
	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	EvmosChain *ibcgotesting.TestChain
	IBCChain   *ibcgotesting.TestChain
	IBCChain2  *ibcgotesting.TestChain

	path  *ibcgotesting.Path
	path2 *ibcgotesting.Path

	sender    string
	senderAcc sdk.AccAddress
}

func (suite *IBCTestingSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1, 2)            // initializes 2 test chains
	suite.EvmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(1)) // convenience and readability
	suite.IBCChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(2))   // convenience and readability
	suite.IBCChain2 = suite.coordinator.GetChain(ibcgotesting.GetChainID(3))  // convenience and readability
	suite.coordinator.CommitNBlocks(suite.EvmosChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCChain2, 2)

	// Mint coins locked on the evmos account generated with secp.
	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10000)))
	err := suite.EvmosChain.App.(*app.Evmos).BankKeeper.MintCoins(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.EvmosChain.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToAccount(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, suite.IBCChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the cosmos side which we'll use to unlock our aevmos
	coins = sdk.NewCoins(sdk.NewCoin("testcoin", sdk.NewInt(10)))
	err = suite.IBCChain.GetSimApp().BankKeeper.MintCoins(suite.IBCChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCChain.GetContext(), minttypes.ModuleName, suite.IBCChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the cosmos side which we'll use to unlock our aevmos
	coins = sdk.NewCoins(sdk.NewCoin("testcoin2", sdk.NewInt(10)))
	err = suite.IBCChain2.GetSimApp().BankKeeper.MintCoins(suite.IBCChain2.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCChain2.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCChain2.GetContext(), minttypes.ModuleName, suite.IBCChain2.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	params := types.DefaultParams()
	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.path = NewTransferPath(suite.IBCChain, suite.EvmosChain) // clientID, connectionID, channelID empty
	suite.path2 = NewTransferPath(suite.IBCChain2, suite.EvmosChain)
	suite.coordinator.Setup(suite.path) // clientID, connectionID, channelID filled
	suite.coordinator.Setup(suite.path2)
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

func (suite *IBCTestingSuite) TestOnReceiveWithdraw() {
	var (
		sender   string
		receiver string
		timeout  uint64
	)

	testcoinDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "testcoin",
	}
	testcoinIbcdenom := testcoinDenomtrace.IBCDenom()

	aevmosDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom := aevmosDenomtrace.IBCDenom()

	testCases := []struct {
		name     string
		malleate func()
		test     func()
	}{
		{
			"correct execution",
			func() {
				// TODO Change IBCChain Bech32 to Cosmos prefix
				// sender := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, suite.IBCChain.SenderAccount.GetAddress())
				sender = suite.IBCChain.SenderAccount.GetAddress().String()
				receiver = suite.IBCChain.SenderAccount.GetAddress().String()
			},
			func() {
				// Recreate packets that were sent in the ibc_callback

				// Coins locked
				transfer2 := transfertypes.FungibleTokenPacketData{
					Amount:   "10000",
					Denom:    "aevmos",
					Receiver: sender,
					Sender:   receiver,
				}
				packet2 := channeltypes.NewPacket(
					transfer2.GetBytes(),
					1,
					"transfer",
					"channel-0",
					"transfer",
					"channel-0",
					clienttypes.ZeroHeight(), // timeout height disabled
					timeout,                  // timeout timestamp disabled
				)

				// Coins transfered
				transfer3 := transfertypes.FungibleTokenPacketData{
					Amount:   "10",
					Denom:    "transfer/channel-0/testcoin",
					Receiver: sender,
					Sender:   receiver,
				}
				packet3 := channeltypes.NewPacket(
					transfer3.GetBytes(),
					2,
					"transfer",
					"channel-0",
					"transfer",
					"channel-0",
					clienttypes.ZeroHeight(), // timeout height disabled
					timeout,                  // timeout timestamp disabled
				)

				// Relay both packets that were sent in the ibc_callback
				err := suite.path.RelayPacket(packet2)
				suite.Require().NoError(err)
				err = suite.path.RelayPacket(packet3)
				suite.Require().NoError(err)

				senderAcc, err := sdk.AccAddressFromBech32(sender)
				suite.Require().NoError(err)
				receiverAcc, err := sdk.AccAddressFromBech32(receiver)
				suite.Require().NoError(err)

				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))
				coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, aevmosIbcdenom)
				suite.Require().Equal(coin.Amount, sdk.NewInt(10000))
				coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, "testcoin")
				suite.Require().Equal(coin.Amount, sdk.NewInt(10))
			},
		},
		{
			"Disabled params",
			func() {
				sender = suite.IBCChain.SenderAccount.GetAddress().String()
				receiver = suite.IBCChain.SenderAccount.GetAddress().String()

				params := types.DefaultParams()
				params.EnableWithdraw = false
				suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)
			},
			func() {
				senderAcc, err := sdk.AccAddressFromBech32(sender)
				suite.Require().NoError(err)
				receiverAcc, err := sdk.AccAddressFromBech32(receiver)
				suite.Require().NoError(err)

				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(10000)))
				coins := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), receiverAcc, testcoinIbcdenom)
				suite.Require().Equal(coins.Amount, sdk.NewInt(10))
			},
		},
		{
			"Different Addresses",
			func() {
				sender = suite.IBCChain.SenderAccount.GetAddress().String()
				receiver = suite.EvmosChain.SenderAccount.GetAddress().String()
			},
			func() {
				senderAcc, err := sdk.AccAddressFromBech32(sender)
				suite.Require().NoError(err)
				receiverAcc, err := sdk.AccAddressFromBech32(receiver)
				suite.Require().NoError(err)

				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(10000)))
				coins := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), receiverAcc, testcoinIbcdenom)
				suite.Require().Equal(coins.Amount, sdk.NewInt(10))
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.path

			tc.malleate()

			// Send IBC transaction of 10 testcoin
			transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.NewCoin("testcoin", sdk.NewInt(10)), sender, receiver, timeoutHeight, 0)
			_, err := suite.IBCChain.SendMsgs(transferMsg)
			suite.Require().NoError(err) // message committed
			timeout = uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour).Add(time.Second * 10).UnixNano())

			transfer := transfertypes.NewFungibleTokenPacketData("testcoin", "10", sender, receiver)
			packet := channeltypes.NewPacket(transfer.GetBytes(), 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// Receive message on the evmos side, and send ack
			err = suite.path.RelayPacket(packet)
			suite.Require().NoError(err)

			tc.test()
		})
	}
}

func (suite *IBCTestingSuite) TestTwoChains() {
	testcoinDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "testcoin2",
	}
	testcoinIbcdenom := testcoinDenomtrace.IBCDenom()

	aevmosDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom := aevmosDenomtrace.IBCDenom()

	params := types.DefaultParams()
	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	sender := suite.IBCChain.SenderAccount.GetAddress().String()
	receiver := suite.IBCChain.SenderAccount.GetAddress().String()

	pathCosmos := suite.path
	pathExtra := suite.path2

	// Send IBC transaction of 10 testcoin
	transferMsg := transfertypes.NewMsgTransfer(pathExtra.EndpointA.ChannelConfig.PortID, pathExtra.EndpointA.ChannelID, sdk.NewCoin("testcoin2", sdk.NewInt(10)), suite.IBCChain2.SenderAccount.GetAddress().String(), receiver, timeoutHeight, 0)
	_, err := suite.IBCChain2.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed

	transfer := transfertypes.NewFungibleTokenPacketData("testcoin2", "10", suite.IBCChain2.SenderAccount.GetAddress().String(), receiver)
	packet := channeltypes.NewPacket(transfer.GetBytes(), 1, pathExtra.EndpointA.ChannelConfig.PortID, pathExtra.EndpointA.ChannelID, pathExtra.EndpointB.ChannelConfig.PortID, pathExtra.EndpointB.ChannelID, timeoutHeight, 0)
	// Receive message on the evmos side, and send ack
	err = pathExtra.RelayPacket(packet)
	suite.Require().NoError(err)

	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	// Send IBC transaction of 10 testcoin
	transferMsg = transfertypes.NewMsgTransfer(pathCosmos.EndpointA.ChannelConfig.PortID, pathCosmos.EndpointA.ChannelID, sdk.NewCoin("testcoin", sdk.NewInt(10)), sender, receiver, timeoutHeight, 0)
	_, err = suite.IBCChain.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed

	timeout := uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour).Add(time.Second * 10).UnixNano())

	transfer = transfertypes.NewFungibleTokenPacketData("testcoin", "10", sender, receiver)
	packet = channeltypes.NewPacket(transfer.GetBytes(), 1, pathCosmos.EndpointA.ChannelConfig.PortID, pathCosmos.EndpointA.ChannelID, pathCosmos.EndpointB.ChannelConfig.PortID, pathCosmos.EndpointB.ChannelID, timeoutHeight, 0)

	// Receive message on the evmos side, and send ack
	err = pathCosmos.RelayPacket(packet)
	suite.Require().NoError(err)

	// Recreate packets that were sent in the ibc_callback

	// Coins locked
	transfer2 := transfertypes.FungibleTokenPacketData{
		Amount:   "10000",
		Denom:    "aevmos",
		Receiver: sender,
		Sender:   receiver,
	}
	packet2 := channeltypes.NewPacket(
		transfer2.GetBytes(),
		1,
		"transfer",
		"channel-0",
		"transfer",
		"channel-0",
		clienttypes.ZeroHeight(), // timeout height disabled
		timeout,                  // timeout timestamp disabled
	)

	// Coins transfered
	transfer3 := transfertypes.FungibleTokenPacketData{
		Amount:   "10",
		Denom:    "transfer/channel-0/testcoin",
		Receiver: sender,
		Sender:   receiver,
	}
	packet3 := channeltypes.NewPacket(
		transfer3.GetBytes(),
		2,
		"transfer",
		"channel-0",
		"transfer",
		"channel-0",
		clienttypes.ZeroHeight(), // timeout height disabled
		timeout,                  // timeout timestamp disabled
	)

	// Relay both packets that were sent in the ibc_callback
	err = suite.path.RelayPacket(packet2)
	suite.Require().NoError(err)
	err = suite.path.RelayPacket(packet3)
	suite.Require().NoError(err)

	senderAcc, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAcc, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, testcoinIbcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(testcoinIbcdenom, sdk.NewInt(10)))
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, "testcoin")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))

}
