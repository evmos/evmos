package keeper_test

import (
	"fmt"
	"strconv"
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

	path        *ibcgotesting.Path
	path2       *ibcgotesting.Path
	pathOutside *ibcgotesting.Path
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
	suite.pathOutside = NewTransferPath(suite.IBCChain2, suite.IBCChain)
	suite.coordinator.Setup(suite.path) // clientID, connectionID, channelID filled
	suite.coordinator.Setup(suite.path2)
	suite.coordinator.Setup(suite.pathOutside)
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

func (suite *IBCTestingSuite) SendAndReceiveMessage(path *ibcgotesting.Path, chain *ibcgotesting.TestChain, coin string, amount int64, sender string, receiver string, seq uint64) {
	// Send IBC transaction of 10 testcoin
	transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.NewCoin(coin, sdk.NewInt(amount)), sender, receiver, timeoutHeight, 0)
	_, err := chain.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed
	transfer := transfertypes.NewFungibleTokenPacketData(coin, strconv.Itoa(int(amount)), sender, receiver)
	packet := channeltypes.NewPacket(transfer.GetBytes(), seq, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)
	// Receive message on the evmos side, and send ack
	err = path.RelayPacket(packet)
	suite.Require().NoError(err)
}

func CreatePacket(amount, denom, sender, receiver, srcPort, srcChannel, dstPort, dstChannel string, seq, timeout uint64) channeltypes.Packet {
	transfer := transfertypes.FungibleTokenPacketData{
		Amount:   amount,
		Denom:    denom,
		Receiver: sender,
		Sender:   receiver,
	}
	return channeltypes.NewPacket(
		transfer.GetBytes(),
		seq,
		srcPort,
		srcChannel,
		dstPort,
		dstChannel,
		clienttypes.ZeroHeight(), // timeout height disabled
		timeout,                  // timeout timestamp disabled
	)
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
				senderAcc, err := sdk.AccAddressFromBech32(sender)
				suite.Require().NoError(err)
				receiverAcc, err := sdk.AccAddressFromBech32(receiver)
				suite.Require().NoError(err)

				// Aevmos were escrowed
				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), transfertypes.GetEscrowAddress("transfer", "channel-0"), "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(10000)))
				// ibccoins were burn
				coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), receiverAcc, testcoinIbcdenom)
				suite.Require().Equal(coin.Amount, sdk.NewInt(0))

				// Recreate packets that were sent in the ibc_callback
				packet2 := CreatePacket("10000", "aevmos", sender, receiver,
					"transfer", "channel-0", "transfer", "channel-0", 1, timeout)

				packet3 := CreatePacket("10", "transfer/channel-0/testcoin", sender, receiver,
					"transfer", "channel-0", "transfer", "channel-0", 2, timeout)

				// Relay both packets that were sent in the ibc_callback
				err = suite.path.RelayPacket(packet2)
				suite.Require().NoError(err)
				err = suite.path.RelayPacket(packet3)
				suite.Require().NoError(err)

				coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
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
			suite.SendAndReceiveMessage(path, suite.IBCChain, "testcoin", 10, sender, receiver, 1)
			timeout = uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

			tc.test()
		})
	}
}

// Send IBC-Coin2 from IBC-Chain2 to Evmos
// Enable Withdraw
// Send IBC-Coin1 From IBC-Chain1 to Evmos
// Aevmos, IBC-Coin1 should be on IBC-Chain1 balance
// IBC-Coin2 should remain on the EvmosChain
// Send IBC-Coin1 From IBC-Chain1 to Evmos
// No changes on balance should occur
func (suite *IBCTestingSuite) TestTwoChains() {
	suite.SetupTest()
	testcoin2Denomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "testcoin2",
	}
	testcoin2Ibcdenom := testcoin2Denomtrace.IBCDenom()

	aevmosDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom := aevmosDenomtrace.IBCDenom()

	params := types.DefaultParams()
	params.EnableWithdraw = false
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	sender := suite.IBCChain.SenderAccount.GetAddress().String()
	receiver := suite.IBCChain.SenderAccount.GetAddress().String()
	pathCosmos := suite.path
	pathExtra := suite.path2

	suite.SendAndReceiveMessage(pathExtra, suite.IBCChain2, "testcoin2", 10, suite.IBCChain2.SenderAccount.GetAddress().String(), receiver, 1)

	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.SendAndReceiveMessage(pathCosmos, suite.IBCChain, "testcoin", 10, sender, receiver, 1)
	timeout := uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

	// Recreate packets that were sent in the ibc_callback
	// Coins locked
	packet2 := CreatePacket("10000", "aevmos", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 1, timeout)

	packet3 := CreatePacket("10", "transfer/channel-0/testcoin", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 2, timeout)

	// Relay both packets that were sent in the ibc_callback
	err := suite.path.RelayPacket(packet2)
	suite.Require().NoError(err)
	err = suite.path.RelayPacket(packet3)
	suite.Require().NoError(err)

	senderAcc, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAcc, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	// Aevmos was withdrawn from user address
	coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))

	// Check that the coin
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, testcoin2Ibcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(testcoin2Ibcdenom, sdk.NewInt(10)))

	// Aevmos was withdrawn from user address and is available on IBCChain
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))

	// Coins used for withdraw were recovered
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, "testcoin")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))

	// Send IBC transaction of 10 testcoin
	suite.SendAndReceiveMessage(pathCosmos, suite.IBCChain, "testcoin", 10, sender, receiver, 2)
	timeout = uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

	packet4 := CreatePacket("10", "transfer/channel-0/testcoin", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 3, timeout)

	err = pathCosmos.RelayPacket(packet4)
	suite.Require().NoError(err)

	// Aevmos was withdrawn from user address
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))

	// Check that the coin
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, testcoin2Ibcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(testcoin2Ibcdenom, sdk.NewInt(10)))

	// Aevmos was withdrawn from user address and is available on IBCChain
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))

	// Coins used for withdraw were recovered
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, "testcoin")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))

}

// Send IBC-Coin1 from IBC-Chain1 to IBC-Chain2
// Send IBC-Coin1 From IBC-Chain2 to Evmos
// Enable Withdraw
// Send IBC-Coin2 from IBC-Chain2 to Evmos
// Aevmos, IBC-Coin1 and IBC-Coin2 should be on IBC-Chain2 balance
func (suite *IBCTestingSuite) TestTwoChainsSendNonNativeCoin() {
	suite.SetupTest()
	testcoin2Denomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "testcoin2",
	}
	testcoin2Ibcdenom := testcoin2Denomtrace.IBCDenom()

	aevmosDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom := aevmosDenomtrace.IBCDenom()

	params := types.DefaultParams()
	params.EnableWithdraw = false
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	sender := suite.IBCChain.SenderAccount.GetAddress().String()
	receiver := suite.IBCChain.SenderAccount.GetAddress().String()

	pathCosmos := suite.path
	pathOutside := suite.pathOutside

	suite.SendAndReceiveMessage(pathOutside, suite.IBCChain2, "testcoin2", 10, suite.IBCChain2.SenderAccount.GetAddress().String(), receiver, 1)

	// Send IBC transaction of 10 testcoin
	transferMsg := transfertypes.NewMsgTransfer(pathCosmos.EndpointA.ChannelConfig.PortID, pathCosmos.EndpointA.ChannelID, sdk.NewCoin(testcoin2Ibcdenom, sdk.NewInt(10)), sender, receiver, timeoutHeight, 0)
	_, err := suite.IBCChain.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed
	transfer := transfertypes.NewFungibleTokenPacketData("transfer/channel-1/testcoin2", "10", sender, receiver)
	packet := channeltypes.NewPacket(transfer.GetBytes(), 1, pathCosmos.EndpointA.ChannelConfig.PortID, pathCosmos.EndpointA.ChannelID, pathCosmos.EndpointB.ChannelConfig.PortID, pathCosmos.EndpointB.ChannelID, timeoutHeight, 0)
	// Receive message on the evmos side, and send ack
	err = pathCosmos.RelayPacket(packet)
	suite.Require().NoError(err)

	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.SendAndReceiveMessage(pathCosmos, suite.IBCChain, "testcoin", 10, sender, receiver, 2)
	timeout := uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

	// Recreate packets that were sent in the ibc_callback
	// Coins locked
	packet2 := CreatePacket("10000", "aevmos", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 1, timeout)

	packet3 := CreatePacket("10", "transfer/channel-0/testcoin", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 2, timeout)

	packet4 := CreatePacket("10", "transfer/channel-0/transfer/channel-1/testcoin2", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 3, timeout)

	// Relay packets that were sent in the ibc_callback
	err = suite.path.RelayPacket(packet2)
	suite.Require().NoError(err)
	err = suite.path.RelayPacket(packet3)
	suite.Require().NoError(err)
	err = suite.path.RelayPacket(packet4)
	suite.Require().NoError(err)

	senderAcc, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAcc, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	// Aevmos was withdrawn from user address
	coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))

	// Check that the coin
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), senderAcc, testcoin2Ibcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(testcoin2Ibcdenom, sdk.NewInt(10)))

	// Aevmos was withdrawn from user address and is available on IBCChain
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))

	// Coins used for withdraw were recovered
	coin = suite.IBCChain.GetSimApp().BankKeeper.GetBalance(suite.IBCChain.GetContext(), receiverAcc, "testcoin")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))
}
