package ante_test

import (
	"time"

	inflationtypes "github.com/evmos/evmos/v11/x/inflation/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
)

var _ = Describe("when sending a Cosmos transaction", func() {
	var (
		addr sdk.AccAddress
		priv *ethsecp256k1.PrivKey
		msg  sdk.Msg
	)

	BeforeEach(func() {
		s.SetupTest()

		addr, priv = testutiltx.NewAccAddressAndKey()

		coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1000000000000000000)))
		err := s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
		Expect(err).To(BeNil())
		err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, inflationtypes.ModuleName, addr, coins)
		Expect(err).To(BeNil())

		msg = &banktypes.MsgSend{
			FromAddress: addr.String(),
			ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
			Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(1e14), Denom: utils.BaseDenom}},
		}

	})

	Context("and the sender account has enough balance to pay for the transaction cost", Ordered, func() {
		var (
			rewardsAmt = sdk.NewInt(1e5)
			balance    = sdk.NewInt(1e18)
		)
		BeforeAll(func() {
			var err error
			s.ctx, _ = testutil.PrepareAccountsForDelegationRewards(
				s.T(), s.ctx, s.app, addr, balance, rewardsAmt,
			)
			s.ctx, err = testutil.Commit(s.ctx, s.app, time.Second*0, nil)
			Expect(err).To(BeNil())
		})

		It("should succeed & not withdraw any staking rewards", func() {
			_, err := testutil.DeliverTx(s.ctx, s.app, priv, nil, msg)
			Expect(err).To(BeNil())

			rewards, err := testutil.GetTotalDelegationRewards(s.ctx, s.app.DistrKeeper, addr)
			Expect(err).To(BeNil())
			Expect(rewards).To(Equal(sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, rewardsAmt))))
		})
	})

	Context("and the sender account neither has enough balance nor sufficient staking rewards to pay for the transaction cost", func() {
		It("should fail", func() {
			Expect(false).To(BeFalse())
		})

		It("should not withdraw any staking rewards", func() {
			Expect(false).To(BeFalse())
		})
	})

	Context("and the sender account has not enough balance but sufficient staking rewards to pay for the transaction cost", func() {
		It("should succeed", func() {
			Expect(false).To(BeFalse())
		})

		It("should withdraw enough staking rewards to cover the transaction cost", func() {
			Expect(false).To(BeFalse())
		})

		It("should only withdraw the rewards that are needed to cover the transaction cost", func() {
			Expect(false).To(BeFalse())
		})
	})
})
