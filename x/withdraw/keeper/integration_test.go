package keeper_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

// The IBC Callback transfers all non-evmos balances from the receiver to the
// sender address for a receiver with a secp256k1 key
var _ = Describe("Performing a IBC transfer with enabled callback ", Ordered, func() {

	coin := sdk.NewCoin("testcoin", sdk.NewInt(10))

	var (
		sender   string
		receiver string
	)

	BeforeEach(func() {
		s.SetupTest()

		sender = s.IBCChain.SenderAccount.GetAddress().String()
		receiver = s.IBCChain.SenderAccount.GetAddress().String()

		fmt.Printf("balanceA1: %s \n", s.IBCChain.GetSimApp().BankKeeper.GetAllBalances(s.IBCChain.GetContext(), s.IBCChain.SenderAccount.GetAddress()))
		fmt.Printf("balanceB1: %s \n", s.EvmosChain.App.(*app.Evmos).BankKeeper.GetAllBalances(s.EvmosChain.GetContext(), s.IBCChain.SenderAccount.GetAddress()))

		// Activate IBC callback
		params := types.DefaultParams()
		params.EnableWithdraw = true
		s.EvmosChain.App.(*app.Evmos).WithdrawKeeper.SetParams(s.EvmosChain.GetContext(), params)
	})

	Context("to a secp256k1 receiver address with balance", func() {
		It("transfers all receiver balances to the respective chains", func() {
			// send coin from IBCChain to EvmosChain
			sendCoinfromAtoBWithIBC(sender, receiver, coin, 1)

			// sender balance is original receiver balance
			balancesSender := s.IBCChain.GetSimApp().BankKeeper.GetAllBalances(s.IBCChain.GetContext(), s.IBCChain.SenderAccount.GetAddress())
			fmt.Printf("balanceA2: %s \n", balancesSender)
			// Expect(balancesSender.IsZero()).To(BeTrue())

			// receiver balance is 0
			balancesReceiver := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetAllBalances(s.EvmosChain.GetContext(), s.EvmosChain.SenderAccount.GetAddress())
			fmt.Printf("balanceB2: %s \n", balancesReceiver)
			Expect(balancesReceiver.IsZero()).To(BeTrue())
		})
	})
})

func sendCoinfromAtoBWithIBC(sender, receiver string, coin sdk.Coin, seq uint64) {
	path := s.path

	// send coin from IBCChain to EvmosChain
	transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coin, sender, receiver, timeoutHeight, 0)
	_, err := s.IBCChain.SendMsgs(transferMsg)
	s.Require().NoError(err) // message committed

	// receive coin on EvmosChain from IBCChain
	fungibleTokenPacket := transfertypes.NewFungibleTokenPacketData(coin.Denom, coin.Amount.String(), sender, receiver)
	packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), seq, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// get proof of packet commitment from IBCChain
	err = path.EndpointB.UpdateClient()
	s.Require().NoError(err)
	packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := path.EndpointA.QueryProof(packetKey)

	recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, receiver)
	_, err = s.EvmosChain.SendMsgsWithAccount(s.EvmosChain.SenderPrivKey, s.IBCChain.SenderAccount, recvMsg)
	s.Require().NoError(err) // message committed
}
