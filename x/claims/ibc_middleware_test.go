package claims_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v3/testing"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/v3/app"
	ibctesting "github.com/tharsis/evmos/v3/ibc/testing"
	"github.com/tharsis/evmos/v3/testutil"
	"github.com/tharsis/evmos/v3/x/claims/types"
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

	err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
	suite.Require().NoError(err)

	suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), addr, claimsRecord)

	err = testutil.FundModuleAccount(suite.chainA.App.(*app.Evmos).BankKeeper, suite.chainA.GetContext(), types.ModuleName, coins)
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

func (suite *IBCTestingSuite) TestOnRecvPacket() {
	var path *ibcgotesting.Path

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default",
			func() {},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			timeoutHeight := clienttypes.NewHeight(0, 100)

			path = ibctesting.NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibcgotesting.FirstChannelID

			module, _, err := suite.chainA.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(suite.chainA.GetContext(), ibcgotesting.TransferPort)
			suite.Require().NoError(err)

			ibcModule, ok := suite.chainA.App.GetIBCKeeper().Router.GetRoute(module)
			suite.Require().True(ok)

			tc.malleate()

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String())
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			ack := ibcModule.OnRecvPacket(suite.chainA.GetContext(), packet, suite.chainA.SenderAccount.GetAddress())

			if tc.expPass {
				suite.Require().True(ack.Success())
			} else {
				suite.Require().False(ack.Success())
			}
		})
	}
}
