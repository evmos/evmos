package keeper_test

import (
	"math/big"
	"strconv"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/contracts"
	ibctesting "github.com/evmos/evmos/v16/ibc/testing"
	teststypes "github.com/evmos/evmos/v16/types/tests"
	"github.com/evmos/evmos/v16/x/erc20/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Convert receiving IBC to Erc20", Ordered, func() {
	var (
		sender, receiver string
		receiverAcc      sdk.AccAddress
		amount           int64 = 10
	)

	// Metadata to register OSMO with a Token Pair for testing
	osmoMeta := banktypes.Metadata{
		Description: "IBC Coin for IBC Osmosis Chain",
		Base:        teststypes.UosmoIbcdenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    teststypes.UosmoDenomtrace.BaseDenom,
				Exponent: 0,
			},
		},
		Name:    teststypes.UosmoIbcdenom,
		Symbol:  erc20Symbol,
		Display: teststypes.UosmoDenomtrace.BaseDenom,
	}

	BeforeEach(func() {
		s.suiteIBCTesting = true
		s.SetupTest()
		s.suiteIBCTesting = false
	})

	Describe("disabled params", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = false
			err := s.app.Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)
			s.Require().NoError(err)

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			receiver = s.EvmosChain.SenderAccount.GetAddress().String()
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
		})
		It("should transfer and not convert to erc20", func() {
			// register the pair to check that it was not converted to ERC-20
			pair, err := s.app.Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)

			// check balance before transfer is 0
			ibcOsmoBalanceBefore := s.app.BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, teststypes.UosmoIbcdenom)
			s.Require().Equal(int64(0), ibcOsmoBalanceBefore.Amount.Int64())

			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, sender, receiver, 1, "")

			// check balance after transfer
			ibcOsmoBalanceAfter := s.app.BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, teststypes.UosmoIbcdenom)
			s.Require().Equal(amount, ibcOsmoBalanceAfter.Amount.Int64())

			// check ERC20 balance - should be zero (no conversion)
			balanceERC20TokenAfter := s.app.Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(int64(0), balanceERC20TokenAfter.Int64())
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
	)

	// Metadata to register OSMO with a Token Pair for testing
	osmoMeta := banktypes.Metadata{
		Description: "IBC Coin for IBC Osmosis Chain",
		Base:        teststypes.UosmoIbcdenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    teststypes.UosmoDenomtrace.BaseDenom,
				Exponent: 0,
			},
		},
		Name:    teststypes.UosmoIbcdenom,
		Symbol:  erc20Symbol,
		Display: teststypes.UosmoDenomtrace.BaseDenom,
	}

	BeforeEach(func() {
		s.suiteIBCTesting = true
		s.SetupTest()
		s.suiteIBCTesting = false
	})

	Describe("disabled params", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			err := s.app.Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)
			s.Require().NoError(err)

			receiver = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			sender = s.EvmosChain.SenderAccount.GetAddress().String()
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
			senderAcc = sdk.MustAccAddressFromBech32(sender)

			// Register ERC20 pair
			addr, err := s.DeployContractToChain("testcoin", "tt", 18)
			s.Require().NoError(err)
			pair, err = s.app.Erc20Keeper.RegisterERC20(s.EvmosChain.GetContext(), addr)
			s.Require().NoError(err)
			s.EvmosChain.Coordinator.CommitBlock()
			erc20params.EnableErc20 = false
			err = s.app.Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)
			s.Require().NoError(err)
		})
		It("should fail transfer and not convert to IBC", func() {
			// Mint tokens and send to receiver
			_, err := s.app.Erc20Keeper.CallEVM(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, common.BytesToAddress(senderAcc.Bytes()), pair.GetERC20Contract(), true, "mint", common.BytesToAddress(senderAcc.Bytes()), big.NewInt(amount))
			s.Require().NoError(err)
			// Check Balance
			balanceToken := s.app.Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())

			path := s.pathOsmosisEvmos
			originEndpoint := path.EndpointB
			destEndpoint := path.EndpointA
			originChain := s.EvmosChain
			coin := pair.Denom
			transfer := transfertypes.NewFungibleTokenPacketData(pair.Denom, strconv.Itoa(int(amount*2)), sender, receiver, "")
			transferMsg := transfertypes.NewMsgTransfer(originEndpoint.ChannelConfig.PortID, originEndpoint.ChannelID, sdk.NewCoin(coin, math.NewInt(amount*2)), sender, receiver, timeoutHeight, 0, "")

			originChain.Coordinator.UpdateTimeForChain(originChain)
			denom := originChain.App.(*app.Evmos).StakingKeeper.BondDenom(originChain.GetContext())
			fee := sdk.Coins{sdk.NewInt64Coin(denom, ibctesting.DefaultFeeAmt)}

			_, _, err = ibctesting.SignAndDeliver(
				originChain.T,
				originChain.TxConfig,
				originChain.App.GetBaseApp(),
				[]sdk.Msg{transferMsg},
				fee,
				originChain.ChainID,
				[]uint64{originChain.SenderAccount.GetAccountNumber()},
				[]uint64{originChain.SenderAccount.GetSequence()},
				false, originChain.SenderPrivKey,
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
			balanceToken = s.app.Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(senderAcc.Bytes()))
			s.Require().Equal(amount, balanceToken.Int64())
		})
	})
	Describe("registered coin", func() {
		BeforeEach(func() {
			receiver = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			sender = s.EvmosChain.SenderAccount.GetAddress().String()
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)
			senderAcc = sdk.MustAccAddressFromBech32(sender)

			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = false
			err := s.app.Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)
			s.Require().NoError(err)

			// Send from osmosis to Evmos
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", amount, receiver, sender, 1, "")
			s.EvmosChain.Coordinator.CommitBlock(s.EvmosChain)
			erc20params.EnableErc20 = true
			err = s.app.Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)
			s.Require().NoError(err)

			// Register uosmo pair
			pair, err = s.app.Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), osmoMeta)
			s.Require().NoError(err)
		})
		It("should transfer available balance", func() {
			uosmoInitialBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")

			balance := s.app.BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, teststypes.UosmoIbcdenom)
			s.Require().Equal(amount, balance.Amount.Int64())

			// Attempt to send erc20 tokens to osmosis and convert automatically
			s.SendBackCoins(s.pathOsmosisEvmos, s.EvmosChain, pair.Denom, amount, sender, receiver, 1, teststypes.UosmoDenomtrace.GetFullDenomPath())
			s.IBCOsmosisChain.Coordinator.CommitBlock()
			// Check balance on the Osmosis chain
			uosmoBalance := s.IBCOsmosisChain.GetSimApp().BankKeeper.GetBalance(s.IBCOsmosisChain.GetContext(), receiverAcc, "uosmo")
			s.Require().Equal(uosmoInitialBalance.Amount.Int64()+amount, uosmoBalance.Amount.Int64())
		})
	})
})
