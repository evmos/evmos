package keeper_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

// The IBC Callback transfers all non-evmos balances from the receiver to the
// sender address for a receiver with a secp256k1 key
var _ = Describe("Performing a IBC transfer with enabled callback ", Ordered, func() {

	coin := sdk.NewCoin("uatom", sdk.NewInt(100))
	coins := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000)))

	var (
		secpAddr sdk.AccAddress
		sender   string
		receiver string
	)

	BeforeEach(func() {
		s.SetupTest()

		// Deactivate IBC callback
		params := types.DefaultParams()
		params.EnableWithdraw = false
		s.chainB.App.(*app.Evmos).WithdrawKeeper.SetParams(s.chainB.GetContext(), params)

		// Set sender with secp256k1 on Cosmos chain and generate sender and receiver addresses from it
		// pk := secp256k1.GenPrivKey()
		secpAddr = s.chainA.SenderAccount.GetAddress()
		secpAddrEvmos := secpAddr.String()
		secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)

		sender = secpAddrCosmos
		receiver = secpAddrEvmos

		baseAcc := authtypes.NewBaseAccountWithAddress(secpAddr)
		s.chainB.App.(*app.Evmos).AccountKeeper.SetAccount(s.chainB.GetContext(), baseAcc)

		// Fund chain A aacount
		err := s.chainA.GetSimApp().BankKeeper.MintCoins(s.chainA.GetContext(), minttypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.chainA.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), minttypes.ModuleName, secpAddr, coins)
		s.Require().NoError(err)

		fmt.Printf("balanceA0: %s \n", s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), secpAddr))
		fmt.Printf("balanceB0: %s \n", s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), secpAddr))

		// Send coins from chainA to chainB over IBC
		sendCoinfromAtoBWithIBC(sender, receiver, coin, 1)

		fmt.Printf("balanceA1: %s \n", s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), secpAddr))
		fmt.Printf("balanceB1: %s \n", s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), secpAddr))

		// Activate IBC callback
		params.EnableWithdraw = true
		s.chainB.App.(*app.Evmos).WithdrawKeeper.SetParams(s.chainB.GetContext(), params)
	})

	Context("to a secp256k1 receiver address with balance", func() {
		It("transfers all receiver balances to the respective chains", func() {
			// send coin from chainA to chainB
			sendCoinfromAtoBWithIBC(sender, receiver, coin, 2)

			// sender balance is original receiver balance
			balancesSender := s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), secpAddr)
			fmt.Printf("balanceA2: %s \n", balancesSender)
			// Expect(balancesSender.IsZero()).To(BeTrue())

			// receiver balance is 0
			balancesReceiver := s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), secpAddr)
			fmt.Printf("balanceB2: %s \n", balancesReceiver)
			Expect(balancesReceiver.IsZero()).To(BeTrue())
		})
	})
})

func sendCoinfromAtoBWithIBC(from, to string, coin sdk.Coin, seq uint64) {
	path := s.path

	// send coin from chainA to chainB
	transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coin, from, to, timeoutHeight, 0)
	_, err := s.chainA.SendMsgs(transferMsg)
	s.Require().NoError(err) // message committed

	// receive coin on chainB from chainA
	fungibleTokenPacket := transfertypes.NewFungibleTokenPacketData(coin.Denom, coin.Amount.String(), from, to)
	packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), seq, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)

	// get proof of packet commitment from chainA
	err = path.EndpointB.UpdateClient()
	s.Require().NoError(err)
	packetKey := host.PacketCommitmentKey(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	proof, proofHeight := path.EndpointA.QueryProof(packetKey)

	recvMsg := channeltypes.NewMsgRecvPacket(packet, proof, proofHeight, to)
	_, err = s.chainB.SendMsgs(recvMsg)
	s.Require().NoError(err) // message committed
}
