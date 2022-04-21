package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
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

func (suite *IBCTestingSuite) TestOnAcknowledgementPacketIBC() {
	sender := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	receiver := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)

	testCases := []struct {
		name            string
		malleate        func(int64)
		claimableAmount int64
		expectedBalance int64
		expPass         bool
	}{
		{
			"correct execution - Claimable Transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.NewClaimsRecord(amt))
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainA.App.(*app.Evmos).BankKeeper, suite.chainA.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimable Transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.NewClaimsRecord(amt))
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainA.App.(*app.Evmos).BankKeeper, suite.chainA.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Claimed transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainA.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, true}})

				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainA.App.(*app.Evmos).BankKeeper, suite.chainA.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			4,
			0,
			true,
		},
		{
			"Disabled by params",
			func(_ int64) {
				params := types.DefaultParams()
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainA.GetContext(), params)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)
			},
			0,
			0,
			false,
		},
		{
			"No claim record",
			func(claimableAmount int64) {
			},
			0,
			0,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.pathEVM

			tc.malleate(tc.claimableAmount)

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", "100", sender, receiver)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			err := path.EndpointA.SendPacket(packet)
			suite.Require().NoError(err)

			err = path.RelayPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
				claim, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderAddr)
				suite.Require().True(found)
				suite.Require().Equal(claim.InitialClaimableAmount, sdk.NewInt(4))
			} else {
				coin := suite.chainA.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainA.GetContext(), senderAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
				_, found := suite.chainA.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainA.GetContext(), senderAddr)
				suite.Require().False(found)
			}

		})
	}
}

func (suite *IBCTestingSuite) TestOnRecvPacketIBC() {
	sender := "evmos1hf0468jjpe6m6vx38s97z2qqe8ldu0njdyf625"
	receiver := "evmos1sv9m0g7ycejwr3s369km58h5qe7xj77hvcxrms"
	triggerAmt := types.IBCTriggerAmt

	senderAddr, err := sdk.AccAddressFromBech32(sender)
	suite.Require().NoError(err)
	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	suite.Require().NoError(err)

	testCases := []struct {
		name            string
		malleate        func(int64)
		additionalTest  func()
		claimableAmount int64
		expectedBalance int64
		expPass         bool
	}{
		{
			"Deactivate claims params",
			func(_ int64) {
				params := types.DefaultParams()
				params.EnableClaims = false
				suite.chainA.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainA.GetContext(), params)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.chainB.GetContext(), params)
			},
			func() {},
			0,
			0,
			false,
		},
		{
			"correct execution - Claimable Transfer - Sender Record - No recipient Record",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.NewClaimsRecord(amt))

				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {
				// Check sender claim was deleted
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().False(found)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Already claimed transfer - Sender Record - No recipient Record",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(claimableAmount/4)))

				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, true, true, true}})
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {
				// Check sender claim was not deleted
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().True(found)
			},
			4,
			0,
			false,
		},
		{
			"correct execution - 0 claimable amount  - Sender Record - No recipient Record",
			func(_ int64) {
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.ZeroInt(), ActionsCompleted: []bool{false, false, false, false}})
			},
			func() {
				// Check sender claim was not deleted
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().True(found)
			},
			0,
			0,
			false,
		},
		{
			"correct execution - claimable amount   - Sender Record - No recipient Record - prevent migration if IBC action is not completed.",
			func(claimableAmount int64) {

				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(claimableAmount/4)))
				//set claims record - set IBC Transfer action to not done i.e. false
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, false}})
				_ = testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)

			},
			func() {
				// Check sender claim record is available
				senderClaimsRecord, foundErr := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)

				//try to migrate records
				if foundErr == true {
					suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, senderClaimsRecord)
					suite.chainB.App.(*app.Evmos).ClaimsKeeper.DeleteClaimsRecord(suite.chainB.GetContext(), senderAddr)
				}

				//Perform check
				//Check that sender claim record is available. if not, fail.
				senderClaimsRecord, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().True(found)
			},
			4,
			1,
			true,
		},
		{
			"correct execution - Merge Transfer -  Sender Record - Recipient Record",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt.Add(amt.QuoRaw(2))))

				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, false, false, false}})
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, true, true, false}})

				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {
				// Check sender claim was deleted after merge
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().False(found)
			},
			4,
			4,
			true,
		},
		{
			"correct execution - Sender Already Claimed Transfer -  Sender Record - Recipient Record",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))

				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, false, false, true}})
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, false}})

				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {
				// Check sender claim was not deleted after merge
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().True(found)
			},
			0,
			0,
			true,
		},
		{
			"correct execution - Sender Has no claimable Amount -  Sender Record - Recipient Record",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt.QuoRaw(2)))

				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), senderAddr, types.ClaimsRecord{InitialClaimableAmount: sdk.ZeroInt(), ActionsCompleted: []bool{false, false, false, false}})
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, true, true, false}})

				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {
				// Check sender claim was deleted after merge
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), senderAddr)
				suite.Require().False(found)
			},
			0,
			0,
			true,
		},
		{
			"correct execution - Recipient Claimable",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, false, false, false}})

				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {},
			4,
			1,
			true,
		},
		{
			"correct execution - Recipient Claimed ibc transfer",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{true, true, true, true}})
			},
			func() {},
			4,
			0,
			true,
		},
		{
			"correct execution - Recipient has no claimable",
			func(claimableAmount int64) {
				amt := sdk.NewInt(claimableAmount)
				suite.chainB.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(suite.chainB.GetContext(), receiverAddr, types.ClaimsRecord{InitialClaimableAmount: amt, ActionsCompleted: []bool{false, false, false, false}})

				coins := sdk.NewCoins(sdk.NewCoin("aevmos", amt))
				// update the escrowed account balance to maintain the invariant
				err := testutil.FundModuleAccount(suite.chainB.App.(*app.Evmos).BankKeeper, suite.chainB.GetContext(), types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			func() {},
			0,
			0,
			true,
		},
		{
			"No claim record",
			func(claimableAmount int64) {
			},
			func() {},
			0,
			0,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			path := suite.pathEVM

			tc.malleate(tc.claimableAmount)

			transfer := transfertypes.NewFungibleTokenPacketData("aevmos", triggerAmt, sender, receiver)
			bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
			packet := channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

			// send on endpointA
			path.EndpointA.SendPacket(packet)

			// receive on endpointB
			err := path.EndpointB.RecvPacket(packet)
			suite.Require().NoError(err)

			if tc.expPass {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiverAddr)
				suite.Require().True(found)
			} else {
				coin := suite.chainB.App.(*app.Evmos).BankKeeper.GetBalance(suite.chainB.GetContext(), receiverAddr, "aevmos")
				suite.Require().Equal(coin.String(), sdk.NewCoin("aevmos", sdk.NewInt(tc.expectedBalance)).String())
				_, found := suite.chainB.App.(*app.Evmos).ClaimsKeeper.GetClaimsRecord(suite.chainB.GetContext(), receiverAddr)
				suite.Require().False(found)
			}

			tc.additionalTest()
		})
	}
}
