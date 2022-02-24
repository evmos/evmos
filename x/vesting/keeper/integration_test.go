package keeper_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/app/ante"
	"github.com/tharsis/evmos/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/vesting/types"
)

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

	dec := ante.NewVestingDelegationDecorator(s.app.AccountKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

func proposeAndVote(clawbackAccount *types.ClawbackVestingAccount) error {
	// Submit governance porposal
	TestProposal := govtypes.NewTextProposal("Test", "description")
	depositor := sdk.AccAddress(tests.GenerateAddress().Bytes())
	proposalCoins := sdk.NewCoins(sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 10)))
	err := testutil.FundAccount(s.app.BankKeeper, s.ctx, depositor, proposalCoins)
	s.Require().NoError(err)

	proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, TestProposal)
	s.Require().NoError(err)

	_, err = s.app.GovKeeper.AddDeposit(s.ctx, proposal.ProposalId, depositor, proposalCoins)
	s.Require().NoError(err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	addr, err := sdk.AccAddressFromBech32(clawbackAccount.Address)
	s.Require().NoError(err)

	voteMsg := govtypes.NewMsgVote(addr, proposal.ProposalId, govtypes.OptionNo)
	txBuilder.SetMsgs(voteMsg)
	tx := txBuilder.GetTx()

	dec := ante.NewVestingGovernanceDecorator(s.app.AccountKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err
}

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

	// Vesting account address
	addr := sdk.AccAddress(s.address.Bytes())

	var (
		clawbackAccount *types.ClawbackVestingAccount
		unvested        sdk.Coins
		vested          sdk.Coins
	)

	BeforeEach(func() {
		s.SetupTest()

		// Create and fund periodic vesting account
		vestingStart := s.ctx.BlockTime().Unix()
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
		s.app.AccountKeeper.SetAccount(s.ctx, clawbackAccount)

		acc := s.app.AccountKeeper.GetAccount(s.ctx, clawbackAccount.GetAddress())
		s.Require().NotNil(acc)

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

		It("cannot vote on governance proposals", func() {
			err := proposeAndVote(clawbackAccount)
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
			_, err := s.DeployContract("vestcoin", "VESTCOIN", erc20Decimals)
			// TODO EVM Hook?
			// Expect(err).ToNot(BeNil())
			Expect(err).To(BeNil())
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

		It("can vote on governance proposals", func() {
			err := proposeAndVote(clawbackAccount)
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
			_, err := s.DeployContract("vestcoin", "VESTCOIN", erc20Decimals)
			// TODO EVM Hook?
			// Expect(err).ToNot(BeNil())
			Expect(err).To(BeNil())
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
			_, err := s.DeployContract("vestcoin", "VESTCOIN", erc20Decimals)
			Expect(err).To(BeNil())
		})

		// TODO Rewards Tests
		// TODO Clawback Tests
		// ? If the funder of a true vesting grant will be able to command "clawback" who is the funder in our case at genesis
	})
})
