package keeper_test

import (
	"math/big"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/common"

	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v6/x/incentives/types"
)

var _ = Describe("Performing EVM transactions", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()

		params := s.app.Erc20Keeper.GetParams(s.ctx)
		params.EnableEVMHook = false
		s.app.Erc20Keeper.SetParams(s.ctx, params)
	})

	// Epoch mechanism for triggering allocation and distribution
	Context("with the ERC20 module disabled", func() {
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the ERC20 module enabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableEVMHook = true
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("Distribution", Ordered, func() {
	var (
		balanceBefore  sdk.Coin
		contractAddr   common.Address
		participantAcc sdk.AccAddress
		moduleAcc      sdk.AccAddress
	)

	BeforeEach(func() {
		s.SetupTest()

		// Enable Inflation
		params := s.app.InflationKeeper.GetParams(s.ctx)
		params.EnableInflation = true
		s.app.InflationKeeper.SetParams(s.ctx, params)

		// set a EOA account for the address
		eoa := &ethermint.EthAccount{
			BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(s.address.Bytes()), nil, 0, 0),
			CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).String(),
		}
		s.app.AccountKeeper.RemoveAccount(s.ctx, eoa)
		s.app.AccountKeeper.SetAccount(s.ctx, eoa)

		acc := s.app.AccountKeeper.GetAccount(s.ctx, s.address.Bytes())
		s.Require().NotNil(acc)

		ethAccount, ok := acc.(ethermint.EthAccountI)
		s.Require().True(ok)
		s.Require().Equal(ethermint.AccountTypeEOA, ethAccount.Type())

		contractAddr = contract
		moduleAcc = s.app.AccountKeeper.GetModuleAddress(types.ModuleName)
		participantAcc = acc.GetAddress()
		// Create incentive
		_, err := s.app.IncentivesKeeper.RegisterIncentive(
			s.ctx,
			contractAddr,
			mintAllocations,
			epochs,
		)
		s.Require().NoError(err)

		// Interact with contract and fund participant account
		amount := big.NewInt(100)
		s.MintERC20Token(contractAddr, s.address, s.address, amount)

		// Check if participant account has zero balance
		balanceBefore = s.app.BankKeeper.GetBalance(s.ctx, participantAcc, denomMint)
		s.Require().True(balanceBefore.IsZero())

		// Check if module account has zero balance
		moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, denomMint)
		s.Require().True(moduleBalance.IsZero())
	})

	// Epoch mechanism for triggering allocation and distribution
	Describe("Commiting a block", func() {
		Context("before a weekly epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)                // Start Epoch
				s.CommitAfter(time.Hour*7*24 - time.Hour) // Before End Epoch
			})
			It("should allocate mint tokens to the usage incentives module", func() {
				balance := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, denomMint)
				Expect(balance.IsZero()).ToNot(BeTrue())
			})
			It("should not reset the participants gas meter", func() {
				gm, _ := s.app.IncentivesKeeper.GetGasMeter(s.ctx, contractAddr, s.address)
				Expect(gm).ToNot(BeZero())
			})
			It("should not distribute usage incentives to the participant", func() {
				actual := s.app.BankKeeper.GetBalance(s.ctx, participantAcc, denomMint)
				Expect(actual).To(Equal(balanceBefore))
			})
		})

		Context("after a weekly epoch ends", func() {
			BeforeEach(func() {
				s.CommitAfter(time.Minute)        // Start Epoch
				s.CommitAfter(time.Hour * 24 * 7) // End Epoch
			})
			It("should allocate some mint tokens from the usage incentives module", func() {
				balance := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, denomMint)
				Expect(balance.IsZero()).ToNot(BeTrue())
			})
			It("should reset the participant gas meter", func() {
				gm, _ := s.app.IncentivesKeeper.GetGasMeter(s.ctx, contractAddr, s.address)
				Expect(gm).To(BeZero())
			})
			It("should distribute usage incentives to the participant", func() {
				actual := s.app.BankKeeper.GetBalance(s.ctx, participantAcc, denomMint)
				Expect(actual).ToNot(Equal(balanceBefore))
			})
		})
	})
})
