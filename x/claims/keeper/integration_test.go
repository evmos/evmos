package keeper_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/testutil"
	"github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/utils"
	"github.com/evmos/evmos/v14/x/claims/types"
	inflationtypes "github.com/evmos/evmos/v14/x/inflation/types"
)

var _ = Describe("Claiming", Ordered, func() {
	claimsAddr := s.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	distrAddr := s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	accountCount := 4

	actionValue := sdk.NewInt(int64(math.Pow10(5) * 10))
	claimValue := actionValue.MulRaw(4)
	totalClaimsAmount := sdk.NewCoin(utils.BaseDenom, claimValue.MulRaw(int64(accountCount)))

	// account initial balances
	initClaimsAmount := sdk.NewInt(types.GenesisDust)
	initBalanceAmount := sdk.NewInt(int64(math.Pow10(18) * 2))
	initStakeAmount := sdk.NewInt(int64(math.Pow10(10) * 2))
	delegateAmount := sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1))
	initBalance := sdk.NewCoins(
		sdk.NewCoin(utils.BaseDenom, initClaimsAmount.Add(initBalanceAmount)), // claimsDenom == evmDenom
	)

	// account for creating the governance proposals
	initClaimsAmount0 := sdk.NewInt(int64(math.Pow10(18) * 2))
	initBalance0 := sdk.NewCoins(
		sdk.NewCoin(stakeDenom, initStakeAmount),
		sdk.NewCoin(utils.BaseDenom, initBalanceAmount.Add(initClaimsAmount0)), // claimsDenom == evmDenom
	)

	var (
		priv0              *ethsecp256k1.PrivKey
		privs              []*ethsecp256k1.PrivKey
		addr0              sdk.AccAddress
		claimsRecords      []types.ClaimsRecord
		params             types.Params
		proposalID         uint64
		totalClaimed       sdk.Coin
		remainderUnclaimed sdk.Coin
		fees               []sdk.Coin
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.ClaimsKeeper.GetParams(s.ctx)
		params.EnableClaims = true
		params.AirdropStartTime = s.ctx.BlockTime()
		err := s.app.ClaimsKeeper.SetParams(s.ctx, params)
		s.Require().NoError(err)

		// mint coins for claiming and send them to the claims module
		coins := sdk.NewCoins(totalClaimsAmount)

		err = testutil.FundModuleAccount(s.ctx, s.app.BankKeeper, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
		s.Require().NoError(err)

		// fund testing accounts and create claim records
		priv0, _ = ethsecp256k1.GenerateKey()
		addr0 = getAddr(priv0)
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, addr0, initBalance0)
		s.Require().NoError(err)

		for i := 0; i < accountCount; i++ {
			priv, _ := ethsecp256k1.GenerateKey()
			privs = append(privs, priv)
			addr := getAddr(priv)
			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, addr, initBalance)
			s.Require().NoError(err)
			claimsRecord := types.NewClaimsRecord(claimValue)
			s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr, claimsRecord)
			acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
			s.app.AccountKeeper.SetAccount(s.ctx, acc)
			claimsRecords = append(claimsRecords, claimsRecord)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom) // claimsDenom == evmDenom == 'aevmos'
			Expect(balance.Amount).To(Equal(initClaimsAmount.Add(initBalanceAmount)))
		}

		// Keep track of the fees paid
		fees = make([]sdk.Coin, len(privs))

		// ensure community pool doesn't have the fund
		poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, utils.BaseDenom)
		Expect(poolBalance.IsZero()).To(BeTrue())

		// ensure module account has the escrow fund
		balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, utils.BaseDenom)
		Expect(balanceClaims).To(Equal(totalClaimsAmount))

		s.Commit()

		proposalID, err = govProposal(priv0)
		s.Require().NoError(err)
	})

	Context("before decay duration", func() {
		var actionV sdk.Coin
		var initialPoolBalance sdk.Coin

		BeforeAll(func() {
			// Community pool will have balance after several blocks because it
			// receives inflation and fees rewards
			initialPoolBalance = s.app.BankKeeper.GetBalance(s.ctx, distrAddr, utils.BaseDenom)
			actionV = sdk.NewCoin(utils.BaseDenom, actionValue)
		})

		It("can claim ActionDelegate", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Delegate(s.ctx, s.app, privs[0], delegateAmount, s.validator)
			s.Require().NoError(err)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Add(actionV).Sub(delegateAmount).Sub(tx.DefaultFee)))
			Expect(balance.Amount).To(Equal(initClaimsAmount.Add(initBalanceAmount).Add(actionV.Amount).Sub(delegateAmount.Amount).Sub(tx.DefaultFee.Amount)))
			fees[0] = tx.DefaultFee
		})

		It("can claim ActionEVM", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			fee := getEthTxFee()
			sendEthToSelf(privs[0])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Add(actionV).Sub(fee)))
			fees[0] = fees[0].Add(fee)
		})

		It("can claim ActionVote", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Vote(s.ctx, s.app, privs[1], proposalID, govv1beta1.OptionAbstain)
			s.Require().NoError(err)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Add(actionV).Sub(tx.DefaultFee)))
			fees[1] = tx.DefaultFee
		})

		It("did not clawback to the community pool", func() {
			// ensure community pool doesn't have the fund
			poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, utils.BaseDenom)
			Expect((poolBalance.Sub(initialPoolBalance)).IsZero()).To(BeTrue())

			// ensure module account has the escrow fund minus what was claimed
			balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, utils.BaseDenom)
			totalClaimed = sdk.NewCoin(utils.BaseDenom, actionV.Amount.MulRaw(3))
			Expect(balanceClaims).To(Equal(totalClaimsAmount.Sub(totalClaimed)))
		})
	})

	Context("at 2/3 decay duration", func() {
		var actionV sdk.Coin
		var unclaimedV sdk.Coin
		var initialPoolBalance sdk.Coin

		BeforeAll(func() {
			actionV = sdk.NewCoin(utils.BaseDenom, actionValue.QuoRaw(3))
			unclaimedV = sdk.NewCoin(utils.BaseDenom, actionValue.Sub(actionV.Amount))
			duration := params.DecayStartTime().Sub(s.ctx.BlockHeader().Time)

			s.CommitAfter(duration)
			duration = params.GetDurationOfDecay() * 2 / 3

			// create another proposal to vote for
			testTime := s.ctx.BlockHeader().Time.Add(duration)
			s.CommitAfter(duration - time.Hour)

			var err error
			proposalID, err = govProposal(priv0)
			s.Require().NoError(err)
			s.CommitAfter(testTime.Sub(s.ctx.BlockHeader().Time))

			// Community pool will have balance after several blocks because it
			// receives inflation and fees rewards
			initialPoolBalance = s.app.BankKeeper.GetBalance(s.ctx, distrAddr, utils.BaseDenom)
		})

		It("can claim ActionDelegate", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Delegate(s.ctx, s.app, privs[1], delegateAmount, s.validator)
			s.Require().NoError(err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Add(actionV).Sub(delegateAmount).Sub(tx.DefaultFee)))
			fees[1] = fees[1].Add(tx.DefaultFee)
		})

		It("can claim ActionEVM", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			fee := getEthTxFee()
			sendEthToSelf(privs[1])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Add(actionV).Sub(fee)))
			fees[1] = fees[1].Add(fee)
			fee = getEthTxFee()
			sendEthToSelf(privs[2])
			fees[2] = fee
		})

		It("can claim ActionVote", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Vote(s.ctx, s.app, privs[0], proposalID, govv1beta1.OptionAbstain)
			s.Require().NoError(err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Add(actionV).Sub(tx.DefaultFee)))
			fees[0] = fees[0].Add(tx.DefaultFee)
		})

		It("cannot claim ActionDelegate a second time", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Delegate(s.ctx, s.app, privs[1], delegateAmount, s.validator)
			s.Require().NoError(err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Sub(delegateAmount).Sub(tx.DefaultFee)))
			fees[1] = fees[1].Add(tx.DefaultFee)
		})

		It("cannot claim ActionEVM a second time", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			fee := getEthTxFee()
			sendEthToSelf(privs[1])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Sub(fee)))
			fees[1] = fees[1].Add(fee)
		})

		It("cannot claim ActionVote a second time", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Vote(s.ctx, s.app, privs[0], proposalID, govv1beta1.OptionAbstain)
			s.Require().NoError(err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Sub(tx.DefaultFee)))
			fees[0] = fees[0].Add(tx.DefaultFee)
		})

		It("did not clawback to the community pool", func() {
			remainderUnclaimed = sdk.NewCoin(utils.BaseDenom, unclaimedV.Amount.MulRaw(4))
			totalClaimed = totalClaimed.Add(sdk.NewCoin(utils.BaseDenom, actionV.Amount.MulRaw(4)))

			// ensure community pool doesn't have the fund
			poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, utils.BaseDenom)
			Expect(poolBalance.Sub(initialPoolBalance)).To(Equal(remainderUnclaimed))

			// ensure module account has the escrow fund minus what was claimed
			balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, utils.BaseDenom)
			Expect(balanceClaims).To(Equal(totalClaimsAmount.Sub(totalClaimed).Sub(remainderUnclaimed)))
		})
	})

	Context("after decay duration", func() {
		BeforeAll(func() {
			duration := params.AirdropEndTime().Sub(s.ctx.BlockHeader().Time) + 1
			s.CommitAfter(duration)

			// ensure module account has the unclaimed amount before airdrop ends
			moduleBalances := s.app.ClaimsKeeper.GetModuleAccountBalances(s.ctx)
			Expect(moduleBalances.AmountOf(utils.BaseDenom)).To(Equal(totalClaimsAmount.Sub(totalClaimed).Sub(remainderUnclaimed).Amount))

			s.Commit()
		})

		It("cannot claim additional actions", func() {
			addr := getAddr(privs[2])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			_, err := testutil.Delegate(s.ctx, s.app, privs[2], delegateAmount, s.validator)
			s.Require().NoError(err)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(balance).To(Equal(prebalance.Sub(delegateAmount).Sub(tx.DefaultFee)))
			fees[2] = fees[2].Add(tx.DefaultFee)
		})

		It("cannot clawback already claimed actions", func() {
			addr := getAddr(privs[0])
			finalBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			claimed := actionValue.MulRaw(2).Add(actionValue.QuoRaw(3))
			Expect(finalBalance.Amount).To(Equal(initClaimsAmount.Add(initBalanceAmount).Add(claimed).Sub(delegateAmount.Amount).Sub(fees[0].Amount)))

			addr = getAddr(privs[1])
			finalBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			claimed = actionValue.MulRaw(2).QuoRaw(3).Add(actionValue)
			Expect(finalBalance.Amount).To(Equal(initClaimsAmount.Add(initBalanceAmount).Add(claimed).Sub(delegateAmount.Amount.MulRaw(2)).Sub(fees[1].Amount)))

			addr = getAddr(privs[2])
			finalBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			claimed = actionValue.QuoRaw(3)
			Expect(finalBalance.Amount).To(Equal(initClaimsAmount.Add(initBalanceAmount).Add(claimed).Sub(delegateAmount.Amount).Sub(fees[2].Amount)))

			// no-op, should have same balance as initial balance
			addr = getAddr(privs[3])
			finalBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, utils.BaseDenom)
			Expect(finalBalance.Amount).To(Equal(initClaimsAmount.Add(initBalanceAmount)))
		})

		It("can clawback unclaimed", func() {
			// ensure module account is empty
			moduleBalance := s.app.ClaimsKeeper.GetModuleAccountBalances(s.ctx)
			Expect(moduleBalance.AmountOf(utils.BaseDenom).IsZero()).To(BeTrue())
		})
	})
})
