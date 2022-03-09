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

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

// The IBC Callback transfers all non-evmos balances from the receiver to the
// sender address for a receiver with a secp256k1 key
var _ = Describe("Performing a IBC transfer with enabled callback ", Ordered, func() {

	coin := sdk.NewCoin("uatom", sdk.NewInt(100))
	coins := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000)))

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

		// Get secp addresses
		pk := secp256k1.GenPrivKey()
		secpAddr := sdk.AccAddress(pk.PubKey().Address())
		baseAcc := authtypes.NewBaseAccountWithAddress(secpAddr)

		secpAddrEvmos := secpAddr.String()
		secpAddrCosmos := sdk.MustBech32ifyAddressBytes(sdk.Bech32MainPrefix, secpAddr)
		fmt.Println(secpAddrEvmos)
		fmt.Println(secpAddrCosmos)

		// Set sender with secp256k1 on Cosmos chain
		s.chainA.SenderPrivKey = pk
		s.chainA.SenderAccount = baseAcc
		sender = s.chainA.SenderAccount.GetAddress()
		s.chainA.GetSimApp().AccountKeeper.SetAccount(s.chainA.GetContext(), baseAcc)

		// TODO Set receiver with secp256k1 on Evmos chain
		// s.chainB.SenderPrivKey = pk
		// s.chainB.SenderAccount = baseAcc
		receiver = s.chainB.SenderAccount.GetAddress()
		// s.chainB.App.(*app.Evmos).AccountKeeper.SetAccount(s.chainB.GetContext(), baseAcc)

		err := s.chainA.GetSimApp().BankKeeper.MintCoins(s.chainA.GetContext(), minttypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.chainA.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), minttypes.ModuleName, s.chainA.SenderAccount.GetAddress(), coins)
		s.Require().NoError(err)

		fmt.Printf("balanceA0: %s \n", s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), sender))
		fmt.Printf("balanceB0: %s \n", s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver))

		// Send coins from chainA to chainB over IBC
		sendCoinfromAtoBWithIBC(sender, receiver, coin)

		fmt.Printf("balanceA1: %s \n", s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), sender))
		fmt.Printf("balanceB1: %s \n", s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver))

		// Activate IBC callback
		params.EnableWithdraw = true
		s.chainB.App.(*app.Evmos).WithdrawKeeper.SetParams(s.chainB.GetContext(), params)
	})

	Context("to a secp256k1 receiver address with balance", func() {
		It("transfers all receiver balances to the respective chains", func() {
			// send coin from chainA to chainB
			sendCoinfromAtoBWithIBC(sender, receiver, coin)

			// sender balance is original receiver balance
			balancesSender := s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), receiver)
			fmt.Printf("balanceA2: %s \n", balancesSender)
			// Expect(balancesSender.IsZero()).To(BeTrue())

			// receiver balance is 0
			balancesReceiver := s.chainB.App.(*app.Evmos).BankKeeper.GetAllBalances(s.chainB.GetContext(), receiver)
			fmt.Printf("balanceB2: %s \n", balancesReceiver)
			Expect(balancesReceiver.IsZero()).To(BeTrue())
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
	fungibleTokenPacket := transfertypes.NewFungibleTokenPacketData(coin.Denom, coin.Amount.String(), from.String(), to.String())
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
