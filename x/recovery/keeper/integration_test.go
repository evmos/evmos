package keeper_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Tokens got stuck at v1.1.2 through the following options:
//  - Sender on Osmosis/Cosmos without claims record sent IBC transfer to Evmos secp256k1 address
//    => tokens from transfer got stuck
//    => Recover with IBC transfer from same-account address without a claims record
//  - Sender on Osmosis/Cosmos with claims record sent IBC transfer to Evmos secp256k1 address
//    => tokens from transfer got stuck
//    => claims record migrated to Evmos 256k1 account

// Stack
// RecvPacket, message that originates from core IBC and goes down to app, the flow is the otherway
// channel.RecvPacket -> transfer.OnRecvPacket -> claim.OnRecvPacket -> recovery.OnRecvPacket

// claim only with sender != recipient
// recovery only with sender == recipient

var _ = Describe("Recovery: Performing an IBC Transfer", Ordered, func() {
	BeforeAll(func() {
		s.SetupTest()
	})

	Describe("from a non-authorized chain", func() {
		It("should not recover any tokens", func() {
			// expect sender balance the same
		})
	})

	Describe("from an authorized, non-EVM chain (e.g. Osmosis)", func() {

		Describe("to a different account on Evmos (sender != recipient)", func() {
			It("should not recover any tokens", func() {
				// expect sender balance the same
				// expect recipient balance the same
			})
		})

		Describe("to the sender's own eth_secp256k1 account on Evmos (sender == recipient)", func() {
			Context("with a sender's claims record", func() {
				Context("without completed actions", func() {
					It("should not transfer or recover tokens", func() {
						// Prevent further funds from getting stuck
					})
				})

				Context("with completed actions", func() {
					// Already has stuck funds
					It("should transfer tokens to the recipient and perform recovery", func() {
						// aevmos & tokens that originated from the sender's chain
					})
					It("should not claim/migrate/merge claims records", func() {
						// prevent further funds to get stuck
					})
				})
			})

			Context("without a sender's claims record", func() {
				It("should transfer tokens to the recipient and perform recovery", func() {
					// aevmos & tokens that originated from the sender's chain
				})
				It("should not recover tokens that originated from other chains", func() {
					// non-native and ibc
				})
				It("should not claim/migrate/merge claims records", func() {
					// prevent further funds to get stuck
				})
			})
		})
	})
})
