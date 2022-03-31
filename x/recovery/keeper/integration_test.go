package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tharsis/evmos/v3/app"
	"github.com/tharsis/evmos/v3/testutil"
	claimtypes "github.com/tharsis/evmos/v3/x/claims/types"
	"github.com/tharsis/evmos/v3/x/recovery/types"
)

// Tokens got stuck at v1.1.2 through the following options:
//  - Sender on Osmosis/Cosmos without claims record sent IBC transfer to Evmos secp256k1 address
//    => tokens from transfer got stuck
//    => Recovery: Send IBC transfer from same-account address without a claims record
//  - Sender on Osmosis/Cosmos with claims record sent IBC transfer to Evmos secp256k1 address
//    => tokens from transfer got stuck
//    => claims record migrated to Evmos 256k1 account and sender record was deleted
//    => Recovery: Chain is restarted with restored Claims records
var _ = Describe("Recovery: Performing an IBC Transfer", Ordered, func() {
	coinEvmos := sdk.NewCoin("aevmos", sdk.NewInt(10000))
	coinOsmo := sdk.NewCoin("uosmo", sdk.NewInt(10))
	// coinAtom := sdk.NewCoin("uatom", sdk.NewInt(10))

	var (
		sender, receiver       string
		senderAcc, receiverAcc sdk.AccAddress
		timeout                uint64
	)

	BeforeEach(func() {
		s.SetupTest()
		// timeout = uint64(s.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())
	})

	Describe("from a non-authorized chain", func() {
		It("should not recover any tokens", func() {
			// expect sender balance the same
		})
	})
	Describe("from a non-authorized chain", func() {
		It("should not recover any tokens", func() {
			// expect sender balance the same
		})
	})

	Describe("from an authorized, non-EVM chain (e.g. Osmosis)", func() {

		Describe("to a different account on Evmos (sender != recipient)", func() {
			BeforeEach(func() {
				sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
				receiver = s.EvmosChain.SenderAccount.GetAddress().String()
				senderAcc, _ = sdk.AccAddressFromBech32(sender)
				receiverAcc, _ = sdk.AccAddressFromBech32(receiver)
			})

			It("should transfer and not recover tokens", func() {
				s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", 10, sender, receiver, 1)

				nativeEvmos := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "aevmos")
				Expect(nativeEvmos).To(Equal(coinEvmos))
				ibcOsmo := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
				Expect(ibcOsmo).To(Equal(sdk.NewCoin(uosmoIbcdenom, coinOsmo.Amount)))
			})
		})

		Describe("to the sender's own eth_secp256k1 account on Evmos (sender == recipient)", func() {
			BeforeEach(func() {
				sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
				receiver = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
				senderAcc, _ = sdk.AccAddressFromBech32(sender)
				receiverAcc, _ = sdk.AccAddressFromBech32(receiver)
			})

			Context("with disabled recovery parameter", func() {
				BeforeEach(func() {

					params := types.DefaultParams()
					params.EnableRecovery = false
					s.EvmosChain.App.(*app.Evmos).RecoveryKeeper.SetParams(s.EvmosChain.GetContext(), params)
				})
				It("should not transfer or recover tokens", func() {
					s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, coinOsmo.Denom, coinOsmo.Amount.Int64(), sender, receiver, 1)

					nativeEvmos := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "aevmos")
					Expect(nativeEvmos).To(Equal(coinEvmos))
					ibcOsmo := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
					Expect(ibcOsmo).To(Equal(sdk.NewCoin(uosmoIbcdenom, coinOsmo.Amount)))

				})
			})

			Context("with a sender's claims record", func() {
				Context("without completed actions", func() {
					BeforeEach(func() {
						amt := sdk.NewInt(int64(100))
						claim := claimtypes.NewClaimsRecord(amt)
						s.EvmosChain.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(s.EvmosChain.GetContext(), senderAcc, claim)
					})

					It("should not transfer or recover tokens", func() {
						// Prevent further funds from getting stuck
						s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, coinOsmo.Denom, coinOsmo.Amount.Int64(), sender, receiver, 1)

						nativeEvmos := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "aevmos")
						Expect(nativeEvmos).To(Equal(coinEvmos))
						ibcOsmo := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
						Expect(ibcOsmo.IsZero()).To(BeTrue())
					})
				})

				Context("with completed actions", func() {
					// Already has stuck funds
					BeforeEach(func() {
						amt := sdk.NewInt(int64(100))
						coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(int64(75))))
						claim := claimtypes.NewClaimsRecord(amt)
						claim.MarkClaimed(claimtypes.ActionIBCTransfer)
						s.EvmosChain.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(s.EvmosChain.GetContext(), senderAcc, claim)

						// update the escrowed account balance to maintain the invariant
						err := testutil.FundModuleAccount(s.EvmosChain.App.(*app.Evmos).BankKeeper, s.EvmosChain.GetContext(), claimtypes.ModuleName, coins)
						s.Require().NoError(err)
					})

					It("should transfer tokens to the recipient and perform recovery", func() {
						// aevmos & ibc tokens that originated from the sender's chain
						s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, coinOsmo.Denom, coinOsmo.Amount.Int64(), sender, receiver, 1)
						timeout = uint64(s.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

						// Escrow before relaying packets
						balanceEscrow := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), transfertypes.GetEscrowAddress("transfer", "channel-0"), "aevmos")
						Expect(balanceEscrow).To(Equal(coinEvmos))
						ibcOsmo := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
						Expect(ibcOsmo.IsZero()).To(BeTrue())

						// Relay both packets that were sent in the ibc_callback
						err := s.pathOsmosisEvmos.RelayPacket(CreatePacket("10000", "aevmos", sender, receiver, "transfer", "channel-0", "transfer", "channel-0", 1, timeout))
						s.Require().NoError(err)
						err = s.pathOsmosisEvmos.RelayPacket(CreatePacket("10", "transfer/channel-0/uosmo", sender, receiver, "transfer", "channel-0", "transfer", "channel-0", 2, timeout))
						s.Require().NoError(err)

						// Check that the aevmos were recovered
						nativeEvmos := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "aevmos")
						Expect(nativeEvmos.IsZero()).To(BeTrue())
						ibcEvmos := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, aevmosIbcdenom)
						Expect(ibcEvmos).To(Equal(sdk.NewCoin(aevmosIbcdenom, coinEvmos.Amount)))

						// Check that the uosmo were recovered
						ibcOsmo = s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
						Expect(ibcOsmo.IsZero()).To(BeTrue())
						nativeOsmo := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
						Expect(nativeOsmo).To(Equal(coinOsmo))
					})

					It("should not claim/migrate/merge claims records", func() {
						// TODO prevent further funds to get stuck
						//
					})
				})
			})

			Context("without a sender's claims record", func() {
				It("should transfer tokens to the recipient and perform recovery", func() {
					// aevmos & ibc tokens that originated from the sender's chain
					s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, coinOsmo.Denom, coinOsmo.Amount.Int64(), sender, receiver, 1)
					timeout = uint64(s.EvmosChain.GetContext().BlockTime().Add(time.Hour * 4).Add(time.Second * -20).UnixNano())

					// Escrow before relaying packets
					balanceEscrow := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), transfertypes.GetEscrowAddress("transfer", "channel-0"), "aevmos")
					Expect(balanceEscrow).To(Equal(coinEvmos))
					ibcOsmo := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
					Expect(ibcOsmo.IsZero()).To(BeTrue())

					// Relay both packets that were sent in the ibc_callback
					err := s.pathOsmosisEvmos.RelayPacket(CreatePacket("10000", "aevmos", sender, receiver, "transfer", "channel-0", "transfer", "channel-0", 1, timeout))
					s.Require().NoError(err)
					err = s.pathOsmosisEvmos.RelayPacket(CreatePacket("10", "transfer/channel-0/uosmo", sender, receiver, "transfer", "channel-0", "transfer", "channel-0", 2, timeout))
					s.Require().NoError(err)

					// Check that the aevmos were recovered
					nativeEvmos := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "aevmos")
					Expect(nativeEvmos.IsZero()).To(BeTrue())
					ibcEvmos := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, aevmosIbcdenom)
					Expect(ibcEvmos).To(Equal(sdk.NewCoin(aevmosIbcdenom, coinEvmos.Amount)))

					// Check that the uosmo were recovered
					ibcOsmo = s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
					Expect(ibcOsmo.IsZero()).To(BeTrue())
					nativeOsmo := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
					Expect(nativeOsmo).To(Equal(coinOsmo))
				})
				It("should not recover tokens that originated from other chains", func() {
					// TODO non-native and ibc
				})
				It("should not claim/migrate/merge claims records", func() {
					// TODO prevent further funds to get stuck
				})
			})
		})
	})
})
