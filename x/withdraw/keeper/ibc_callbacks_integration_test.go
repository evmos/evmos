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
	claimtypes "github.com/tharsis/evmos/v2/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"
	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	EvmosChain      *ibcgotesting.TestChain
	IBCOsmosisChain *ibcgotesting.TestChain
	IBCCosmosChain  *ibcgotesting.TestChain

	pathOsmosisEvmos  *ibcgotesting.Path
	pathCosmosEvmos   *ibcgotesting.Path
	pathOsmosisCosmos *ibcgotesting.Path
}

func (suite *IBCTestingSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1, 2)                 // initializes 2 test chains
	suite.EvmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(1))      // convenience and readability
	suite.IBCOsmosisChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(2)) // convenience and readability
	suite.IBCCosmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(3))  // convenience and readability
	suite.coordinator.CommitNBlocks(suite.EvmosChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCOsmosisChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCCosmosChain, 2)

	// Mint coins locked on the evmos account generated with secp.
	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10000)))
	err := suite.EvmosChain.App.(*app.Evmos).BankKeeper.MintCoins(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.EvmosChain.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToAccount(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, suite.IBCOsmosisChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the osmosis side which we'll use to unlock our aevmos
	coins = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10)))
	err = suite.IBCOsmosisChain.GetSimApp().BankKeeper.MintCoins(suite.IBCOsmosisChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCOsmosisChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCOsmosisChain.GetContext(), minttypes.ModuleName, suite.IBCOsmosisChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the cosmos side which we'll use to unlock our aevmos
	coins = sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(10)))
	err = suite.IBCCosmosChain.GetSimApp().BankKeeper.MintCoins(suite.IBCCosmosChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCCosmosChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCCosmosChain.GetContext(), minttypes.ModuleName, suite.IBCCosmosChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	claimparams := claimtypes.DefaultParams()
	claimparams.AirdropStartTime = suite.EvmosChain.GetContext().BlockTime()
	claimparams.EnableClaims = true
	suite.EvmosChain.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.EvmosChain.GetContext(), claimparams)

	params := types.DefaultParams()
	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.pathOsmosisEvmos = NewTransferPath(suite.IBCOsmosisChain, suite.EvmosChain) // clientID, connectionID, channelID empty
	suite.pathCosmosEvmos = NewTransferPath(suite.IBCCosmosChain, suite.EvmosChain)
	suite.pathOsmosisCosmos = NewTransferPath(suite.IBCCosmosChain, suite.IBCOsmosisChain)
	suite.coordinator.Setup(suite.pathOsmosisEvmos) // clientID, connectionID, channelID filled
	suite.coordinator.Setup(suite.pathCosmosEvmos)
	suite.coordinator.Setup(suite.pathOsmosisCosmos)
	suite.Require().Equal("07-tendermint-0", suite.pathOsmosisEvmos.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.pathOsmosisEvmos.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.pathOsmosisEvmos.EndpointA.ChannelID)
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
	// Send IBC transaction of 10 uosmo
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

	uosmoDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uosmo",
	}
	uosmoIbcdenom := uosmoDenomtrace.IBCDenom()

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
				// TODO Change IBCOsmosisChain Bech32 to Cosmos prefix
				// sender := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, suite.IBCOsmosisChain.SenderAccount.GetAddress())
				sender = suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
				receiver = suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
			},
			func() {
				senderAcc, err := sdk.AccAddressFromBech32(sender)
				suite.Require().NoError(err)
				receiverAcc, err := sdk.AccAddressFromBech32(receiver)
				suite.Require().NoError(err)

				// Aevmos were escrowed
				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), transfertypes.GetEscrowAddress("transfer", "channel-0"), "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(10000)))
				// ibccoins were burned
				coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
				suite.Require().Equal(coin.Amount, sdk.NewInt(0))

				// Recreate packets that were sent in the ibc_callback
				packet2 := CreatePacket("10000", "aevmos", sender, receiver,
					"transfer", "channel-0", "transfer", "channel-0", 1, timeout)

				packet3 := CreatePacket("10", "transfer/channel-0/uosmo", sender, receiver,
					"transfer", "channel-0", "transfer", "channel-0", 2, timeout)

				// Relay both packets that were sent in the ibc_callback
				err = suite.pathOsmosisEvmos.RelayPacket(packet2)
				suite.Require().NoError(err)
				err = suite.pathOsmosisEvmos.RelayPacket(packet3)
				suite.Require().NoError(err)

				coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))
				coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, aevmosIbcdenom)
				suite.Require().Equal(coin.Amount, sdk.NewInt(10000))
				coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
				suite.Require().Equal(coin.Amount, sdk.NewInt(10))
			},
		},
		{
			"Disabled params",
			func() {
				sender = suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
				receiver = suite.IBCOsmosisChain.SenderAccount.GetAddress().String()

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
				coins := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
				suite.Require().Equal(coins.Amount, sdk.NewInt(10))
			},
		},
		{
			"Different Addresses",
			func() {
				sender = suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
				receiver = suite.EvmosChain.SenderAccount.GetAddress().String()
			},
			func() {
				senderAcc, err := sdk.AccAddressFromBech32(sender)
				suite.Require().NoError(err)
				receiverAcc, err := sdk.AccAddressFromBech32(receiver)
				suite.Require().NoError(err)

				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(10000)))
				coins := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
				suite.Require().Equal(coins.Amount, sdk.NewInt(10))
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.pathOsmosisEvmos

			tc.malleate()
			suite.SendAndReceiveMessage(path, suite.IBCOsmosisChain, "uosmo", 10, sender, receiver, 1)
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
	uatomDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "uatom",
	}
	uatomIbcdenom := uatomDenomtrace.IBCDenom()

	aevmosDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom := aevmosDenomtrace.IBCDenom()

	params := types.DefaultParams()
	params.EnableWithdraw = false
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	sender := suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
	receiver := suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
	pathOsmosisEvmos := suite.pathOsmosisEvmos
	pathCosmosEvmos := suite.pathCosmosEvmos

	suite.SendAndReceiveMessage(pathCosmosEvmos, suite.IBCCosmosChain, "uatom", 10, suite.IBCCosmosChain.SenderAccount.GetAddress().String(), receiver, 1)

	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.SendAndReceiveMessage(pathOsmosisEvmos, suite.IBCOsmosisChain, "uosmo", 10, sender, receiver, 1)
	timeout := uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

	// Recreate packets that were sent in the ibc_callback
	// Coins locked
	packet2 := CreatePacket("10000", "aevmos", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 1, timeout)

	packet3 := CreatePacket("10", "transfer/channel-0/uosmo", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 2, timeout)

	// Relay both packets that were sent in the ibc_callback
	err := pathOsmosisEvmos.RelayPacket(packet2)
	suite.Require().NoError(err)
	err = pathOsmosisEvmos.RelayPacket(packet3)
	suite.Require().NoError(err)

	senderAcc, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAcc, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	// Aevmos was withdrawn from user address
	coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))

	// Check that the coin
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, uatomIbcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(uatomIbcdenom, sdk.NewInt(10)))

	// Aevmos was withdrawn from user address and is available on IBCOsmosisChain
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))

	// Coins used for withdraw were recovered
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))

	// Send IBC transaction of 10 uosmo
	suite.SendAndReceiveMessage(pathOsmosisEvmos, suite.IBCOsmosisChain, "uosmo", 10, sender, receiver, 2)
	timeout = uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

	packet4 := CreatePacket("10", "transfer/channel-0/uosmo", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 3, timeout)

	err = pathOsmosisEvmos.RelayPacket(packet4)
	suite.Require().NoError(err)

	// Aevmos was withdrawn from user address
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))

	// Check that the coin
	coin = suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, uatomIbcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(uatomIbcdenom, sdk.NewInt(10)))

	// Aevmos was withdrawn from user address and is available on IBCOsmosisChain
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))

	// Coins used for withdraw were recovered
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))

}

// Send IBC-Coin1 from IBC-Chain1 to IBC-Chain2
// Send IBC-Coin1 From IBC-Chain2 to Evmos
// Enable Withdraw
// Send IBC-Coin2 from IBC-Chain2 to Evmos
// Aevmos, IBC-Coin1 and IBC-Coin2 should be on IBC-Chain2 balance
func (suite *IBCTestingSuite) TestTwoChainsSendNonNativeCoin() {
	suite.SetupTest()
	uatomDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "uatom",
	}
	uatomIbcdenom := uatomDenomtrace.IBCDenom()

	aevmosDenomtrace := transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom := aevmosDenomtrace.IBCDenom()

	params := types.DefaultParams()
	params.EnableWithdraw = false
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	sender := suite.IBCOsmosisChain.SenderAccount.GetAddress().String()
	receiver := suite.IBCOsmosisChain.SenderAccount.GetAddress().String()

	pathOsmosisEvmos := suite.pathOsmosisEvmos
	pathOsmosisCosmos := suite.pathOsmosisCosmos

	suite.SendAndReceiveMessage(pathOsmosisCosmos, suite.IBCCosmosChain, "uatom", 10, suite.IBCCosmosChain.SenderAccount.GetAddress().String(), receiver, 1)

	// Send IBC transaction of 10 uosmo
	transferMsg := transfertypes.NewMsgTransfer(pathOsmosisEvmos.EndpointA.ChannelConfig.PortID, pathOsmosisEvmos.EndpointA.ChannelID, sdk.NewCoin(uatomIbcdenom, sdk.NewInt(10)), sender, receiver, timeoutHeight, 0)
	_, err := suite.IBCOsmosisChain.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed
	transfer := transfertypes.NewFungibleTokenPacketData("transfer/channel-1/uatom", "10", sender, receiver)
	packet := channeltypes.NewPacket(transfer.GetBytes(), 1, pathOsmosisEvmos.EndpointA.ChannelConfig.PortID, pathOsmosisEvmos.EndpointA.ChannelID, pathOsmosisEvmos.EndpointB.ChannelConfig.PortID, pathOsmosisEvmos.EndpointB.ChannelID, timeoutHeight, 0)
	// Receive message on the evmos side, and send ack
	err = pathOsmosisEvmos.RelayPacket(packet)
	suite.Require().NoError(err)

	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.SendAndReceiveMessage(pathOsmosisEvmos, suite.IBCOsmosisChain, "uosmo", 10, sender, receiver, 2)
	timeout := uint64(suite.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

	// Recreate packets that were sent in the ibc_callback
	// Coins locked
	packet2 := CreatePacket("10000", "aevmos", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 1, timeout)

	packet3 := CreatePacket("10", "transfer/channel-0/uosmo", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 2, timeout)

	packet4 := CreatePacket("10", "transfer/channel-0/transfer/channel-1/uatom", sender, receiver,
		"transfer", "channel-0", "transfer", "channel-0", 3, timeout)

	// Relay packets that were sent in the ibc_callback
	err = pathOsmosisEvmos.RelayPacket(packet2)
	suite.Require().NoError(err)
	err = pathOsmosisEvmos.RelayPacket(packet3)
	suite.Require().NoError(err)
	err = pathOsmosisEvmos.RelayPacket(packet4)
	suite.Require().NoError(err)

	senderAcc, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAcc, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	// Aevmos was withdrawn from user address
	coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), senderAcc, "aevmos")
	suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))

	// Check that the coin
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), senderAcc, uatomIbcdenom)
	suite.Require().Equal(coin, sdk.NewCoin(uatomIbcdenom, sdk.NewInt(10)))

	// Aevmos was withdrawn from user address and is available on IBCOsmosisChain
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, aevmosIbcdenom)
	suite.Require().Equal(coin.Amount, sdk.NewInt(10000))

	// Coins used for withdraw were recovered
	coin = suite.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(suite.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
	suite.Require().Equal(coin.Amount, sdk.NewInt(10))
}
