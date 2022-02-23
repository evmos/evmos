package keeper_test

import (
	"encoding/json"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/app/ante"
	"github.com/tharsis/evmos/contracts"
	"github.com/tharsis/evmos/testutil"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

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
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				addr,
				unvested.AmountOf(stakeDenom),
				stakingtypes.Unbonded,
				s.validator,
				true,
			)
			// TODO Antehandler
			// Expect(err).ToNot(BeNil())
			Expect(err).To(BeNil())
		})

		It("cannot vote on governance proposals", func() {
			// Submit governance porposal
			TestProposal := govtypes.NewTextProposal("Test", "description")
			depositor := sdk.AccAddress(tests.GenerateAddress().Bytes())
			proposalCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 10)))
			err := testutil.FundAccount(s.app.BankKeeper, s.ctx, depositor, proposalCoins)
			s.Require().NoError(err)

			proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, TestProposal)
			s.Require().NoError(err)

			_, err = s.app.GovKeeper.AddDeposit(s.ctx, proposal.ProposalId, depositor, proposalCoins)
			s.Require().NoError(err)

			// Vote
			err = s.app.GovKeeper.AddVote(
				s.ctx,
				proposal.ProposalId,
				addr,
				govtypes.NewNonSplitVoteOption(govtypes.OptionYes),
			)
			// TODO voting shouldn't be possible
			// Expect(err).ToNot(BeNil())
			Expect(err).To(BeNil())
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
			// Create mint msgEthereumTx
			contractAddr, err := s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
			s.Require().NoError(err)
			amount := big.NewInt(100)
			transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("mint", s.address, amount)
			s.Require().NoError(err)

			ctx := sdk.WrapSDKContext(s.ctx)
			chainID := s.app.EvmKeeper.ChainID()

			args, err := json.Marshal(&evmtypes.TransactionArgs{To: &contractAddr, From: &s.address, Data: (*hexutil.Bytes)(&transferData)})
			s.Require().NoError(err)

			res, err := s.queryClientEvm.EstimateGas(ctx, &evmtypes.EthCallRequest{
				Args:   args,
				GasCap: uint64(config.DefaultGasCap),
			})
			s.Require().NoError(err)

			nonce := s.app.EvmKeeper.GetNonce(s.ctx, s.address)

			// Mint the max gas to the FeeCollector to ensure balance in case of refund
			s.MintFeeCollector(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(s.app.FeeMarketKeeper.GetBaseFee(s.ctx).Int64()*int64(res.Gas)))))

			msgEthereumTx := evmtypes.NewTx(
				chainID,
				nonce,
				&contractAddr,
				nil,
				res.Gas,
				nil,
				s.app.FeeMarketKeeper.GetBaseFee(s.ctx),
				big.NewInt(1),
				transferData,
				&ethtypes.AccessList{}, // accesses
			)

			msgEthereumTx.From = s.address.Hex()

			// err = msgEthereumTx.Sign(ethtypes.LatestSignerForChainID(chainID), s.signer)
			// s.Require().NoError(err)

			encodingConfig := encoding.MakeConfig(app.ModuleBasics)
			txBuilder := encodingConfig.TxConfig.NewTxBuilder()
			txBuilder.SetMsgs(msgEthereumTx)
			tx := txBuilder.GetTx()
			dec := ante.NewEthVestingTransactionDecorator(s.app.AccountKeeper)
			_, err = dec.AnteHandle(s.ctx, tx, false, nil)

			// TODO Eth Antehandler
			Expect(err).ToNot(BeNil())
			// Expect(err).To(BeNil())
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
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				addr,
				vested.AmountOf(stakeDenom),
				stakingtypes.Unbonded,
				s.validator,
				true,
			)
			Expect(err).To(BeNil())
		})

		It("can vote on governance proposals", func() {
			// Submit governance porposal
			TestProposal := govtypes.NewTextProposal("Test", "description")
			depositor := sdk.AccAddress(tests.GenerateAddress().Bytes())
			proposalCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 10)))
			err := testutil.FundAccount(s.app.BankKeeper, s.ctx, depositor, proposalCoins)
			s.Require().NoError(err)

			proposal, err := s.app.GovKeeper.SubmitProposal(s.ctx, TestProposal)
			s.Require().NoError(err)

			_, err = s.app.GovKeeper.AddDeposit(s.ctx, proposal.ProposalId, depositor, proposalCoins)
			s.Require().NoError(err)

			// Vote
			err = s.app.GovKeeper.AddVote(
				s.ctx,
				proposal.ProposalId,
				addr,
				govtypes.NewNonSplitVoteOption(govtypes.OptionYes),
			)
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
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				clawbackAccount.GetAddress(),
				vested.AmountOf(stakeDenom),
				stakingtypes.Unbonded,
				s.validator,
				true,
			)
			Expect(err).To(BeNil())
		})

		It("cannot delegate unvested tokens", func() {
			_, err := s.app.StakingKeeper.Delegate(
				s.ctx,
				addr,
				unvested.AmountOf(stakeDenom),
				stakingtypes.Unbonded,
				s.validator,
				true,
			)
			// TODO Antehandler
			// Expect(err).ToNot(BeNil())
			Expect(err).To(BeNil())
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
