package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v9/app"
	"github.com/evmos/evmos/v9/contracts"
	"github.com/evmos/evmos/v9/x/erc20/types"
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Recovery: Performing an IBC Transfer", Ordered, func() {
	// coinEvmos := sdk.NewCoin("aevmos", sdk.NewInt(10000))
	// coinOsmo := sdk.NewCoin("uosmo", sdk.NewInt(10))
	// coinAtom := sdk.NewCoin("uatom", sdk.NewInt(10))

	var (
		sender, receiver string
		receiverAcc      sdk.AccAddress
		senderAcc        sdk.AccAddress
		amount           int64 = 10
		pair             *types.TokenPair
		// timeout                uint64
		// claim                  claimtypes.ClaimsRecord
	)

	// Metadata to register OSMO with a Token Pair for testing
	osmoMeta := banktypes.Metadata{
		Description: "IBC Coin for IBC Osmosis Chain",
		Base:        uosmoIbcdenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    uosmoDenomtrace.BaseDenom,
				Exponent: 0,
			},
		},
		Name:    uosmoIbcdenom,
		Symbol:  erc20Symbol,
		Display: uosmoDenomtrace.BaseDenom,
	}

	BeforeEach(func() {
		s.suiteIBCTesting = true
		s.SetupTest()
	})

	Describe("disabled params", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = false
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			receiver = s.EvmosChain.SenderAccount.GetAddress().String()
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
		})
		It("should transfer and not convert to erc20", func() {
			// register the pair to check that it was not converted to ERC-20
			pair, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)

			// check balance before transfer is 0
			ibcOsmoBalanceBefore := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcOsmoBalanceBefore.Amount.Int64())

			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// check balance after transfer
			ibcOsmoBalanceAfter := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(amount, ibcOsmoBalanceAfter.Amount.Int64())

			// check ERC20 balance - should be zero (no conversion)
			balanceERC20TokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceERC20TokenAfter.Int64())
		})
	})
	Describe("enabled params and registered uosmo", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			receiver = s.EvmosChain.SenderAccount.GetAddress().String()
			senderAcc = sdk.MustAccAddressFromBech32(sender)
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)

			// Register uosmo pair
			var err error
			pair, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)
		})
		It("should transfer and convert uosmo to tokens", func() {
			// Check receiver's balance for IBC and ERC-20 before transfer. Should be zero
			balanceTokenBefore :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceTokenBefore.Int64())

			ibcOsmoBalanceBefore := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcOsmoBalanceBefore.Amount.Int64())

			s.EvmosChain.Coordinator.CommitBlock()
			// Send coins
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// Check ERC20 balances
			balanceTokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(amount, balanceTokenAfter.Int64())

			// Check IBC uosmo coin balance - should be zero
			ibcOsmoBalanceAfter := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcOsmoBalanceAfter.Amount.Int64())
		})
		It("should transfer and not convert unregistered coin (uatom)", func() {
			sender = s.IBCCosmosChain.SenderAccount.GetAddress().String()

			// check balance before transfer is 0
			ibcAtomBalanceBefore := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uatomIbcdenom)
			s.Require().Equal(int64(0), ibcAtomBalanceBefore.Amount.Int64())

			s.EvmosChain.Coordinator.CommitBlock()
			s.SendAndReceiveMessage(s.pathCosmosEvmos, s.IBCCosmosChain, "uatom", amount, sender, receiver, 1)

			// check balance after transfer
			ibcAtomBalanceAfter := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uatomIbcdenom)
			s.Require().Equal(amount, ibcAtomBalanceAfter.Amount.Int64())
		})
		It("should transfer and not convert aevmos", func() {
			// check balance after transfer
			aevmosInitialBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, "aevmos")
			s.Require().Greater(aevmosInitialBalance.Amount.Int64(), (int64(0)))

			s.EvmosChain.Coordinator.CommitBlock()

			// 1. Send aevmos from Evmos to Osmosis
			// send aevmos to Osmosis
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.EvmosChain, "aevmos", amount, receiver, sender, 1)

			// check balance after transfer to Osmosis Chain
			aevmosBalanceAfter := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, "aevmos")
			s.Require().Equal(aevmosInitialBalance.Amount.Int64()-amount, aevmosBalanceAfter.Amount.Int64())

			// check ibc aevmos coins balance on Osmosis
			aevmosIBCBalanceAfter := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, aevmosIbcdenom)
			s.Require().Equal(amount, aevmosIBCBalanceAfter.Amount.Int64())

			// 2. Send aevmos IBC coins from Osmosis to Evmos
			// receive aevmos from osmosis
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, aevmosIbcdenom, amount, sender, receiver, 1)

			// check ibc aevmos coins balance on Osmosis - should be zero
			aevmosIBCSenderFinalBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, aevmosIbcdenom)
			s.Require().Equal(int64(0), aevmosIBCSenderFinalBalance.Amount.Int64())

			// TODO: where did the coins go?? BondDenom is stake.. maybe that's the issue
			// check balance after transfer
			aevmosFinalBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, "aevmos")
			s.Require().Equal(aevmosInitialBalance.Amount.Int64()-amount, aevmosFinalBalance.Amount.Int64())

			// aevmosIBCReceiverFinalBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, aevmosIbcdenom)
			// s.Require().Equal(amount, aevmosIBCReceiverFinalBalance.Amount.Int64())

		})
		It("should transfer and convert original erc20", func() {
			// TODO: figure out how override the IBC Transfer method with the custom one
			// uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "uosmo")

			// // Send coins
			// s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// // transfer back the erc20 to original coin
			// uosmoERC20 := pair.GetERC20Contract().String()
			// s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.EvmosChain, uosmoERC20, amount, receiver, sender, 1)
			
			// balanceTokenAfter :=
			// 	s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			// s.Require().Equal(int64(0), balanceTokenAfter.Int64())

			// uosmoFinalBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "uosmo")
			// s.Require().Equal(uosmoInitialBalance.Amount.Int64(), uosmoFinalBalance.Amount.Int64())

		})
	})
	Describe("Performing recovery with registered coin", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			receiver = s.EvmosChain.SenderAccount.GetAddress().String()
			// senderAcc = sdk.MustAccAddressFromBech32(sender)
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)

		})
		It("should recover and not convert uosmo to tokens", func() {

		})
	})

})
