package ante_test

import (
	sdkmath "cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	commonfactory "github.com/evmos/evmos/v20/testutil/integration/common/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	integrationutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	testutiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/utils"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

type IntegrationTestSuite struct {
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

var _ = Describe("when sending a Cosmos transaction", Label("AnteHandler"), Ordered, func() {
	var (
		s    *IntegrationTestSuite
		addr sdk.AccAddress
		priv cryptotypes.PrivKey
		msg  sdk.Msg
	)

	BeforeAll(func() {
		keyring := testkeyring.New(3)

		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)
		s = &IntegrationTestSuite{
			network:     integrationNetwork,
			factory:     txFactory,
			grpcHandler: grpcHandler,
			keyring:     keyring,
		}
	})

	Context("and the sender account has enough balance to pay for the transaction cost", Ordered, func() {
		var (
			// rewards are the real accrued rewards
			rewards sdk.DecCoins
			// minExpRewards are the minimun rewards that should be accrued
			// for the test case
			minExpRewards  = sdk.DecCoins{sdk.DecCoin{Amount: sdkmath.LegacyNewDec(1e5), Denom: utils.BaseDenom}}
			delegationCoin = sdk.Coin{Amount: sdkmath.NewInt(1e15), Denom: utils.BaseDenom}
			transferAmt    = sdkmath.NewInt(1e14)
		)

		BeforeEach(func() {
			key := s.keyring.GetKey(0)
			addr = key.AccAddr
			priv = key.Priv

			msg = &banktypes.MsgSend{
				FromAddress: addr.String(),
				ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
				Amount:      sdk.Coins{sdk.Coin{Amount: transferAmt, Denom: utils.BaseDenom}},
			}

			valAddr := s.network.GetValidators()[0].OperatorAddress
			err := s.factory.Delegate(priv, valAddr, delegationCoin)
			Expect(err).To(BeNil())

			rewards, err = integrationutils.WaitToAccrueRewards(s.network, s.grpcHandler, addr.String(), minExpRewards)
			Expect(err).To(BeNil())
		})

		It("should succeed & not withdraw any staking rewards", func() {
			prevBalanceRes, err := s.grpcHandler.GetBalance(addr, s.network.GetDenom())
			Expect(err).To(BeNil())

			baseFeeRes, err := s.grpcHandler.GetBaseFee()
			Expect(err).To(BeNil())

			res, err := s.factory.ExecuteCosmosTx(
				priv,
				commonfactory.CosmosTxArgs{
					Msgs:     []sdk.Msg{msg},
					GasPrice: baseFeeRes.BaseFee,
				},
			)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())

			// include the tx in a block to update state
			err = s.network.NextBlock()
			Expect(err).To(BeNil())

			// fees should be deducted from balance
			feesAmt := sdkmath.NewInt(res.GasWanted).Mul(*baseFeeRes.BaseFee)
			balanceRes, err := s.grpcHandler.GetBalance(addr, s.network.GetDenom())
			Expect(err).To(BeNil())
			Expect(balanceRes.Balance.Amount).To(Equal(prevBalanceRes.Balance.Amount.Sub(transferAmt).Sub(feesAmt)))

			rewardsRes, err := s.grpcHandler.GetDelegationTotalRewards(addr.String())
			Expect(err).To(BeNil())

			// rewards should not be used. Should be more
			// than the previous value queried
			Expect(rewardsRes.Total.Sub(rewards).IsAllPositive()).To(BeTrue())
		})
	})

	Context("and the sender account neither has enough balance nor sufficient staking rewards to pay for the transaction cost", func() {
		BeforeEach(func() {
			addr, priv = testutiltx.NewAccAddressAndKey()

			// this is a new address that does not exist on chain.
			// Transfer 1 aevmos to this account so it is
			// added on chain
			err := s.factory.FundAccount(
				s.keyring.GetKey(0),
				addr,
				sdk.Coins{
					sdk.Coin{
						Amount: sdkmath.NewInt(1),
						Denom:  utils.BaseDenom,
					},
				},
			)
			Expect(err).To(BeNil())
			// persist the state changes
			Expect(s.network.NextBlock()).To(BeNil())

			msg = &banktypes.MsgSend{
				FromAddress: addr.String(),
				ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
				Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(1e14), Denom: utils.BaseDenom}},
			}
		})

		It("should fail", func() {
			var gas uint64 = 200_000 // specify gas to avoid failing on simulation tx (internal call in the ExecuteCosmosTx if gas not specified)
			res, err := s.factory.ExecuteCosmosTx(
				priv,
				commonfactory.CosmosTxArgs{
					Msgs: []sdk.Msg{msg},
					Gas:  &gas,
				},
			)
			Expect(res.IsErr()).To(BeTrue())
			Expect(res.GetLog()).To(ContainSubstring("insufficient funds"))
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())
		})

		It("should not withdraw any staking rewards", func() {
			rewardsRes, err := s.grpcHandler.GetDelegationTotalRewards(addr.String())
			Expect(err).To(BeNil())
			Expect(rewardsRes.Total.Empty()).To(BeTrue())
		})
	})

	Context("and the sender account has not enough balance but sufficient staking rewards to pay for the transaction cost", func() {
		// minExpRewards are the minimun rewards that should be accrued
		// for the test case
		minExpRewards := sdk.DecCoins{sdk.DecCoin{Amount: sdkmath.LegacyNewDec(1e8), Denom: utils.BaseDenom}}

		BeforeEach(func() {
			addr, priv = testutiltx.NewAccAddressAndKey()

			// this is a new address that does not exist on chain.
			// Transfer some funds to stake
			err := s.factory.FundAccount(
				s.keyring.GetKey(0),
				addr,
				sdk.Coins{
					sdk.Coin{
						Amount: sdkmath.NewInt(1e18),
						Denom:  utils.BaseDenom,
					},
				},
			)
			Expect(err).To(BeNil())
			// persist the state changes
			Expect(s.network.NextBlock()).To(BeNil())

			// delegate some tokens and make sure the remaining balance is not sufficient to cover the tx fees
			valAddr := s.network.GetValidators()[1].OperatorAddress
			err = s.factory.Delegate(priv, valAddr, sdk.NewCoin(s.network.GetDenom(), sdkmath.NewInt(9888e14)))
			Expect(err).To(BeNil())

			_, err = integrationutils.WaitToAccrueRewards(s.network, s.grpcHandler, addr.String(), minExpRewards)
			Expect(err).To(BeNil())

			msg = &banktypes.MsgSend{
				FromAddress: addr.String(),
				ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
				Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(1), Denom: utils.BaseDenom}},
			}
		})
	})
})
