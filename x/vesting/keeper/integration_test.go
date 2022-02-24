package keeper_test

import (
	"math/big"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/app/ante"
	"github.com/tharsis/evmos/testutil"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/vesting/types"
)

// Clawback vesting with Cliff and Lock. In this case the cliff is reached
// before the lockup period is reached to represent the scenario in which an
// employee starts before mainnet launch (periodsCliff < lockupPeriod)

// Example:
// 21/10 Employee joins Evmos and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
var _ = Describe("Clawback Vesting Accounts", Ordered, func() {
	// Monthly vesting period
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	amt := sdk.NewInt(1)
	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

	// 4 years vesting total
	periodsTotal := int64(48)
	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(periodsTotal))))

	// 6 month cliff
	cliff := int64(6)
	cliffLength := vestingLength * cliff
	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

	// 12 month lockup
	lockup := int64(12) // 12 year
	lockupLength := vestingLength * lockup
	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: vestingAmtTotal}
	lockupPeriods := sdkvesting.Periods{lockupPeriod}

	// Create vesting periods with initial cliff
	vestingPeriods := sdkvesting.Periods{cliffPeriod}
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	var (
		clawbackAccount *types.ClawbackVestingAccount
		unvested        sdk.Coins
		vested          sdk.Coins
	)

	BeforeEach(func() {
		s.SetupTest()

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime()
		baseAccount := authtypes.NewBaseAccountWithAddress(addr)
		funder := sdk.AccAddress(types.ModuleName)
		clawbackAccount = types.NewClawbackVestingAccount(
			baseAccount,
			funder,
			vestingAmtTotal,
			vestingStart,
			lockupPeriods,
			vestingPeriods,
		)
		err := testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, vestingAmtTotal)
		s.Require().NoError(err)
		acc := s.app.AccountKeeper.NewAccount(s.ctx, clawbackAccount)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)

		// Check if all tokens are unvested at vestingStart
		unvested = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
		vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
		s.Require().Equal(vestingAmtTotal, unvested)
		s.Require().True(vested.IsZero())
	})

	Context("before cliff", func() {
		It("cannot delegate tokens", func() {
			err := delegate(clawbackAccount, 100)
			Expect(err).ToNot(BeNil())
		})

		It("cannot transfer tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				unvested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot perform Ethereum tx", func() {
			err := performEthTx(clawbackAccount)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("after cliff and before lockup", func() {
		BeforeEach(func() {
			// Surpass cliff but not lockup duration
			cliffDuration := time.Duration(cliffLength)
			s.CommitAfter(cliffDuration * time.Second)

			// Check if some, but not all tokens are vested
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(cliff))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)
		})

		It("can delegate vested tokens", func() {
			err := delegate(clawbackAccount, 1)
			Expect(err).To(BeNil())
		})

		It("cannot transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("cannot perform Ethereum tx", func() {
			err := performEthTx(clawbackAccount)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("after cliff and lockup", func() {
		BeforeEach(func() {
			// Surpass lockup duration
			lockupDuration := time.Duration(lockupLength)
			s.CommitAfter(lockupDuration * time.Second)

			// Check if some, but not all tokens are vested
			unvested = clawbackAccount.GetVestingCoins(s.ctx.BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.ctx.BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(sdk.NewInt(lockup))))
			s.Require().NotEqual(vestingAmtTotal, vested)
			s.Require().Equal(expVested, vested)
		})

		It("can delegate vested tokens", func() {
			err := delegate(clawbackAccount, 1)
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			err := delegate(clawbackAccount, 30)
			Expect(err).ToNot(BeNil())
		})

		It("can transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				vested,
			)
			Expect(err).To(BeNil())
		})

		It("cannot transfer unvested tokens", func() {
			err := s.app.BankKeeper.SendCoins(
				s.ctx,
				addr,
				sdk.AccAddress(tests.GenerateAddress().Bytes()),
				unvested,
			)
			Expect(err).ToNot(BeNil())
		})

		It("can perform ethereum tx", func() {
			err := performEthTx(clawbackAccount)
			Expect(err).To(BeNil())
		})
	})
})

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func delegate(clawbackAccount *types.ClawbackVestingAccount, amount int64) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)
	//
	val, err := sdk.ValAddressFromBech32("evmosvaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn")
	s.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(amount)))
	txBuilder.SetMsgs(delegateMsg)
	tx := txBuilder.GetTx()

	dec := ante.NewVestingDelegationDecorator(s.app.AccountKeeper, s.app.StakingKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

func performEthTx(clawbackAccount *types.ClawbackVestingAccount) error {
	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(addr.Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &from, nil, 100000, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgEthereumTx)
	tx := txBuilder.GetTx()

	// Call Ante decorator
	dec := ante.NewEthVestingTransactionDecorator(s.app.AccountKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}
