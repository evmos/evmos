package keeper_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/withdraw/types"
)

// The IBC Callback transfers all non-evmos balances from the receiver to the
// sender address for a receiver with a secp256k1 key
var _ = Describe("Performing a IBC transfer with enabled callback ", Ordered, func() {

	coin := sdk.NewCoin("aevmos", sdk.NewInt(100))

	var (
		sender   sdk.AccAddress
		receiver sdk.AccAddress
	)

	BeforeEach(func() {
		s.SetupTest()

		// Deactivate IBC callback
		params := types.DefaultParams()
		params.EnableWithdraw = false
		s.chainB.App.(*app.Evmos).WithdrawKeeper.SetParams(s.chainB.GetContext(), params)

		fmt.Printf("balanceA0: %s \n", s.chainA.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainA.GetContext(), sender))
		fmt.Printf("balanceB0: %s \n", s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver))

		// Set sender
		sender = s.chainA.SenderAccount.GetAddress()

		// Create receiver with secp256k1 address
		receiver = s.chainB.SenderAccount.GetAddress()
		// priv := secp256k1.GenPrivKey()
		// receiverHash := common.BytesToAddress(priv.PubKey().Address().Bytes())
		// receiver = sdk.AccAddress(receiverHash.Bytes())
		// baseAcc := authtypes.NewBaseAccountWithAddress(receiver)
		// s.chainB.App.(*app.Evmos).AccountKeeper.SetAccount(s.chainB.GetContext(), baseAcc)

		// Send coins from chainA to chainB over IBC
		sendCoinfromAtoBWithIBC(sender, receiver, coin)

		fmt.Printf("balanceA1: %s \n", s.chainA.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainA.GetContext(), sender))
		fmt.Printf("balanceB1: %s \n", s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver))

		// Activate IBC callback
		params.EnableWithdraw = true
		params.EnabledChannels = []string{"channel-0"}
		s.chainB.App.(*app.Evmos).WithdrawKeeper.SetParams(s.chainB.GetContext(), params)
	})

	Context("to a secp256k1 receiver address with balance", func() {
		It("transfers all receiver balances to the respective chains", func() {
			// send coin from chainA to chainB
			sendCoinfromAtoBWithIBC(sender, receiver, coin)

			// receiver balance is 0
			balances := s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver)
			fmt.Printf("balanceA2: %s \n", s.chainA.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainA.GetContext(), sender))
			fmt.Printf("balanceB2: %s \n", balances)
			Expect(balances.IsZero()).To(BeTrue())

			// sender balance is original receiver balance

		})
	})
})

func sendCoinfromAtoBWithIBC(from, to sdk.AccAddress, coin sdk.Coin) {
	path := s.path

	// send coin from chainA to chainB
	transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coin, from.String(), to.String(), timeoutHeight, 0)
	_, err := s.chainA.SendMsgs(transferMsg)
	s.Require().NoError(err) // message committed

	// receive coin on chainB from chainA
	fungibleTokenPacket := transfertypes.NewFungibleTokenPacketData("aevmos", "100", from.String(), to.String())
	packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// get proof of packet commitment from chainA
	err = path.EndpointB.UpdateClient()
	s.Require().NoError(err)
	packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := path.EndpointA.QueryProof(packetKey)

	recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, to.String())
	_, err = s.chainB.SendMsgs(recvMsg)
	s.Require().NoError(err) // message committed
}
