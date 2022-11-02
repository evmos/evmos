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
		// timeout                uint64
		// claim                  claimtypes.ClaimsRecord
	)

	// Register ATOM with a Token Pair for testing
	validMetadata := banktypes.Metadata{
		Description: "IBC Coin for IBC Cosmos Chain",
		Base:        uosmoIbcdenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    uatomDenomtrace.BaseDenom,
				Exponent: 0,
			},
		},
		Name:    uosmoIbcdenom,
		Symbol:  erc20Symbol,
		Display: uatomDenomtrace.BaseDenom,
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
			// senderAcc = sdk.MustAccAddressFromBech32(sender)
			// receiverAcc = sdk.MustAccAddressFromBech32(receiver)
		})
		It("should transfer and not convert to erc20", func() {
			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", 10, sender, receiver, 1)

			// nativeEvmos := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), senderAcc, "aevmos")
			// Expect(nativeEvmos).To(Equal(coinEvmos))
			// ibcOsmo := s.EvmosChain.App.(*app.Evmos).BankKeeper.GetBalance(s.EvmosChain.GetContext(), receiverAcc, uosmoIbcdenom)
			// Expect(ibcOsmo).To(Equal(sdk.NewCoin(uosmoIbcdenom, coinOsmo.Amount)))
		})
	})
	Describe("enabled params and registered uatom", func() {
		BeforeEach(func() {
			erc20params := types.DefaultParams()
			erc20params.EnableErc20 = true
			s.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(s.EvmosChain.GetContext(), erc20params)

			sender = s.IBCOsmosisChain.SenderAccount.GetAddress().String()
			receiver = s.EvmosChain.SenderAccount.GetAddress().String()
			// senderAcc = sdk.MustAccAddressFromBech32(sender)
			receiverAcc = sdk.MustAccAddressFromBech32(receiver)

		})
		It("should transfer and convert uosmo to tokens", func() {
			pair, err := s.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(s.EvmosChain.GetContext(), validMetadata)
			s.Require().NoError(err)

			s.EvmosChain.Coordinator.CommitBlock()

			s.SendAndReceiveMessage(s.pathOsmosisEvmos, s.IBCOsmosisChain, "uosmo", 10, sender, receiver, 1)
			// Check ERC20 balances
			balanceTokenAfter :=
				s.EvmosChain.App.(*app.Evmos).Erc20Keeper.BalanceOf(s.EvmosChain.GetContext(), contracts.ERC20MinterBurnerDecimalsContract.ABI, pair.GetERC20Contract(), common.BytesToAddress(receiverAcc.Bytes()))
			s.Require().Equal(balanceTokenAfter.Int64(), int64(10))

			//TODO: check balances after and before of uosmo coins on receiver to verify that it doesnt have both representations.
		})
		It("should transfer and not convert unregistered coin", func() {
		})
		It("should transfer and not convert aevmos", func() {
		})
		It("should transfer and convert original erc20", func() {
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
