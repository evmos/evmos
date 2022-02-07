package keeper_test

import (
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	types "github.com/tharsis/evmos/x/incentives/types"
)

var _ = Describe("Integration", Ordered, func() {
	var (
		balanceBefore sdk.Coin
		contractAddr  common.Address
	)
	moduleAcc := s.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	participantAcc := sdk.AccAddress(s.address.Bytes())

	BeforeEach(func() {
		s.SetupTest()

		// Deploy contract
		contractAddr = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
		fmt.Printf("\n contractAddr: %v", contractAddr)
		fmt.Printf("\n s.address: %v", s.address)

		// Create incentive
		in, err := s.app.IncentivesKeeper.RegisterIncentive(
			s.ctx,
			contractAddr,
			mintAllocations,
			epochs,
		)
		s.Require().NoError(err)
		fmt.Printf("\n in: %v", in)

		// Fund participant account with contract tokens
		amount := big.NewInt(100)
		s.MintERC20Token(contractAddr, s.address, s.address, amount)

		// Interact with the contract
		balanceBefore = s.app.BankKeeper.GetBalance(s.ctx, participantAcc, denomMint)
		s.TransferERC20Token(contractAddr, s.address, participant2, amount)

		// Check if module account has zero balance
		moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, denomMint)
		Expect(moduleBalance.IsZero()).To(BeTrue())
	})

	// Epoch mechanism for triggering allocation and distribution
	Describe("Commiting a block", func() {

		// Context("before a weekly epoch ends", func() {
		// 	BeforeEach(func() {
		// 		s.CommitAfter(time.Minute)        // Start Epoch
		// 		s.CommitAfter(time.Hour * 24 * 6) // End Epoch
		// 	})
		// 	It("should allocate mint tokens to the usage incentives module", func() {
		// 		balance := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, denomMint)
		// 		Expect(balance.IsZero()).ToNot(BeTrue())
		// 		// fmt.Print("\nmodule balance before %w", balance)
		// 	})
		// 	It("should not reset the participants gas meter", func() {
		// 		gm, _ := s.app.IncentivesKeeper.GetGasMeter(s.ctx, contractAddr, s.address)
		// 		Expect(gm).ToNot(BeZero())
		// 	})
		// 	It("should not distribute usage incentives to the participant", func() {
		// 		actual := s.app.BankKeeper.GetBalance(s.ctx, participantAcc, denomMint)
		// 		Expect(actual).To(Equal(balanceBefore))
		// 		// fmt.Print("\nparticipant before: %w", actual)
		// 	})
		// })

		Context("after a weekly epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)        // Start Epoch
				s.CommitAfter(time.Hour * 24 * 7) // End Epoch
			})
			// It("should allocate some mint tokens from the usage incentives module", func() {
			// 	balance := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, denomMint)
			// 	Expect(balance.IsZero()).ToNot(BeTrue())
			// 	fmt.Print("\nmodule balance after %w", balance)
			// })
			// It("should reset the participants gas meter", func() {
			// 	gm, _ := s.app.IncentivesKeeper.GetGasMeter(s.ctx, contractAddr, s.address)
			// 	Expect(gm).To(BeZero())
			// })
			It("should distribute usage incentives to the participant", func() {
				regIn, _ := s.app.IncentivesKeeper.GetIncentive(s.ctx, contractAddr)
				fmt.Printf("\nincentive after: %v", regIn)

				actual := s.app.BankKeeper.GetBalance(s.ctx, participantAcc, denomMint)
				Expect(actual).ToNot(Equal(balanceBefore))
				// fmt.Print("\nparticipant after: %w", actual)
			})
		})
	})
})
