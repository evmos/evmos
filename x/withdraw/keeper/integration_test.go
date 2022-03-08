package keeper_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/ethereum/go-ethereum/common"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/withdraw/types"
)

// The IBC Callback transfers all non-evmos balances from the receiver to the
// sender address for a receiver with a secp256k1 key
var _ = Describe("Performing a IBC transfer with enabled callback ", Ordered, func() {

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

		// Set sender
		sender = s.chainA.SenderAccount.GetAddress()

		// Create receiver with secp256k1 address
		priv := secp256k1.GenPrivKey()
		receiverHash := common.BytesToAddress(priv.PubKey().Address().Bytes())
		receiver = sdk.AccAddress(receiverHash.Bytes())
		baseAcc := authtypes.NewBaseAccountWithAddress(receiver)
		s.chainB.App.(*app.Evmos).AccountKeeper.SetAccount(s.chainB.GetContext(), baseAcc)

		// Send coins from chainA to chainB over IBC
		sendCoinsWithIBC(sender, receiver)

		// Activate IBC callback
		params.EnableWithdraw = true
		s.chainB.App.(*app.Evmos).WithdrawKeeper.SetParams(s.chainB.GetContext(), params)
	})

	Context("to a secp256k1 receiver address with balance", func() {
		It("transfers all receiver balances to the respective chains", func() {
			// send coin from chainA to chainB
			sendCoinsWithIBC(sender, receiver)

			// receiver balance is 0
			balances := s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver)
			fmt.Println(balances)
			Expect(balances.IsZero()).To(BeTrue())

			// sender balance is original receiver balance

		})
	})
})

func sendCoinsWithIBC(from, to sdk.AccAddress) {
	path := s.path

	// send coin from chainA to chainB
	coin := sdk.NewCoin("aevmos", sdk.NewInt(100))
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
