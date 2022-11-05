package keeper_test

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v9/app"
	"github.com/evmos/evmos/v9/contracts"
	"github.com/evmos/evmos/v9/testutil"
	claimstypes "github.com/evmos/evmos/v9/x/claims/types"
	"github.com/evmos/evmos/v9/x/erc20/types"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Recovery: Performing an IBC Transfer", Ordered, func() {

	var (
		sender, receiver string
		receiverAcc      sdk.AccAddress
		senderAcc        sdk.AccAddress
		amount           int64 = 10
		pair             *types.TokenPair
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

	evmosMeta := banktypes.Metadata{
		Description: "Base Denom for Evmos Chain",
		Base:        claimstypes.DefaultClaimsDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    aevmosDenomtrace.BaseDenom,
				Exponent: 0,
			},
		},
		Name:    claimstypes.DefaultClaimsDenom,
		Symbol:  erc20Symbol,
		Display: aevmosDenomtrace.BaseDenom,
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
			// Register 'aevmos' coin in ERC-20 keeper to validate it is not converting the coins when receiving 'aevmos' thru IBC
			pair, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), evmosMeta)
			s.Require().NoError(err)

			aevmosInitialBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, claimstypes.DefaultClaimsDenom)

			// 1. Send aevmos from Evmos to Osmosis
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.EvmosChain, claimstypes.DefaultClaimsDenom, amount, receiver, sender, 1)

			aevmosAfterBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, claimstypes.DefaultClaimsDenom)
			s.Require().Equal(aevmosInitialBalance.Amount.Int64()-amount, aevmosAfterBalance.Amount.Int64())

			// check ibc aevmos coins balance on Osmosis
			aevmosIBCBalanceBefore := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, aevmosIbcdenom)
			s.Require().Equal(amount, aevmosIBCBalanceBefore.Amount.Int64())

			// 2. Send aevmos IBC coins from Osmosis to Evmos
			ibcCoinMeta := fmt.Sprintf("%s/%s", aevmosDenomtrace.Path, aevmosDenomtrace.BaseDenom)
			s.SendBackCoins(s.pathOsmosisEvmos, s.IBCOsmosisChain, aevmosIbcdenom, amount, sender, receiver, 1, ibcCoinMeta)

			// check ibc aevmos coins balance on Osmosis - should be zero
			aevmosIBCSenderFinalBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, aevmosIbcdenom)
			s.Require().Equal(int64(0), aevmosIBCSenderFinalBalance.Amount.Int64())

			// check aevmos balance after transfer - should be equal to initial balance
			aevmosFinalBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, claimstypes.DefaultClaimsDenom)
			s.Require().Equal(aevmosInitialBalance.Amount.Int64(), aevmosFinalBalance.Amount.Int64())

			// check IBC Coin balance - should be zero
			ibcCoinsBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, aevmosIbcdenom)
			s.Require().Equal(int64(0), ibcCoinsBalance.Amount.Int64())

			// Check ERC20 balances - should be zero
			balanceTokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceTokenAfter.Int64())
		})
		It("should transfer and convert original erc20", func() {
			uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")

			// 1. Send 'uosmo' from Osmosis to Evmos
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

            // validate 'uosmo' was transfered successfully and converted to ERC20
            balanceERC20Token :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(amount, balanceERC20Token.Int64())

			// 2. Transfer back the erc20 from Evmos to Osmosis
			ibcCoinMeta := fmt.Sprintf("%s/%s", uosmoDenomtrace.Path, uosmoDenomtrace.BaseDenom)
			uosmoERC20 := pair.GetERC20Contract().String()
			s.SendBackCoins(s.pathOsmosisEvmos, s.EvmosChain, uosmoERC20, amount, receiver, sender, 1, ibcCoinMeta)

			// after transfer, ERC-20 token balance should be zero
			balanceTokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceTokenAfter.Int64())

			// check IBC Coin balance - should be zero
			ibcCoinsBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcCoinsBalance.Amount.Int64())

			// Final balance on Osmosis should be equal to initial balance
			uosmoFinalBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")
			s.Require().Equal(uosmoInitialBalance.Amount.Int64(), uosmoFinalBalance.Amount.Int64())
		})
	})
	Describe("Performing recovery with registered coin", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			// receiver address is on Osmosis Chain also,
			// but funds are transfered to this address in Evmos chain
			receiver = sender
			senderAcc = sdk.MustAccAddressFromBech32(sender)
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)

			// Register uosmo pair
			var err error
			pair, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)
		})
		It("should recover and not convert uosmo to tokens", func() {
			uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")

			// Send 'uosmo' to Osmosis address in Evmos Chain (locked funds)
			// sender_addr == receiver_addr
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// recovery should trigger and send back the funds to origin account
			// in the Osmosis Chain

			// ERC-20 balance should be zero
			balanceTokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceTokenAfter.Int64())

			// IBC coin balance should be zero
			ibcCoinsBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcCoinsBalance.Amount.Int64())

			// validate that Osmosis address final balance == initial balance
			uosmoFinalBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")
			s.Require().Equal(uosmoInitialBalance.Amount.Int64(), uosmoFinalBalance.Amount.Int64())
		})
	})

	Describe("Performing claims with registered coin", func() {
		BeforeEach(func() {
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), types.DefaultParams())

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			// receiver address is on Osmosis Chain also,
			// but funds are transfered to this address in Evmos chain
			receiver = s.EvmosChain.SenderAccount.GetAddress().String()
			senderAcc = sdk.MustAccAddressFromBech32(sender)
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)

			// Register uosmo pair
			var err error
			pair, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)

			// Authorize channel-0 for claims (Evmos-Osmosis)
			params := s.EvmosChain.App.(*app.Evmos).ClaimsKeeper.GetParams(s.EvmosChain.GetContext())
			params.AuthorizedChannels = []string{
				"channel-0",
			}
			s.EvmosChain.App.(*app.Evmos).ClaimsKeeper.SetParams(s.EvmosChain.GetContext(), params)
		})
		It("claim - sender ≠ recipient, recipient claims record found, where ibc is last action", func() {
			// Register claims record
			initialClaimAmount := sdk.NewInt(100)
			claimableAmount := sdk.NewInt(25)
			s.EvmosChain.App.(*app.Evmos).ClaimsKeeper.SetClaimsRecord(s.EvmosChain.GetContext(), senderAcc, claimstypes.ClaimsRecord{
				InitialClaimableAmount: initialClaimAmount,
				ActionsCompleted:       []bool{true, true, true, false},
			})

			// escrow coins in module
			coins := sdk.NewCoins(sdk.NewCoin(claimstypes.DefaultClaimsDenom, claimableAmount))
			err := testutil.FundModuleAccount(s.EvmosChain.GetContext(), s.EvmosChain.App.(*app.Evmos).BankKeeper, claimstypes.ModuleName, coins)
			s.Require().NoError(err)

			receiverInitialAevmosBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, claimstypes.DefaultClaimsDenom)

			uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")

			// Send 'uosmo' from Osmosis address with claims to Evmos address
			// send the corresponding amount to trigger the claim
			amount, _ := strconv.ParseInt(claimstypes.IBCTriggerAmt, 10, 64)
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// should trigger claims logic and send aevmos coins from claims to receiver

			// ERC-20 balance should be the transfered amount
			balanceTokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(amount, balanceTokenAfter.Int64())

			// IBC coin balance should be zero
			ibcCoinsBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcCoinsBalance.Amount.Int64())

			// validate that Osmosis address balance is correct
			uosmoFinalBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")
			s.Require().Equal(uosmoInitialBalance.Amount.Int64()-amount, uosmoFinalBalance.Amount.Int64())

			// validate that Receiver address on Evmos got the claims tokens
			receiverFinalAevmosBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, claimstypes.DefaultClaimsDenom)
			s.Require().Equal(receiverInitialAevmosBalance.Amount.Int64()+claimableAmount.Int64(), receiverFinalAevmosBalance.Amount.Int64())
		})
	})

})
