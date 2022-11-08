package keeper_test

import (
	"fmt"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v5/testing/simapp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v10/app"
	"github.com/evmos/evmos/v10/contracts"
	"github.com/evmos/evmos/v10/testutil"
	claimstypes "github.com/evmos/evmos/v10/x/claims/types"
	"github.com/evmos/evmos/v10/x/erc20/types"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Convert receiving IBC to Erc20", Ordered, func() {

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
			balanceERC20TokenAfter := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
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
			balanceTokenBefore := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceTokenBefore.Int64())

			ibcOsmoBalanceBefore := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			s.Require().Equal(int64(0), ibcOsmoBalanceBefore.Amount.Int64())

			s.EvmosChain.Coordinator.CommitBlock()
			// Send coins
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// Check ERC20 balances
			balanceTokenAfter := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
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
			balanceTokenAfter := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceTokenAfter.Int64())
		})
		It("should transfer and convert original erc20", func() {
			uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), senderAcc, "uosmo")

			// 1. Send 'uosmo' from Osmosis to Evmos
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1)

			// validate 'uosmo' was transfered successfully and converted to ERC20
			balanceERC20Token := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(amount, balanceERC20Token.Int64())

			// 2. Transfer back the erc20 from Evmos to Osmosis
			ibcCoinMeta := fmt.Sprintf("%s/%s", uosmoDenomtrace.Path, uosmoDenomtrace.BaseDenom)
			uosmoERC20 := pair.GetERC20Contract().String()
			s.SendBackCoins(s.pathOsmosisEvmos, s.EvmosChain, uosmoERC20, amount, receiver, sender, 1, ibcCoinMeta)

			// after transfer, ERC-20 token balance should be zero
			balanceTokenAfter := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
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
			balanceTokenAfter := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
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
			balanceTokenAfter := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
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

var _ = Describe("Convert outgoing ERC20 to IBC", Ordered, func() {
	var (
		sender, receiver string
		receiverAcc      sdk.AccAddress
		senderAcc        sdk.AccAddress
		amount           int64 = 10
		pair             *types.TokenPair
		erc20Denomtrace  transfertypes.DenomTrace
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
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			receiver = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			sender = s.EvmosChain.SenderAccount.GetAddress().String()
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
			senderAcc = sdk.MustAccAddressFromBech32(sender)

			// Register ERC20 pair
			var err error
			addr, err := s.DeployContractToChain("testcoin", "tt", 18)
			s.Require().NoError(err)
			pair, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterERC20(s.EvmosChain.GetContext(), addr)
			s.Require().NoError(err)
			s.EvmosChain.Coordinator.CommitBlock()
			erc20params.EnableErc20 = false
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)
		})
		It("should fail transfer and not convert to IBC", func() {
			// Mint tokens and send to receiver
			_, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.CallEVM(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, common.BytesToAddress(senderAcc.Bytes()), pair.GetERC20Contract(), true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(amount))
			s.Require().NoError(err)
			// Check Balance
			balanceToken :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())

			path := s.pathOsmosisEvmos
			originEndpoint := path.EndpointB
			destEndpoint := path.EndpointA
			originChain := s.EvmosChain
			coin := pair.Erc20Address
			transfer := transfertypes.NewFungibleTokenPacketData(pair.Denom, strconv.Itoa(int(amount*2)), sender, receiver)
			transferMsg := transfertypes.NewMsgTransfer(originEndpoint.ChannelConfig.PortID, originEndpoint.ChannelID, sdk.NewCoin(coin, sdk.NewInt(amount*2)), sender, receiver, timeoutHeight, 0)

			originChain.Coordinator.UpdateTimeForChain(originChain)

			_, _, err = simapp.SignAndDeliver(
				originChain.T,
				originChain.TxConfig,
				originChain.App.GetBaseApp(),
				originChain.GetContext().BlockHeader(),
				[]sdk.Msg{transferMsg},
				originChain.ChainID,
				[]uint64{originChain.SenderAccount.GetAccountNumber()},
				[]uint64{originChain.SenderAccount.GetSequence()},
				true, false, originChain.SenderPrivKey,
			)
			s.Require().Error(err)
			// NextBlock calls app.Commit()
			originChain.NextBlock()

			// increment sequence for successful transaction execution
			err = originChain.SenderAccount.SetSequence(originChain.SenderAccount.GetSequence() + 1)
			s.Require().NoError(err)
			originChain.Coordinator.IncrementTime()

			packet := channeltypes.NewPacket(transfer.GetBytes(), 1, originEndpoint.ChannelConfig.PortID, originEndpoint.ChannelID, destEndpoint.ChannelConfig.PortID, destEndpoint.ChannelID, timeoutHeight, 0)
			// Receive message on the counterparty side, and send ack
			err = path.RelayPacket(packet)
			s.Require().Error(err)

			// Check Balance didnt change
			balanceToken =
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())

		})
	})
	Describe("registered erc20", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			receiver = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			sender = s.EvmosChain.SenderAccount.GetAddress().String()
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
			senderAcc = sdk.MustAccAddressFromBech32(sender)

			// Register ERC20 pair
			var err error
			addr, err := s.DeployContractToChain("testcoin", "tt", 18)
			s.Require().NoError(err)
			pair, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterERC20(s.EvmosChain.GetContext(), addr)
			s.Require().NoError(err)

			erc20Denomtrace = transfertypes.DenomTrace{
				Path:      "transfer/channel-0",
				BaseDenom: pair.Denom,
			}

		})
		It("should transfer available balance", func() {
			// Mint tokens and send to receiver
			_, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.CallEVM(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, common.BytesToAddress(senderAcc.Bytes()), pair.GetERC20Contract(), true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(amount))
			s.Require().NoError(err)
			// Check Balance
			balanceToken :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())

			msgConvertERC20 := types.NewMsgConvertERC20(
				sdk.NewInt(amount),
				senderAcc,
				pair.GetERC20Contract(),
				common.BytesToAddress(senderAcc.Bytes()),
			)

			err = msgConvertERC20.ValidateBasic()
			s.Require().NoError(err)
			// Use MsgConvertERC20 to convert the ERC20 to a Cosmos IBC Coin
			_, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.ConvertERC20(sdk.WrapSDKContext(s.EvmosChain.GetContext()), msgConvertERC20)
			s.Require().NoError(err)
			// Check Balance
			balanceToken =
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(int64(0), balanceToken.Int64())

			// IBC coin balance should be amount
			erc20CoinsBalance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, pair.Denom)
			s.Require().Equal(amount, erc20CoinsBalance.Amount.Int64())

			s.EvmosChain.Coordinator.CommitBlock()

			// TODO: find what is causing this error.
			s.EvmosChain.SenderAccount.SetSequence(s.EvmosChain.SenderAccount.GetSequence() + 1)

			// Attempt to send
			s.SendBackCoins(s.pathOsmosisEvmos, s.EvmosChain, pair.Erc20Address, amount, sender, receiver, 1, pair.Denom)
			s.IBCOsmosisChain.Coordinator.CommitBlock()

			// Check balance on the Osmosis chain
			erc20IBCBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, erc20Denomtrace.IBCDenom())
			s.Require().Equal(amount, erc20IBCBalance.Amount.Int64())
		})
		It("should convert and transfer if no ibc balance", func() {
			// Mint tokens and send to receiver
			_, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.CallEVM(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, common.BytesToAddress(senderAcc.Bytes()), pair.GetERC20Contract(), true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(amount))
			s.Require().NoError(err)

			// Check Balance
			balanceToken :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())

			s.EvmosChain.SenderAccount.SetSequence(s.EvmosChain.SenderAccount.GetSequence() + 1)
			// Attempt to send
			s.SendBackCoins(s.pathOsmosisEvmos, s.EvmosChain, pair.Erc20Address, amount, sender, receiver, 1, pair.Denom)

			s.EvmosChain.Coordinator.CommitBlock()
			// Check balance of erc20 depleted
			balanceToken =
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(int64(0), balanceToken.Int64())

			// Check balance on the Osmosis chain
			ibcOsmosBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, erc20Denomtrace.IBCDenom())
			s.Require().Equal(amount, ibcOsmosBalance.Amount.Int64())

		})
		It("should fail if balance is not enough", func() {
			// Mint tokens and send to receiver
			_, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.CallEVM(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, common.BytesToAddress(senderAcc.Bytes()), pair.GetERC20Contract(), true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(amount))
			s.Require().NoError(err)

			// Check Balance
			balanceToken :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())

			// Attempt to send
			s.EvmosChain.SenderAccount.SetSequence(s.EvmosChain.SenderAccount.GetSequence() + 1)

			path := s.pathOsmosisEvmos
			originEndpoint := path.EndpointB
			destEndpoint := path.EndpointA
			originChain := s.EvmosChain
			coin := pair.Erc20Address
			transfer := transfertypes.NewFungibleTokenPacketData(pair.Denom, strconv.Itoa(int(amount*2)), sender, receiver)
			transferMsg := transfertypes.NewMsgTransfer(originEndpoint.ChannelConfig.PortID, originEndpoint.ChannelID, sdk.NewCoin(coin, sdk.NewInt(amount*2)), sender, receiver, timeoutHeight, 0)

			originChain.Coordinator.UpdateTimeForChain(originChain)

			_, _, err = simapp.SignAndDeliver(
				originChain.T,
				originChain.TxConfig,
				originChain.App.GetBaseApp(),
				originChain.GetContext().BlockHeader(),
				[]sdk.Msg{transferMsg},
				originChain.ChainID,
				[]uint64{originChain.SenderAccount.GetAccountNumber()},
				[]uint64{originChain.SenderAccount.GetSequence()},
				true, false, originChain.SenderPrivKey,
			)
			s.Require().Error(err)
			// NextBlock calls app.Commit()
			originChain.NextBlock()

			// increment sequence for successful transaction execution
			err = originChain.SenderAccount.SetSequence(originChain.SenderAccount.GetSequence() + 1)
			s.Require().NoError(err)
			originChain.Coordinator.IncrementTime()

			packet := channeltypes.NewPacket(transfer.GetBytes(), 1, originEndpoint.ChannelConfig.PortID, originEndpoint.ChannelID, destEndpoint.ChannelConfig.PortID, destEndpoint.ChannelID, timeoutHeight, 0)
			// Receive message on the counterparty side, and send ack
			err = path.RelayPacket(packet)
			s.Require().Error(err)

			// Check Balance didnt change
			ibcOsmosBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, erc20Denomtrace.IBCDenom())
			s.Require().Equal(int64(0), ibcOsmosBalance.Amount.Int64())
			balanceToken =
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())
		})
	})
	Describe("registered coin", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			receiver = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			sender = s.EvmosChain.SenderAccount.GetAddress().String()
			print(sender)
			print(receiver)
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
			senderAcc = sdk.MustAccAddressFromBech32(sender)

			// Register uosmo pair
			var err error
			pair, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)
		})
		It("should transfer available balance", func() {
			uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")

			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, receiver, sender, 1)
			s.EvmosChain.Coordinator.CommitBlock(s.EvmosChain)

			// FIXME: ibc transfer not beeing received on evmos. quick patch mint on evmos side.
			coins := sdk.NewCoins(sdk.NewCoin(uosmoIbcdenom, sdk.NewInt(amount)))
			err := s.EvmosChain.App.(*app.Evmos).BankKeeper.MintCoins(s.EvmosChain.GetContext(), types.ModuleName, coins)
			s.Require().NoError(err)
			err = s.EvmosChain.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToAccount(s.EvmosChain.GetContext(), types.ModuleName, senderAcc, coins)
			s.Require().NoError(err)

			balance := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, uosmoIbcdenom)
			s.Require().Equal(amount, balance.Amount.Int64())

			msgConvertCoin := types.NewMsgConvertCoin(
				sdk.NewCoin(pair.Denom, sdk.NewInt(amount)),
				common.BytesToAddress(senderAcc.Bytes()),
				senderAcc,
			)

			err = msgConvertCoin.ValidateBasic()
			s.Require().NoError(err)
			// Use MsgConvertERC20 to convert the ERC20 to a Cosmos IBC Coin
			_, err = s.EvmosChain.App.(*app.Evmos).Erc20Keeper.ConvertCoin(sdk.WrapSDKContext(s.EvmosChain.GetContext()), msgConvertCoin)
			s.Require().NoError(err)

			s.EvmosChain.Coordinator.CommitBlock()
			// Attempt to send
			s.SendBackCoins(s.pathOsmosisEvmos, s.EvmosChain, pair.Erc20Address, amount, sender, receiver, 1, uosmoDenomtrace.GetFullDenomPath())
			s.IBCOsmosisChain.Coordinator.CommitBlock()
			// Check balance on the Osmosis chain
			uosmoBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
			s.Require().Equal(uosmoInitialBalance, uosmoBalance)
		})
	})
})
