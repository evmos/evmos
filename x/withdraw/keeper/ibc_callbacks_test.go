package keeper_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

	path *ibcgotesting.Path

	sender    string
	senderAcc sdk.AccAddress
}

var s *IBCTestingSuite

func TestIBCTestingSuite(t *testing.T) {
	s = new(IBCTestingSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "IBC Testing Suite")
}

func (suite *IBCTestingSuite) SetupTest() {
	suite.coordinator = ibctesting.NewMixedCoordinator(suite.T(), 1, 1)       // initializes 2 test chains
	suite.EvmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(1)) // convenience and readability
	suite.IBCChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(2))   // convenience and readability
	suite.coordinator.CommitNBlocks(suite.EvmosChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCChain, 2)

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

	params := types.DefaultParams()
	params.EnableWithdraw = true
	suite.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(suite.EvmosChain.GetContext(), params)

	suite.path = NewTransferPath(suite.IBCChain, suite.EvmosChain) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(suite.path)                            // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", suite.path.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.path.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.path.EndpointA.ChannelID)
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

	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"correct execution",
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.path

			// TODO Change IBCChain Bech32 to Cosmos prefix
			// sender := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, suite.IBCChain.SenderAccount.GetAddress())
			sender := suite.IBCChain.SenderAccount.GetAddress().String()
			receiver := suite.IBCChain.SenderAccount.GetAddress().String()

			// Send IBC transaction of 10 testcoin
			transferMsg := transfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, sdk.NewCoin("testcoin", sdk.NewInt(10)), sender, receiver, timeoutHeight, 0)
			_, err := suite.IBCChain.SendMsgs(transferMsg)
			suite.Require().NoError(err) // message committed

			transfer := transfertypes.NewFungibleTokenPacketData("testcoin", "10", sender, receiver)
			packet := channeltypes.NewPacket(transfer.GetBytes(), 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// Receive message on the evmos side, and send ack
			err = suite.path.RelayPacket(packet)
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
				1677926229000000000,      // timeout timestamp disabled
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
				1677926229000000000,      // timeout timestamp disabled
			)

			// Relay both packets that were sent in the ibc_callback
			err = suite.path.RelayPacket(packet2)
			suite.Require().NoError(err)
			err = suite.path.RelayPacket(packet3)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(suite.EvmosChain.GetContext(), suite.IBCChain.SenderAccount.GetAddress(), "aevmos")
				suite.Require().Equal(coin, sdk.NewCoin("aevmos", sdk.NewInt(0)))
				coins := suite.IBCChain.GetSimApp().BankKeeper.GetAllBalances(suite.IBCChain.GetContext(), suite.IBCChain.SenderAccount.GetAddress())
				suite.Require().Equal(coins[0].Amount, sdk.NewInt(10000))
				suite.Require().Equal(coins[2].Amount, sdk.NewInt(10))
			}
		})
	}
}
