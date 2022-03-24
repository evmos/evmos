package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/AstraProtocol/astraos/v1/app"
	ibctesting "github.com/AstraProtocol/astraos/v1/ibc/testing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"

	erc20types "github.com/tharsis/evmos/v3/x/erc20/types"
)

type IBCTestingSuite struct {
	suite.Suite
	coordinator *ibcgotesting.Coordinator

	// testing chains used for convenience and readability
	chainA      *ibcgotesting.TestChain // Evmos chain A
	chainB      *ibcgotesting.TestChain // Evmos chain B
	chainC      *ibcgotesting.TestChain // Evmos chain C
	chainCosmos *ibcgotesting.TestChain // Cosmos chain

	pathEVM    *ibcgotesting.Path // chainA (Evmos) <-->  chainB (Evmos)
	pathCosmos *ibcgotesting.Path // chainA (Evmos) <--> chainCosmos
}

func (suite *IBCTestingSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3, 1) // initializes 2 Evmos test chains and 1 Cosmos Chain
	suite.chainA = suite.coordinator.GetChain(ibcgotesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibcgotesting.GetChainID(2))
	suite.chainC = suite.coordinator.GetChain(ibcgotesting.GetChainID(3))
	suite.chainCosmos = suite.coordinator.GetChain(ibcgotesting.GetChainID(4))

	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)
	suite.coordinator.CommitNBlocks(suite.chainCosmos, 2)

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10000)))
	err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(), erc20types.ModuleName, coins)
	suite.Require().NoError(err)

	err = suite.chainA.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainA.GetContext(), erc20types.ModuleName, coins)
	suite.Require().NoError(err)

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

func (suite *IBCTestingSuite) Test2() {
	suite.SetupTest()
	sender := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"
	receiver := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	ibcDaemon := "ibc/8EAC8061F4499F03D2D1419A3E73D346289AE9DB89CAB1486B72539572B1915E"

	//validMetadata := banktypes.Metadata{
	//	Description: "description of the token",
	//	Base:        ibcDaemon,
	//	// NOTE: Denom units MUST be increasing
	//	DenomUnits: []*banktypes.DenomUnit{
	//		{
	//			Denom:    ibcDaemon,
	//			Exponent: 0,
	//		},
	//	},
	//	Name:    "IBC-USDT",
	//	Symbol:  "USDT",
	//	Display: ibcDaemon,
	//}

	err := suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(),
		erc20types.ModuleName, sdk.Coins{sdk.NewInt64Coin("aevmos", 1000)})
	suite.Require().NoError(err)

	err = suite.chainB.App.(*app.Evmos).BankKeeper.MintCoins(suite.chainB.GetContext(),
		erc20types.ModuleName, sdk.Coins{sdk.NewInt64Coin(ibcDaemon, 1)})
	suite.Require().NoError(err)

	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), sdk.AccAddress(sender), "aevmos")
	println(coin.Amount.Int64())

	coin = suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, ibcDaemon)
	suite.Require().Equal(coin.Amount.Int64(), int64(0))

	path := suite.pathEVM

	transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
	bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
	packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// send on endpointA
	err = path.EndpointA.SendPacket(packet)
	suite.Require().NoError(err)

	// receive on endpointB
	err = path.EndpointB.RecvPacket(packet)
	suite.Require().NoError(err)

	coin = suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), sdk.AccAddress(sender), "aevmos")
	println(coin.Amount.Int64())

	coin = suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, ibcDaemon)
	suite.Require().Equal(coin.Amount.Int64(), int64(100))

	//senderAddr := common.BytesToAddress(suite.chainB.SenderPrivKey.PubKey().Address().Bytes())

	//msg := erc20types.NewMsgConvertCoin(
	//	sdk.NewCoin(pair.Denom, sdk.NewInt(10)),
	//	senderAddr,
	//	sdk.AccAddress(suite.chainB.SenderPrivKey.PubKey().Address()),
	//)
	//ctx := sdk.WrapSDKContext(suite.chainB.GetContext())
	//
	//id, ok := suite.chainB.App.(*app.Evmos).Erc20Keeper.GetTokenPair(suite.chainB.GetContext(), pair.GetID())
	//println(ok)
	//println(id.Denom)
	//
	//res, err := suite.chainB.App.(*app.Evmos).Erc20Keeper.ConvertCoin(ctx, msg)
	//suite.Require().NoError(err)
	//println(res.String())
}
