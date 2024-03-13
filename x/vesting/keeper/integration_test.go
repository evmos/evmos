package keeper_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/encoding"
	"github.com/evmos/evmos/v16/testutil/integration/common/factory"
	evmosfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"github.com/evmos/evmos/v16/x/vesting/types"
)

func TestKeeperTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

// TestClawbackAccount is a struct to store all relevant information that is corresponding
// to a clawback vesting account.
type TestClawbackAccount struct {
	privKey         cryptotypes.PrivKey
	address         sdk.AccAddress
	clawbackAccount *types.ClawbackVestingAccount
}

// Initialize general error variable for easier handling in loops throughout this test suite.
var err error

// Clawback vesting with Cliff and Lock. In this case the cliff is reached
// before the lockup period is reached to represent the scenario in which an
// employee starts before mainnet launch (periodsCliff < lockupPeriod)
//
// Example:
// 21/10 Employee joins Evmos and vesting starts
// 22/03 Mainnet launch
// 22/09 Cliff ends
// 23/02 Lock ends
var _ = Describe("Clawback Vesting Accounts", Ordered, func() {
	var (
		s                 *KeeperTestSuite
		funder            keyring.Key
		vestingAccs       []keyring.Key
		clawbackAccount   *types.ClawbackVestingAccount
		unvested          sdk.Coins
		vested            sdk.Coins
		twoThirdsOfVested sdk.Coins
	)

	// Initialized vars
	var (
		numTestMsgs                     = 3
		vestingAccInitialBalance        = network.PrefundedAccountInitialBalance
		gasPrice                        = math.NewInt(200_000_000)
		remainingAmtToPayFees           = math.NewInt(1e16)
		gasLimit                 uint64 = 200_000
		dest                            = utiltx.GenerateAddress()

		// Monthly vesting period
		stakeDenom    = utils.BaseDenom
		amt           = math.NewInt(1e17)
		vestingLength = int64(60 * 60 * 24 * 30) // 30 days in seconds
		vestingAmt    = sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
		vestingPeriod = sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

		// 4 years vesting total
		periodsTotal    = int64(48)
		vestingAmtTotal = sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))

		// 6 month cliff
		cliff       = int64(6)
		cliffLength = vestingLength * cliff
		cliffAmt    = sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
		cliffPeriod = sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

		// 12 month lockup
		lockup       = int64(12) // 12 months
		lockupLength = vestingLength * lockup

		// Unlock every 12 months
		numLockupPeriods = int64(4)

		// Unlock 1/4th of the total vest in each unlock event. By default, all tokens are
		// unlocked after surpassing the final period.
		unlockedPerLockup    = vestingAmtTotal.QuoInt(math.NewInt(numLockupPeriods))
		unlockedPerLockupAmt = unlockedPerLockup[0].Amount
		lockupPeriod         = sdkvesting.Period{Length: lockupLength, Amount: unlockedPerLockup}
		lockupPeriods        = make(sdkvesting.Periods, numLockupPeriods)
		vestingPeriods       = sdkvesting.Periods{cliffPeriod}
	)

	for i := range lockupPeriods {
		lockupPeriods[i] = lockupPeriod
	}

	// Create vesting periods with initial cliff
	for p := int64(1); p <= periodsTotal-cliff; p++ {
		vestingPeriods = append(vestingPeriods, vestingPeriod)
	}

	BeforeEach(func() {
		s = new(KeeperTestSuite)
		// create 5 prefunded accounts:
		// index 0-3 will be vesting accounts
		// and index 4 will be the funder
		keys := keyring.New(5)
		nw := network.NewUnitTestNetwork(
			network.WithPreFundedAccounts(keys.GetAllAccAddrs()...),
		)
		gh := grpc.NewIntegrationHandler(nw)
		tf := evmosfactory.New(nw, gh)

		s.network = nw
		s.factory = tf
		s.handler = gh
		s.keyring = keys

		// index 0 will be the funder
		// index 1-4 will be vesting accounts
		funder = keys.GetKey(0)
		vestingAccs = keys.GetKeys()[1:4]

		// Initialize all vesting accounts
		for _, account := range vestingAccs {
			// Create and fund periodic vesting account

			// send a create vesting account tx
			createAccMsg := types.NewMsgCreateClawbackVestingAccount(funder.AccAddr, account.AccAddr, false)
			res, err := s.factory.ExecuteCosmosTx(account.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{createAccMsg}})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// Fund the clawback vesting accounts
			vestingStart := s.network.GetContext().BlockTime()
			fundMsg := types.NewMsgFundVestingAccount(funder.AccAddr, account.AccAddr, vestingStart, lockupPeriods, vestingPeriods)
			res, err = s.factory.ExecuteCosmosTx(funder.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{fundMsg}})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			acc, err := s.handler.GetAccount(account.AccAddr.String())
			Expect(err).To(BeNil())
			var ok bool
			clawbackAccount, ok = acc.(*types.ClawbackVestingAccount)
			Expect(ok).To(BeTrue())

			// Check if all tokens are unvested at vestingStart
			unvested = clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			Expect(vestingAmtTotal).To(Equal(unvested))
			Expect(vested.IsZero()).To(BeTrue())
		}
	})

	Context("before first vesting period", func() {
		BeforeEach(func() {
			// Ensure no tokens are vested
			vested := clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			unlocked := clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.ZeroInt()))
			Expect(zeroCoins).To(Equal(vested))
			Expect(zeroCoins).To(Equal(unlocked))
		})

		It("cannot delegate tokens", func() {
			err := s.factory.Delegate(
				vestingAccs[0].Priv,
				s.network.GetValidators()[0].OperatorAddress,
				sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Add(math.NewInt(1))),
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("account has no vested coins"))
		})

		It("can transfer spendable tokens", func() {
			account := vestingAccs[0]
			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			sendAmt := vestingAccInitialBalance.Sub(remainingAmtToPayFees)
			spendableCoin := sdk.NewCoin(stakeDenom, sendAmt)
			coins := sdk.NewCoins(spendableCoin)
			msg := banktypes.NewMsgSend(account.AccAddr, dest.Bytes(), coins)
			res, err := s.factory.ExecuteCosmosTx(account.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, GasPrice: &gasPrice})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasWanted))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(sendAmt)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(sendAmt))
		})

		It("cannot transfer unvested tokens", func() {
			account := vestingAccs[0]
			coins := unvested.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))
			msg := banktypes.NewMsgSend(account.AccAddr, dest.Bytes(), coins)
			_, err = s.factory.ExecuteCosmosTx(account.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, GasPrice: &gasPrice})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("spendable balance"))
			Expect(err.Error()).To(ContainSubstring("is smaller than"))
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := vestingAccs[0]

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			sendAmt := vestingAccInitialBalance.Sub(remainingAmtToPayFees)
			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), Amount: sendAmt.BigInt()})
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasWanted))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(sendAmt)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(sendAmt))
		})

		It("cannot perform Ethereum tx with unvested balance", func() {
			account := vestingAccs[0]
			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unvested.AmountOf(stakeDenom)).BigInt()

			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), Amount: txAmount})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("insufficient unlocked tokens"))
			Expect(res.IsOK()).To(BeFalse())
		})
	})

	Context("after first vesting period and before lockup", func() {
		BeforeEach(func() {
			// Surpass cliff but none of lockup duration
			cliffDuration := time.Duration(cliffLength)
			Expect(s.network.NextBlockAfter(cliffDuration * time.Second)).To(BeNil())

			acc, err := s.handler.GetAccount(vestingAccs[0].AccAddr.String())
			Expect(err).To(BeNil())
			var ok bool
			clawbackAccount, ok = acc.(*types.ClawbackVestingAccount)
			Expect(ok).To(BeTrue())

			// Check if some, but not all tokens are vested
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
			Expect(vestingAmtTotal).NotTo(Equal(vested))
			Expect(expVested).To(Equal(vested))

			twoThirdsOfVested = vested.Sub(vested.QuoInt(math.NewInt(3))...)
		})

		It("can delegate vested tokens and update spendable balance", func() {
			account := vestingAccs[0]
			// Verify that the total spendable coins decreases after staking
			// vested tokens.
			res, err := s.handler.GetSpendableBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			spendablePre := res.Balance

			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())

			delegationCoin := vestedCoin.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))
			err = s.factory.Delegate(
				account.Priv,
				s.network.GetValidators()[0].OperatorAddress,
				delegationCoin,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			res, err = s.handler.GetSpendableBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			spendablePost := res.Balance
			Expect(spendablePost.Amount.GT(spendablePre.Amount))
		})

		It("cannot delegate unvested tokens", func() {
			ok, vestedCoin := vestingAmtTotal.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			delegationCoin := vestedCoin.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))
			err := s.factory.Delegate(
				vestingAccs[0].Priv,
				s.network.GetValidators()[0].OperatorAddress,
				delegationCoin,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("cannot delegate unvested coins"))
		})

		It("cannot delegate and then send tokens", func() {
			account := vestingAccs[0]

			err = s.factory.Delegate(
				account.Priv,
				s.network.GetValidators()[0].OperatorAddress,
				twoThirdsOfVested[0],
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			sendCoins := twoThirdsOfVested.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))

			msg := banktypes.NewMsgSend(account.AccAddr, dest.Bytes(), sendCoins)
			_, err = s.factory.ExecuteCosmosTx(account.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, GasPrice: &gasPrice})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("spendable balance"))
			Expect(err.Error()).To(ContainSubstring("is smaller than"))
		})

		It("cannot transfer vested tokens", func() {
			msg := banktypes.NewMsgSend(vestingAccs[0].AccAddr, dest.Bytes(), vested.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees))))
			_, err = s.factory.ExecuteCosmosTx(vestingAccs[0].Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, GasPrice: &gasPrice})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("spendable balance"))
			Expect(err.Error()).To(ContainSubstring("is smaller than"))
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := vestingAccs[0]

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees)
			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasWanted))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(txAmount)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(txAmount))
		})

		It("cannot perform Ethereum tx with locked balance", func() {
			account := vestingAccs[0]
			txAmount := vestingAccInitialBalance.Add(vested.AmountOf(stakeDenom)).Sub(remainingAmtToPayFees)
			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), Amount: txAmount.BigInt()})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
			Expect(res.IsOK()).To(BeFalse())
		})
	})

	Context("Between first and second lockup periods", func() {
		BeforeEach(func() {
			// Surpass first lockup
			vestDuration := time.Duration(lockupLength)
			Expect(s.network.NextBlockAfter(vestDuration * time.Second)).To(BeNil())

			// Check if some, but not all tokens are vested and unlocked
			for _, account := range vestingAccs {
				acc, err := s.handler.GetAccount(account.AccAddr.String())
				Expect(err).To(BeNil())
				vestAcc, ok := acc.(*types.ClawbackVestingAccount)
				Expect(ok).To(BeTrue())

				vested := vestAcc.GetVestedOnly(s.network.GetContext().BlockTime())
				unlocked := vestAcc.GetUnlockedOnly(s.network.GetContext().BlockTime())
				expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup))))

				Expect(vestingAmtTotal).ToNot(Equal(vested))
				Expect(expVested).To(Equal(vested))
				Expect(unlocked).To(Equal(unlockedPerLockup))
			}
		})

		It("should enable access to unlocked EVM tokens (single-account, single-msg)", func() {
			account := vestingAccs[0]

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unlockedPerLockupAmt)
			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasWanted))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(txAmount)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(txAmount))
		})

		It("should enable access to unlocked EVM tokens (single-account, multiple-msgs)", func() {
			account := vestingAccs[0]

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			// Split the total unlocked amount into numTestMsgs equally sized tx's
			msgs := make([]sdk.Msg, numTestMsgs)
			// send all the account's spendable balance
			// initial_balance + unlocked in several messages
			totalSendAmt := vestingAccInitialBalance.Add(unlockedPerLockupAmt).Sub(remainingAmtToPayFees)
			txAmount := totalSendAmt.QuoRaw(int64(numTestMsgs))

			for i := 0; i < numTestMsgs; i++ {
				msg, err := s.factory.GenerateSignedMsgEthereumTx(account.Priv, evmtypes.EvmTxArgs{Nonce: uint64(i + 1), To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
				Expect(err).To(BeNil())
				msgs[i] = &msg
			}

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(totalSendAmt)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(totalSendAmt))
		})

		It("should enable access to unlocked EVM tokens (multi-account, single-tx)", func() {

			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unlockedPerLockupAmt)

			msgs := make([]sdk.Msg, len(vestingAccs))
			for i, grantee := range vestingAccs {
				msg, err := s.factory.GenerateSignedMsgEthereumTx(grantee.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
				Expect(err).To(BeNil())
				msgs[i] = &msg
			}

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			balRes, err := s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(txAmount.MulRaw(int64(len(vestingAccs)))))
		})

		It("should enable access to unlocked EVM tokens (multi-account, multiple-msgs)", func() {
			amtSentByAcc := vestingAccInitialBalance.Add(unlockedPerLockupAmt).Sub(remainingAmtToPayFees)
			txAmount := amtSentByAcc.QuoRaw(int64(numTestMsgs))

			msgs := []sdk.Msg{}
			for _, grantee := range vestingAccs {
				for i := 0; i < numTestMsgs; i++ {
					msg, err := s.factory.GenerateSignedMsgEthereumTx(grantee.Priv, evmtypes.EvmTxArgs{Nonce: uint64(i + 1), To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
					Expect(err).To(BeNil())
					msgs = append(msgs, &msg)
				}
			}

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			balRes, err := s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(amtSentByAcc.MulRaw(int64(len(vestingAccs)))))
		})

		It("should not enable access to locked EVM tokens (single-account, single-msg)", func() {
			testAccount := vestingAccs[0]
			// Attempt to spend entire vesting balance
			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(vestingAmtTotal.AmountOf(stakeDenom))

			msg, err := s.factory.GenerateSignedMsgEthereumTx(testAccount.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, &msg)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsErr()).To(BeTrue())
			Expect(res.Log).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
		})

		It("should not enable access to locked EVM tokens (single-account, multiple-msgs)", func() {
			account := vestingAccs[0]
			msgs := make([]sdk.Msg, numTestMsgs+1)
			amt := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unlockedPerLockupAmt)
			txAmount := amt.QuoRaw(int64(numTestMsgs))

			// Add additional message that exceeds unlocked balance
			for i := 0; i < numTestMsgs+1; i++ {
				msg, err := s.factory.GenerateSignedMsgEthereumTx(account.Priv, evmtypes.EvmTxArgs{Nonce: uint64(i + 1), To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
				Expect(err).To(BeNil())
				msgs[i] = &msg
			}

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsErr()).To(BeTrue())
			Expect(res.Log).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
		})

		It("should not enable access to locked EVM tokens (multi-account, single-msg)", func() {
			numVestAccounts := len(vestingAccs)
			msgs := make([]sdk.Msg, numVestAccounts+1)
			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unlockedPerLockupAmt)

			for i, account := range vestingAccs {
				msg, err := s.factory.GenerateSignedMsgEthereumTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
				Expect(err).To(BeNil())
				msgs[i] = &msg
			}

			// Add additional message that exceeds unlocked balance
			msg, err := s.factory.GenerateSignedMsgEthereumTx(vestingAccs[0].Priv, evmtypes.EvmTxArgs{Nonce: uint64(2), To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			msgs[numVestAccounts] = &msg

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsErr()).To(BeTrue())
			Expect(res.Log).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
		})

		It("should not enable access to locked EVM tokens (multi-account, multiple-msgs)", func() {
			msgs := []sdk.Msg{}
			amt := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unlockedPerLockupAmt)
			txAmount := amt.QuoRaw(int64(numTestMsgs))

			for _, account := range vestingAccs {
				for i := 0; i < numTestMsgs; i++ {
					msg, err := s.factory.GenerateSignedMsgEthereumTx(account.Priv, evmtypes.EvmTxArgs{Nonce: uint64(i + 1), To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
					Expect(err).To(BeNil())
					msgs = append(msgs, &msg)
				}
			}

			// Add additional message that exceeds unlocked balance
			msg, err := s.factory.GenerateSignedMsgEthereumTx(vestingAccs[0].Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			msgs = append(msgs, &msg)

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsErr()).To(BeTrue())
			Expect(res.Log).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
		})

		It("should not short-circuit with a normal account", func() {
			vestAcc := vestingAccs[0]
			normalAcc := funder

			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(vestingAmtTotal.AmountOf(stakeDenom))

			// Get message from a normal account to try to short-circuit the AnteHandler
			normAccMsg, err := s.factory.GenerateSignedMsgEthereumTx(normalAcc.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: big.NewInt(100_000)})
			Expect(err).To(BeNil())

			// Attempt to spend entire balance
			vestAccMsg, err := s.factory.GenerateSignedMsgEthereumTx(vestAcc.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())

			msgs := []sdk.Msg{&normAccMsg, &vestAccMsg}

			txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
			tx, err := utiltx.PrepareEthTx(txConfig, s.network.App, nil, msgs...)
			Expect(err).To(BeNil())

			txBytes, err := txConfig.TxEncoder()(tx)
			Expect(err).To(BeNil())

			res, err := s.network.BroadcastTxSync(txBytes)
			Expect(err).To(BeNil())
			Expect(res.IsErr()).To(BeTrue())
			Expect(res.Log).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
		})
	})

	Context("after first lockup and additional vest", func() {
		BeforeEach(func() {
			vestDuration := time.Duration(lockupLength + vestingLength)
			s.network.NextBlockAfter(vestDuration * time.Second)

			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup+1))))

			unlocked := clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
			expUnlocked := unlockedPerLockup

			Expect(expVested).To(Equal(vested))
			Expect(expUnlocked).To(Equal(unlocked))
		})

		It("should enable access to unlocked EVM tokens", func() {
			account := vestingAccs[0]
			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(unlockedPerLockupAmt)

			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(txAmount)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(txAmount))
		})

		It("should not enable access to locked EVM tokens", func() {
			account := vestingAccs[0]

			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(vested.AmountOf(stakeDenom))

			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("clawback vesting account has insufficient unlocked tokens to execute transaction"))
			Expect(res.IsErr()).To(BeTrue())
		})
	})

	Context("after half of vesting period and half lockups", func() {
		BeforeEach(func() {
			// Surpass half lockup duration
			passedLockups := numLockupPeriods / 2
			twoLockupsDuration := time.Duration(lockupLength * passedLockups)
			s.network.NextBlockAfter(twoLockupsDuration * time.Second)

			// Check if some, but not all tokens are vested
			unvested = clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(lockup*passedLockups))))
			Expect(vestingAmtTotal).NotTo(Equal(vested))
			Expect(expVested).To(Equal(vested))
		})

		It("can delegate vested tokens", func() {
			account := vestingAccs[0]
			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			// save some balance to pay fees
			delCoin := vestedCoin.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))
			err = s.factory.Delegate(
				account.Priv,
				s.network.GetValidators()[0].OperatorAddress,
				delCoin,
			)
			Expect(err).To(BeNil())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			// remaining balance should be less than prevBalance - delegated amount
			// cause should pay for fees too
			Expect(balancePost.Amount.LT(balancePrev.Amount.Sub(delCoin.Amount))).To(BeTrue())

		})

		It("cannot delegate unvested tokens", func() {
			ok, vestedCoin := vestingAmtTotal.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			delCoin := vestedCoin.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))
			err := s.factory.Delegate(
				vestingAccs[0].Priv,
				s.network.GetValidators()[0].OperatorAddress,
				delCoin,
			)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("cannot delegate unvested coins"))
		})

		It("can transfer vested tokens", func() {
			account := vestingAccs[0]

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			// save some balance to pay fees
			coins := vested.Add(sdk.NewCoin(stakeDenom, vestingAccInitialBalance.Sub(remainingAmtToPayFees)))
			sendAmt := coins[0].Amount

			msg := banktypes.NewMsgSend(account.AccAddr, dest.Bytes(), coins)
			res, err := s.factory.ExecuteCosmosTx(account.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, GasPrice: &gasPrice})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasWanted))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(sendAmt)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(sendAmt))
		})

		It("cannot transfer unvested tokens", func() {
			account := vestingAccs[0]
			// save some balance to pay fees
			sendAmt := vestingAccInitialBalance.Sub(remainingAmtToPayFees)
			coins := vestingAmtTotal.Add(sdk.NewCoin(stakeDenom, sendAmt))

			msg := banktypes.NewMsgSend(account.AccAddr, dest.Bytes(), coins)
			_, err := s.factory.ExecuteCosmosTx(account.Priv, factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}, GasPrice: &gasPrice})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("spendable balance"))
			Expect(err.Error()).To(ContainSubstring("is smaller than"))
		})

		It("can perform Ethereum tx with spendable balance", func() {
			account := vestingAccs[0]
			// save some balance to pay fees
			ok, vestedCoin := vested.Find(utils.BaseDenom)
			Expect(ok).To(BeTrue())
			txAmount := vestingAccInitialBalance.Sub(remainingAmtToPayFees).Add(vestedCoin.Amount)

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(txAmount)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(txAmount))
		})
	})

	Context("after entire vesting period and all lockups", func() {
		BeforeEach(func() {
			// Surpass vest duration
			vestDuration := time.Duration(vestingLength * periodsTotal)
			s.network.NextBlockAfter(vestDuration * time.Second)

			// Check that all tokens are vested and unlocked
			unvested = clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())
			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
			locked := clawbackAccount.LockedCoins(s.network.GetContext().BlockTime())

			zeroCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.ZeroInt()))
			Expect(vestingAmtTotal).To(Equal(vested))
			Expect(zeroCoins).To(Equal(locked))
			Expect(zeroCoins).To(Equal(unvested))
		})

		It("can send entire balance", func() {
			account := vestingAccs[0]

			balRes, err := s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePrev := balRes.Balance

			txAmount := vestingAmtTotal.AmountOf(stakeDenom).Add(vestingAccInitialBalance.Sub(remainingAmtToPayFees))

			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
			Expect(s.network.NextBlock()).To(BeNil())

			// check final balance is as expected - transferred spendable tokens
			fees := gasPrice.Mul(math.NewInt(res.GasUsed))
			balRes, err = s.handler.GetBalance(account.AccAddr, stakeDenom)
			Expect(err).To(BeNil())
			balancePost := balRes.Balance
			Expect(balancePost.Amount).To(Equal(balancePrev.Amount.Sub(fees).Sub(txAmount)))

			balRes, err = s.handler.GetBalance(dest.Bytes(), stakeDenom)
			Expect(err).To(BeNil())
			destBalance := balRes.Balance
			Expect(destBalance.Amount).To(Equal(txAmount))
		})

		It("cannot exceed balance", func() {
			account := vestingAccs[0]

			txAmount := vestingAmtTotal.AmountOf(stakeDenom).Add(vestingAccInitialBalance).Mul(math.NewInt(2))
			res, err := s.factory.ExecuteEthTx(account.Priv, evmtypes.EvmTxArgs{To: &dest, GasPrice: gasPrice.BigInt(), GasLimit: gasLimit, Amount: txAmount.BigInt()})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("sender balance < tx cost"))
			Expect(res.IsErr()).To(BeTrue())
		})
	})
})

// // Example:
// // 21/10 Employee joins Evmos and vesting starts
// // 22/03 Mainnet launch
// // 22/09 Cliff ends
// // 23/02 Lock ends
// var _ = Describe("Clawback Vesting Accounts - claw back tokens", func() {
// 	var s *KeeperTestSuite
// 	// Monthly vesting period
// 	stakeDenom := utils.BaseDenom
// 	amt := math.NewInt(1)
// 	vestingLength := int64(60 * 60 * 24 * 30) // in seconds
// 	vestingAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt))
// 	vestingPeriod := sdkvesting.Period{Length: vestingLength, Amount: vestingAmt}

// 	// 4 years vesting total
// 	periodsTotal := int64(48)
// 	vestingTotal := amt.Mul(math.NewInt(periodsTotal))
// 	vestingAmtTotal := sdk.NewCoins(sdk.NewCoin(stakeDenom, vestingTotal))

// 	// 6 month cliff
// 	cliff := int64(6)
// 	cliffLength := vestingLength * cliff
// 	cliffAmt := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(cliff))))
// 	cliffPeriod := sdkvesting.Period{Length: cliffLength, Amount: cliffAmt}

// 	// 12 month lockup
// 	lockup := int64(12) // 12 year
// 	lockupLength := vestingLength * lockup
// 	lockupPeriod := sdkvesting.Period{Length: lockupLength, Amount: vestingAmtTotal}
// 	lockupPeriods := sdkvesting.Periods{lockupPeriod}

// 	// Create vesting periods with initial cliff
// 	vestingPeriods := sdkvesting.Periods{cliffPeriod}
// 	for p := int64(1); p <= periodsTotal-cliff; p++ {
// 		vestingPeriods = append(vestingPeriods, vestingPeriod)
// 	}

// 	var (
// 		clawbackAccount *types.ClawbackVestingAccount
// 		vesting         sdk.Coins
// 		vested          sdk.Coins
// 		unlocked        sdk.Coins
// 		free            sdk.Coins
// 		isClawback      bool
// 	)

// 	vestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
// 	funder, funderPriv := utiltx.NewAccAddressAndKey()
// 	dest := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

// 	BeforeEach(func() {
// 		s = new(KeeperTestSuite)
// 		s.SetupTest()
// 		vestingStart := s.network.GetContext().BlockTime()

// 		// Initialize account at vesting address by funding it with tokens
// 		// and then send them over to the vesting funder
// 		err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, vestingAddr, vestingAmtTotal)
// 		Expect(err).ToNot(HaveOccurred(), "failed to fund target account")
// 		err = s.network.App.BankKeeper.SendCoins(s.network.GetContext(), vestingAddr, funder, vestingAmtTotal)
// 		Expect(err).ToNot(HaveOccurred(), "failed to send coins to funder")

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceDest := balRes.Balance
// 		Expect(balanceFunder).To(Equal(vestingAmtTotal[0]), "expected different funder balance")
// 		Expect(balanceGrantee.IsZero()).To(BeTrue(), "expected balance of vesting account to be zero")
// 		Expect(balanceDest.IsZero()).To(BeTrue(), "expected destination balance to be zero")

// 		msg := types.NewMsgCreateClawbackVestingAccount(funder, vestingAddr, true)

// 		_, err = s.network.App.VestingKeeper.CreateClawbackVestingAccount(s.network.GetContext(), msg)
// 		Expect(err).ToNot(HaveOccurred(), "expected creating clawback vesting account to succeed")

// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 		clawbackAccount, isClawback = acc.(*types.ClawbackVestingAccount)
// 		Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

// 		// fund the vesting account
// 		msgFund := types.NewMsgFundVestingAccount(funder, vestingAddr, vestingStart, lockupPeriods, vestingPeriods)
// 		_, err = s.network.App.VestingKeeper.FundVestingAccount(s.network.GetContext(), msgFund)
// 		Expect(err).ToNot(HaveOccurred(), "expected funding vesting account to succeed")

// 		acc = s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 		Expect(acc).ToNot(BeNil(), "expected account to exist")
// 		clawbackAccount, isClawback = acc.(*types.ClawbackVestingAccount)
// 		Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

// 		// Check if all tokens are unvested and locked at vestingStart
// 		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
// 		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 		Expect(vesting).To(Equal(vestingAmtTotal), "expected difference vesting tokens")
// 		Expect(vested.IsZero()).To(BeTrue(), "expected no tokens to be vested")
// 		Expect(unlocked.IsZero()).To(BeTrue(), "expected no tokens to be unlocked")

// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee = balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceDest = balRes.Balance

// 		Expect(bF.IsZero()).To(BeTrue(), "expected funder balance to be zero")
// 		Expect(balanceGrantee).To(Equal(vestingAmtTotal[0]), "expected all tokens to be locked")
// 		Expect(balanceDest.IsZero()).To(BeTrue(), "expected no tokens to be unlocked")
// 	})

// 	It("should fail if there is no vesting or lockup schedule set", func() {
// 		ctx := s.network.GetContext()
// 		emptyVestingAddr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
// 		err := testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, emptyVestingAddr, vestingAmtTotal)
// 		Expect(err).ToNot(HaveOccurred(), "failed to fund target account")

// 		msg := types.NewMsgCreateClawbackVestingAccount(funder, emptyVestingAddr, false)

// 		_, err = s.network.App.VestingKeeper.CreateClawbackVestingAccount(s.network.GetContext(), msg)
// 		Expect(err).ToNot(HaveOccurred(), "expected creating clawback vesting account to succeed")

// 		clawbackMsg := types.NewMsgClawback(funder, emptyVestingAddr, dest)
// 		_, err = s.network.App.VestingKeeper.Clawback(ctx, clawbackMsg)
// 		Expect(err).To(HaveOccurred())
// 		Expect(err.Error()).To(ContainSubstring("has no vesting or lockup periods"))
// 	})

// 	It("should claw back unvested amount before cliff", func() {
// 		ctx := s.network.GetContext()

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceDest := balRes.Balance

// 		// Perform clawback before cliff
// 		msg := types.NewMsgClawback(funder, vestingAddr, dest)
// 		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 		Expect(err).To(BeNil())
// 		Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

// 		// All initial vesting amount goes to dest
// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bG := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bD := balRes.Balance

// 		Expect(bF).To(Equal(balanceFunder), "expected funder balance to be unchanged")
// 		Expect(bG.IsZero()).To(BeTrue(), "expected all tokens to be clawed back")
// 		Expect(bD).To(Equal(balanceDest.Add(vestingAmtTotal[0])), "expected all tokens to be clawed back to the destination account")
// 	})

// 	It("should claw back any unvested amount after cliff before unlocking", func() {
// 		// Surpass cliff but not lockup duration
// 		cliffDuration := time.Duration(cliffLength)
// 		s.network.NextBlockAfter(cliffDuration * time.Second)

// 		// Check that all tokens are locked and some, but not all tokens are vested
// 		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 		free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
// 		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
// 		expVestedAmount := amt.Mul(math.NewInt(cliff))
// 		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
// 		unvested := vestingAmtTotal.Sub(vested...)

// 		Expect(expVested).To(Equal(vested))
// 		s.Require().True(expVestedAmount.GT(math.NewInt(0)))
// 		s.Require().True(free.IsZero())
// 		Expect(vesting).To(Equal(vestingAmtTotal))

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceDest := balRes.Balance

// 		// Perform clawback
// 		msg := types.NewMsgClawback(funder, vestingAddr, dest)
// 		ctx := s.network.GetContext()
// 		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 		Expect(err).To(BeNil())
// 		Expect(res.Coins).To(Equal(unvested), "expected unvested coins to be clawed back")

// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bG := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bD := balRes.Balance

// 		expClawback := clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())

// 		// Any unvested amount is clawed back
// 		Expect(balanceFunder).To(Equal(bF))
// 		Expect(balanceGrantee.Sub(expClawback[0]).Amount).To(Equal(bG.Amount))
// 		Expect(balanceDest.Add(expClawback[0]).Amount).To(Equal(bD.Amount))
// 	})

// 	It("should claw back any unvested amount after cliff and unlocking", func() {
// 		// Surpass lockup duration
// 		// A strict `if t < clawbackTime` comparison is used in ComputeClawback
// 		// so, we increment the duration with 1 for the free token calculation to match
// 		lockupDuration := time.Duration(lockupLength + 1)
// 		s.network.NextBlockAfter(lockupDuration * time.Second)

// 		// Check if some, but not all tokens are vested and unlocked
// 		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 		free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
// 		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
// 		expVestedAmount := amt.Mul(math.NewInt(lockup))
// 		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
// 		unvested := vestingAmtTotal.Sub(vested...)

// 		Expect(free).To(Equal(vested))
// 		Expect(expVested).To(Equal(vested))
// 		s.Require().True(expVestedAmount.GT(math.NewInt(0)))
// 		Expect(vesting).To(Equal(unvested))

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceDest := balRes.Balance

// 		// Perform clawback
// 		msg := types.NewMsgClawback(funder, vestingAddr, dest)
// 		ctx := s.network.GetContext()
// 		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 		Expect(err).To(BeNil())
// 		Expect(res.Coins).To(Equal(unvested), "expected only coins to be clawed back")

// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bG := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bD := balRes.Balance

// 		// Any unvested amount is clawed back
// 		Expect(balanceFunder).To(Equal(bF))
// 		Expect(balanceGrantee.Sub(vesting[0]).Amount).To(Equal(bG.Amount))
// 		Expect(balanceDest.Add(vesting[0]).Amount).To(Equal(bD.Amount))
// 	})

// 	It("should not claw back any amount after vesting periods end", func() {
// 		// Surpass vesting periods
// 		vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
// 		s.network.NextBlockAfter(vestingDuration * time.Second)

// 		// Check if some, but not all tokens are vested and unlocked
// 		vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 		unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 		free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
// 		vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())

// 		expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))
// 		unvested := vestingAmtTotal.Sub(vested...)

// 		Expect(free).To(Equal(vested))
// 		Expect(expVested).To(Equal(vested))
// 		Expect(expVested).To(Equal(vestingAmtTotal))
// 		Expect(unlocked).To(Equal(vestingAmtTotal))
// 		Expect(vesting).To(Equal(unvested))
// 		s.Require().True(vesting.IsZero())

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceDest := balRes.Balance

// 		// Perform clawback
// 		msg := types.NewMsgClawback(funder, vestingAddr, dest)
// 		ctx := s.network.GetContext()
// 		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 		Expect(err).To(BeNil(), "expected no error during clawback")
// 		Expect(res).ToNot(BeNil(), "expected response not to be nil")
// 		Expect(res.Coins).To(BeEmpty(), "expected nothing to be clawed back")

// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bG := balRes.Balance
// 		balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bD := balRes.Balance

// 		// No amount is clawed back
// 		Expect(balanceFunder).To(Equal(bF))
// 		Expect(balanceGrantee).To(Equal(bG))
// 		Expect(balanceDest).To(Equal(bD))
// 	})

// 	Context("while there is an active governance proposal for the vesting account", func() {
// 		var clawbackProposalID uint64

// 		BeforeEach(func() {
// 			// submit a different proposal to simulate having multiple proposals of different types
// 			// on chain.
// 			msgSubmitProposal, err := govv1beta1.NewMsgSubmitProposal(
// 				&erc20types.RegisterERC20Proposal{
// 					Title:          "test gov upgrade",
// 					Description:    "this is an example of a governance proposal to upgrade the evmos app",
// 					Erc20Addresses: []string{},
// 				},
// 				sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(1e9))),
// 				s.keyring.GetAddr(0).Bytes(),
// 			)
// 			Expect(err).ToNot(HaveOccurred(), "expected no error creating the proposal submission message")

// 			_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, s.keyring.GetPrivKey(0), nil, msgSubmitProposal)
// 			Expect(err).ToNot(HaveOccurred(), "expected no error during proposal submission")

// 			// submit clawback proposal
// 			govClawbackProposal := &types.ClawbackProposal{
// 				Title:              "test gov clawback",
// 				Description:        "this is an example of a governance proposal to clawback vesting coins",
// 				Address:            vestingAddr.String(),
// 				DestinationAddress: funder.String(),
// 			}

// 			deposit := sdk.Coins{sdk.Coin{Denom: stakeDenom, Amount: math.NewInt(1)}}

// 			// Create the message to submit the proposal
// 			msgSubmit, err := govv1beta1.NewMsgSubmitProposal(
// 				govClawbackProposal, deposit, s.keyring.GetAddr(0).Bytes(),
// 			)
// 			Expect(err).ToNot(HaveOccurred(), "expected no error creating the proposal submission message")
// 			// deliver the proposal
// 			_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, s.keyring.GetPrivKey(0), nil, msgSubmit)
// 			Expect(err).ToNot(HaveOccurred(), "expected no error during proposal submission")

// 			Expect(s.network.NextBlock()).To(BeNil())

// 			// Check if the proposal was submitted
// 			res, err := s.network.GetGovClient().Proposals(s.network.GetContext(), &govv1.QueryProposalsRequest{})
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(res).ToNot(BeNil())

// 			Expect(len(res.Proposals)).To(Equal(2), "expected two proposals to be found")
// 			proposal := res.Proposals[len(res.Proposals)-1]
// 			clawbackProposalID = proposal.Id
// 			Expect(proposal.GetTitle()).To(Equal("test gov clawback"), "expected different proposal title")
// 			Expect(proposal.Status).To(Equal(govv1.StatusDepositPeriod), "expected proposal to be in deposit period")
// 		})

// 		Context("with deposit made", func() {
// 			BeforeEach(func() {
// 				params, err := s.network.App.GovKeeper.Params.Get(s.network.GetContext())
// 				Expect(err).ToNot(HaveOccurred())
// 				depositAmount := params.MinDeposit[0].Amount.Sub(math.NewInt(1))
// 				deposit := sdk.Coins{sdk.Coin{Denom: params.MinDeposit[0].Denom, Amount: depositAmount}}

// 				// Deliver the deposit
// 				msgDeposit := govv1beta1.NewMsgDeposit(s.keyring.GetAddr(0).Bytes(), clawbackProposalID, deposit)
// 				_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, s.keyring.GetPrivKey(0), nil, msgDeposit)
// 				Expect(err).ToNot(HaveOccurred(), "expected no error during proposal deposit")

// 				Expect(s.network.NextBlock()).To(BeNil())

// 				// Check the proposal is in voting period
// 				proposal, found := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), clawbackProposalID)
// 				Expect(found).To(BeTrue(), "expected proposal to be found")
// 				Expect(proposal.Status).To(Equal(govv1.StatusVotingPeriod), "expected proposal to be in voting period")

// 				// Check the store entry was set correctly
// 				hasActivePropposal := s.network.App.VestingKeeper.HasActiveClawbackProposal(s.network.GetContext(), vestingAddr)
// 				Expect(hasActivePropposal).To(BeTrue(), "expected an active clawback proposal for the vesting account")
// 			})

// 			It("should not allow clawback", func() {
// 				// Try to clawback tokens
// 				msgClawback := types.NewMsgClawback(funder, vestingAddr, dest)
// 				_, err = s.network.App.VestingKeeper.Clawback(s.network.GetContext(), msgClawback)
// 				Expect(err).To(HaveOccurred(), "expected error during clawback while there is an active governance proposal")
// 				Expect(err.Error()).To(ContainSubstring("clawback is disabled while there is an active clawback proposal"))

// 				// Check that the clawback was not performed
// 				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 				Expect(acc).ToNot(BeNil(), "expected account to exist")
// 				_, isClawback := acc.(*types.ClawbackVestingAccount)
// 				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

// 				balances, err := s.network.App.VestingKeeper.Balances(s.network.GetContext(), &types.QueryBalancesRequest{
// 					Address: vestingAddr.String(),
// 				})
// 				Expect(err).ToNot(HaveOccurred(), "expected no error during balances query")
// 				Expect(balances.Unvested).To(Equal(vestingAmtTotal), "expected no tokens to be clawed back")

// 				// Delegate some funds to the suite validators in order to vote on proposal with enough voting power
// 				// using only the suite private key
// 				priv, ok := s.keyring.GetPrivKey(0).(*ethsecp256k1.PrivKey)
// 				Expect(ok).To(BeTrue(), "expected private key to be of type ethsecp256k1.PrivKey")
// 				validators, err := s.network.App.StakingKeeper.GetBondedValidatorsByPower(s.network.GetContext())
// 				Expect(err).ToNot(HaveOccurred())
// 				err = testutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, s.keyring.GetAddr(0).Bytes(), 5e18)
// 				Expect(err).ToNot(HaveOccurred(), "expected no error during funding of account")
// 				for _, val := range validators {
// 					res, err := testutil.Delegate(s.network.GetContext(), s.network.App, priv, sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)), val)
// 					Expect(err).ToNot(HaveOccurred(), "expected no error during delegation")
// 					Expect(res.Code).To(BeZero(), "expected delegation to succeed")
// 				}

// 				// Vote on proposal
// 				res, err := testutil.Vote(s.network.GetContext(), s.network.App, priv, clawbackProposalID, govv1beta1.OptionYes)
// 				Expect(err).ToNot(HaveOccurred(), "failed to vote on proposal %d", clawbackProposalID)
// 				Expect(res.Code).To(BeZero(), "expected proposal voting to succeed")

// 				// Check that the funds are clawed back after the proposal has ended
// 				s.network.NextBlockAfter(time.Hour * 24 * 365) // one year
// 				// Commit again because EndBlocker is run with time of the previous block and gov proposals are ended in EndBlocker
// 				Expect(s.network.NextBlock()).To(BeNil())

// 				// Check that proposal has passed
// 				proposal, found := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), clawbackProposalID)
// 				Expect(found).To(BeTrue(), "expected proposal to exist")
// 				Expect(proposal.Status).ToNot(Equal(govv1.StatusVotingPeriod), "expected proposal to not be in voting period anymore")
// 				Expect(proposal.Status).To(Equal(govv1.StatusPassed), "expected proposal to have passed")

// 				// Check that the account was converted to a normal account
// 				acc = s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 				Expect(acc).ToNot(BeNil(), "expected account to exist")
// 				_, isClawback = acc.(*types.ClawbackVestingAccount)
// 				Expect(isClawback).To(BeFalse(), "expected account to be a normal account")

// 				hasActiveProposal := s.network.App.VestingKeeper.HasActiveClawbackProposal(s.network.GetContext(), vestingAddr)
// 				Expect(hasActiveProposal).To(BeFalse(), "expected no active clawback proposal")
// 			})

// 			It("should not allow changing the vesting funder", func() {
// 				msgUpdateFunder := types.NewMsgUpdateVestingFunder(funder, dest, vestingAddr)
// 				_, err = s.network.App.VestingKeeper.UpdateVestingFunder(s.network.GetContext(), msgUpdateFunder)
// 				Expect(err).To(HaveOccurred(), "expected error during update funder while there is an active governance proposal")
// 				Expect(err.Error()).To(ContainSubstring("cannot update funder while there is an active clawback proposal"))

// 				// Check that the funder was not updated
// 				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 				Expect(acc).ToNot(BeNil(), "expected account to exist")
// 				clawbackAcc, isClawback := acc.(*types.ClawbackVestingAccount)
// 				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")
// 				Expect(clawbackAcc.FunderAddress).To(Equal(funder.String()), "expected funder to be unchanged")
// 			})
// 		})

// 		Context("without deposit made", func() {
// 			It("allows clawback and changing the funder before the deposit period ends", func() {
// 				newFunder, newPriv := utiltx.NewAccAddressAndKey()

// 				// fund accounts
// 				err = testutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, newFunder, 5e18)
// 				Expect(err).ToNot(HaveOccurred(), "failed to fund target account")
// 				err = testutil.FundAccountWithBaseDenom(s.network.GetContext(), s.network.App.BankKeeper, funder, 5e18)
// 				Expect(err).ToNot(HaveOccurred(), "failed to fund target account")

// 				msgUpdateFunder := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
// 				_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, funderPriv, nil, msgUpdateFunder)
// 				Expect(err).ToNot(HaveOccurred(), "expected no error during update funder while there is an active governance proposal")

// 				// Check that the funder was updated
// 				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 				Expect(acc).ToNot(BeNil(), "expected account to exist")
// 				_, isClawback := acc.(*types.ClawbackVestingAccount)
// 				Expect(isClawback).To(BeTrue(), "expected account to be clawback vesting account")

// 				// Claw back tokens
// 				msgClawback := types.NewMsgClawback(newFunder, vestingAddr, funder)
// 				_, err = testutil.DeliverTx(s.network.GetContext(), s.network.App, newPriv, nil, msgClawback)
// 				Expect(err).ToNot(HaveOccurred(), "expected no error during clawback while there is no deposit made")

// 				// Check account is converted to a normal account
// 				acc = s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 				Expect(acc).ToNot(BeNil(), "expected account to exist")
// 				_, isClawback = acc.(*types.ClawbackVestingAccount)
// 				Expect(isClawback).To(BeFalse(), "expected account to be a normal account")
// 			})

// 			It("should remove the store entry after the deposit period ends", func() {
// 				s.network.NextBlockAfter(time.Hour * 24 * 365) // one year
// 				// Commit again because EndBlocker is run with time of the previous block and gov proposals are ended in EndBlocker
// 				Expect(s.network.NextBlock()).To(BeNil())

// 				// Check that the proposal has ended -- since deposit failed it's removed from the store
// 				_, found := s.network.App.GovKeeper.Proposals.Get(s.network.GetContext(), clawbackProposalID)
// 				Expect(found).To(BeFalse(), "expected proposal not to be found")

// 				// Check that the store entry was removed
// 				hasActiveProposal := s.network.App.VestingKeeper.HasActiveClawbackProposal(s.network.GetContext(), vestingAddr)
// 				Expect(hasActiveProposal).To(BeFalse(),
// 					"expected no active clawback proposal for address %q",
// 					vestingAddr.String(),
// 				)
// 			})
// 		})
// 	})

// 	It("should update vesting funder and claw back unvested amount before cliff", func() {
// 		ctx := s.network.GetContext()
// 		newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceNewFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee := balRes.Balance

// 		// Update clawback vesting account funder
// 		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
// 		_, err = s.network.App.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
// 		Expect(err).To(BeNil())

// 		// Perform clawback before cliff - funds should go to new funder (no dest address defined)
// 		msg := types.NewMsgClawback(newFunder, vestingAddr, sdk.AccAddress([]byte{}))
// 		res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 		Expect(err).To(BeNil())
// 		Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

// 		// All initial vesting amount goes to funder
// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bNewF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bG := balRes.Balance

// 		// Original funder balance should not change
// 		Expect(bF).To(Equal(balanceFunder))
// 		// New funder should get the vested tokens
// 		Expect(balanceNewFunder.Add(vestingAmtTotal[0]).Amount).To(Equal(bNewF.Amount))
// 		Expect(balanceGrantee.Sub(vestingAmtTotal[0]).Amount).To(Equal(bG.Amount))
// 	})

// 	It("should update vesting funder and first funder cannot claw back unvested before cliff", func() {
// 		ctx := s.network.GetContext()
// 		newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

// 		balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceNewFunder := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		balanceGrantee := balRes.Balance

// 		// Update clawback vesting account funder
// 		updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
// 		_, err = s.network.App.VestingKeeper.UpdateVestingFunder(ctx, updateFunderMsg)
// 		Expect(err).To(BeNil())

// 		// Original funder tries to perform clawback before cliff - is not the current funder
// 		msg := types.NewMsgClawback(funder, vestingAddr, sdk.AccAddress([]byte{}))
// 		_, err = s.network.App.VestingKeeper.Clawback(ctx, msg)
// 		s.Require().Error(err)

// 		// All balances should remain the same
// 		balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bNewF := balRes.Balance
// 		balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 		Expect(err).To(BeNil())
// 		bG := balRes.Balance

// 		Expect(bF).To(Equal(balanceFunder))
// 		Expect(balanceNewFunder).To(Equal(bNewF))
// 		Expect(balanceGrantee).To(Equal(bG))
// 	})

// 	Context("governance clawback to community pool", func() {
// 		It("should claw back unvested amount before cliff", func() {
// 			ctx := s.network.GetContext()

// 			// initial balances
// 			balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceFunder := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceGrantee := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceDest := balRes.Balance

// 			poolRes, err := s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())
// 			balanceCommPool := poolRes.Pool[0]

// 			// Perform clawback before cliff
// 			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
// 			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 			Expect(err).To(BeNil())
// 			Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

// 			// All initial vesting amount goes to community pool instead of dest
// 			balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bF := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bG := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bD := balRes.Balance

// 			poolRes, err = s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())
// 			bCP := poolRes.Pool[0]

// 			Expect(bF).To(Equal(balanceFunder))
// 			Expect(balanceGrantee.Sub(vestingAmtTotal[0]).Amount).To(Equal(bG.Amount))
// 			// destination address should remain unchanged
// 			Expect(balanceDest.Amount).To(Equal(bD.Amount))
// 			// vesting amount should go to community pool
// 			Expect(balanceCommPool.Amount.Add(math.LegacyNewDec(vestingAmtTotal[0].Amount.Int64()))).To(Equal(bCP.Amount))
// 			Expect(stakeDenom).To(Equal(bCP.Denom))
// 		})

// 		It("should claw back any unvested amount after cliff before unlocking", func() {
// 			// Surpass cliff but not lockup duration
// 			cliffDuration := time.Duration(cliffLength)
// 			s.network.NextBlockAfter(cliffDuration * time.Second)

// 			// Check that all tokens are locked and some, but not all tokens are vested
// 			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 			unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 			free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
// 			vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
// 			expVestedAmount := amt.Mul(math.NewInt(cliff))
// 			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
// 			unvested := vestingAmtTotal.Sub(vested...)

// 			Expect(expVested).To(Equal(vested))
// 			s.Require().True(expVestedAmount.GT(math.NewInt(0)))
// 			s.Require().True(free.IsZero())
// 			Expect(vesting).To(Equal(vestingAmtTotal))

// 			balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceFunder := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceGrantee := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceDest := balRes.Balance

// 			poolRes, err := s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())

// 			balanceCommPool := poolRes.Pool[0]

// 			testClawbackAccount := TestClawbackAccount{
// 				privKey:         nil,
// 				address:         vestingAddr,
// 				clawbackAccount: clawbackAccount,
// 			}
// 			// stake vested tokens
// 			ok, vestedCoin := vested.Find(utils.BaseDenom)
// 			Expect(ok).To(BeTrue())
// 			err = s.factory.Delegate(
// 				testClawbackAccount.privKey,
// 				s.network.GetValidators()[0].OperatorAddress,
// 				vestedCoin,
// 			)
// 			Expect(err).To(BeNil())
// 			Expect(s.network.NextBlock()).To(BeNil())

// 			// Perform clawback
// 			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
// 			ctx := s.network.GetContext()
// 			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 			Expect(err).To(BeNil())
// 			Expect(res.Coins).To(Equal(unvested), "expected unvested coins to be clawed back")

// 			balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bF := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bG := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bD := balRes.Balance

// 			poolRes, err = s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())

// 			bCP := poolRes.Pool[0]

// 			expClawback := clawbackAccount.GetUnvestedOnly(s.network.GetContext().BlockTime())

// 			// Any unvested amount is clawed back to community pool
// 			Expect(balanceFunder).To(Equal(bF))
// 			Expect(balanceGrantee.Sub(expClawback[0]).Amount).To(Equal(bG.Amount))
// 			Expect(balanceDest.Amount).To(Equal(bD.Amount))
// 			// vesting amount should go to community pool
// 			Expect(balanceCommPool.Amount.Add(math.LegacyNewDec(expClawback[0].Amount.Int64()))).To(Equal(bCP.Amount))
// 			Expect(stakeDenom).To(Equal(bCP.Denom))
// 		})

// 		It("should claw back any unvested amount after cliff and unlocking", func() {
// 			// Surpass lockup duration
// 			// A strict `if t < clawbackTime` comparison is used in ComputeClawback
// 			// so, we increment the duration with 1 for the free token calculation to match
// 			lockupDuration := time.Duration(lockupLength + 1)
// 			s.network.NextBlockAfter(lockupDuration * time.Second)

// 			// Check if some, but not all tokens are vested and unlocked
// 			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 			unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 			free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
// 			vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())
// 			expVestedAmount := amt.Mul(math.NewInt(lockup))
// 			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, expVestedAmount))
// 			unvested := vestingAmtTotal.Sub(vested...)

// 			Expect(free).To(Equal(vested))
// 			Expect(expVested).To(Equal(vested))
// 			s.Require().True(expVestedAmount.GT(math.NewInt(0)))
// 			Expect(vesting).To(Equal(unvested))

// 			balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceFunder := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceGrantee := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceDest := balRes.Balance

// 			testClawbackAccount := TestClawbackAccount{
// 				privKey:         nil,
// 				address:         vestingAddr,
// 				clawbackAccount: clawbackAccount,
// 			}
// 			// stake vested tokens
// 			ok, vestedCoin := vested.Find(utils.BaseDenom)
// 			Expect(ok).To(BeTrue())
// 			err = s.factory.Delegate(
// 				testClawbackAccount.privKey,
// 				s.network.GetValidators()[0].OperatorAddress,
// 				vestedCoin,
// 			)
// 			Expect(err).To(BeNil())

// 			// Perform clawback
// 			msg := types.NewMsgClawback(funder, vestingAddr, dest)
// 			ctx := s.network.GetContext()
// 			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 			Expect(err).To(BeNil())
// 			Expect(res.Coins).To(Equal(unvested), "expected only unvested coins to be clawed back")

// 			balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bF := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bG := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bD := balRes.Balance

// 			// Any unvested amount is clawed back
// 			Expect(balanceFunder).To(Equal(bF))
// 			Expect(balanceGrantee.Sub(vesting[0]).Amount).To(Equal(bG.Amount))
// 			Expect(balanceDest.Add(vesting[0]).Amount).To(Equal(bD.Amount))
// 		})

// 		It("should not claw back any amount after vesting periods end", func() {
// 			// Surpass vesting periods
// 			vestingDuration := time.Duration(periodsTotal*vestingLength + 1)
// 			s.network.NextBlockAfter(vestingDuration * time.Second)

// 			// Check if some, but not all tokens are vested and unlocked
// 			vested = clawbackAccount.GetVestedOnly(s.network.GetContext().BlockTime())
// 			unlocked = clawbackAccount.GetUnlockedOnly(s.network.GetContext().BlockTime())
// 			free = clawbackAccount.GetVestedCoins(s.network.GetContext().BlockTime())
// 			vesting = clawbackAccount.GetVestingCoins(s.network.GetContext().BlockTime())

// 			expVested := sdk.NewCoins(sdk.NewCoin(stakeDenom, amt.Mul(math.NewInt(periodsTotal))))
// 			unvested := vestingAmtTotal.Sub(vested...)

// 			Expect(free).To(Equal(vested))
// 			Expect(expVested).To(Equal(vested))
// 			Expect(expVested).To(Equal(vestingAmtTotal))
// 			Expect(unlocked).To(Equal(vestingAmtTotal))
// 			Expect(vesting).To(Equal(unvested))
// 			Expect(vesting.IsZero()).To(BeTrue())

// 			balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceFunder := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceGrantee := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceDest := balRes.Balance

// 			poolRes, err := s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())

// 			balanceCommPool := poolRes.Pool[0]

// 			testClawbackAccount := TestClawbackAccount{
// 				privKey:         nil,
// 				address:         vestingAddr,
// 				clawbackAccount: clawbackAccount,
// 			}
// 			// stake vested tokens
// 			ok, vestedCoin := vested.Find(stakeDenom)
// 			Expect(ok).To(BeTrue())
// 			err = s.factory.Delegate(
// 				testClawbackAccount.privKey,
// 				s.network.GetValidators()[0].OperatorAddress,
// 				vestedCoin,
// 			)
// 			Expect(err).To(BeNil())

// 			// Perform clawback
// 			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
// 			ctx := s.network.GetContext()
// 			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 			Expect(err).To(BeNil(), "expected no error during clawback")
// 			Expect(res.Coins).To(BeEmpty(), "expected nothing to be clawed back after end of vesting schedules")

// 			balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bF := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bG := balRes.Balance
// 			balRes, err = s.handler.GetBalance(dest, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bD := balRes.Balance

// 			poolRes, err = s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())

// 			bCP := poolRes.Pool[0]

// 			// No amount is clawed back
// 			Expect(balanceFunder).To(Equal(bF))
// 			Expect(balanceGrantee).To(Equal(bG))
// 			Expect(balanceDest).To(Equal(bD))
// 			Expect(balanceCommPool.Amount).To(Equal(bCP.Amount))
// 		})

// 		It("should update vesting funder and claw back unvested amount before cliff", func() {
// 			ctx := s.network.GetContext()
// 			newFunder := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

// 			balRes, err := s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceFunder := balRes.Balance
// 			balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceNewFunder := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			balanceGrantee := balRes.Balance

// 			poolRes, err := s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())

// 			balanceCommPool := poolRes.Pool[0]

// 			// Update clawback vesting account funder
// 			updateFunderMsg := types.NewMsgUpdateVestingFunder(funder, newFunder, vestingAddr)
// 			_, err = s.factory.ExecuteCosmosTx(funderPriv, factory.CosmosTxArgs{Msgs: []sdk.Msg{updateFunderMsg}})
// 			Expect(err).To(BeNil())
// 			Expect(s.network.NextBlock()).To(BeNil())

// 			// Perform clawback before cliff - funds should go to new funder (no dest address defined)
// 			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, nil)
// 			res, err := s.network.App.VestingKeeper.Clawback(ctx, msg)
// 			Expect(err).To(BeNil())
// 			Expect(res.Coins).To(Equal(vestingAmtTotal), "expected different coins to be clawed back")

// 			// All initial vesting amount goes to funder
// 			balRes, err = s.handler.GetBalance(funder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bF := balRes.Balance
// 			balRes, err = s.handler.GetBalance(newFunder, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bNewF := balRes.Balance
// 			balRes, err = s.handler.GetBalance(vestingAddr, stakeDenom)
// 			Expect(err).To(BeNil())
// 			bG := balRes.Balance

// 			poolRes, err = s.handler.GetCommunityPool()
// 			Expect(err).To(BeNil())

// 			bCP := poolRes.Pool[0]

// 			// Original funder balance should not change
// 			Expect(bF).To(Equal(balanceFunder))
// 			// New funder should not get the vested tokens
// 			Expect(balanceNewFunder.Amount).To(Equal(bNewF.Amount))
// 			Expect(balanceGrantee.Sub(vestingAmtTotal[0]).Amount).To(Equal(bG.Amount))
// 			// vesting amount should go to community pool
// 			Expect(balanceCommPool.Amount.Add(math.LegacyNewDec(vestingAmtTotal[0].Amount.Int64()))).To(Equal(bCP.Amount))
// 		})

// 		It("should not claw back when governance clawback is disabled", func() {
// 			// disable governance clawback
// 			s.network.App.VestingKeeper.SetGovClawbackDisabled(s.network.GetContext(), vestingAddr)

// 			// Perform clawback before cliff
// 			msg := types.NewMsgClawback(authtypes.NewModuleAddress(govtypes.ModuleName), vestingAddr, dest)
// 			_, err := s.network.App.VestingKeeper.Clawback(s.network.GetContext(), msg)
// 			Expect(err).To(HaveOccurred(), "expected error")
// 			Expect(err.Error()).To(ContainSubstring("%s: account does not have governance clawback enabled", vestingAddr.String()))
// 		})
// 	})
// })

// // Testing that smart contracts cannot be converted to clawback vesting accounts
// //
// // NOTE: For smart contracts, it is not possible to directly call keeper methods
// // or send SDK transactions. They go exclusively through the EVM, which is tested
// // in the precompiles package.
// // The test here is just confirming the expected behavior on the module level.
// var _ = Describe("Clawback Vesting Account - Smart contract", func() {
// 	var (
// 		s            *KeeperTestSuite
// 		contractAddr common.Address
// 		contract     evmtypes.CompiledContract
// 		err          error
// 	)

// 	BeforeEach(func() {
// 		s = new(KeeperTestSuite)
// 		s.SetupTest()
// 		contract = contracts.ERC20MinterBurnerDecimalsContract
// 		contractAddr, err = testutil.DeployContract(
// 			s.network.GetContext(),
// 			s.network.App,
// 			s.keyring.GetPrivKey(0),
// 			s.network.GetEvmClient(),
// 			contract,
// 			"Test", "TTT", uint8(18),
// 		)
// 		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")
// 	})

// 	It("should not convert a smart contract to a clawback vesting account", func() {
// 		msgCreate := types.NewMsgCreateClawbackVestingAccount(
// 			s.keyring.GetAccAddr(0),
// 			contractAddr.Bytes(),
// 			false,
// 		)
// 		_, err := s.network.App.VestingKeeper.CreateClawbackVestingAccount(s.network.GetContext(), msgCreate)
// 		Expect(err).To(HaveOccurred(), "expected error")
// 		Expect(err.Error()).To(ContainSubstring(
// 			fmt.Sprintf(
// 				"account %s is a contract account and cannot be converted in a clawback vesting account",
// 				sdk.AccAddress(contractAddr.Bytes()).String()),
// 		))

// 		// Check that the account was not converted
// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), contractAddr.Bytes())
// 		Expect(acc).ToNot(BeNil(), "smart contract should be found")
// 		_, ok := acc.(*types.ClawbackVestingAccount)
// 		Expect(ok).To(BeFalse(), "account should not be a clawback vesting account")

// 		// Check that the contract code was not deleted
// 		//
// 		// NOTE: When it was possible to create clawback vesting accounts for smart contracts,
// 		// the contract code was deleted from the EVM state. This checks that this is not the case.
// 		res, err := s.network.App.EvmKeeper.Code(s.network.GetContext(), &evmtypes.QueryCodeRequest{Address: contractAddr.String()})
// 		Expect(err).ToNot(HaveOccurred(), "failed to query contract code")
// 		Expect(res.Code).ToNot(BeEmpty(), "contract code should not be empty")
// 	})
// })

// // Trying to replicate the faulty behavior in MsgCreateClawbackVestingAccount,
// // that was disclosed as a potential attack vector in relation to the Barberry
// // security patch.
// //
// // It was possible to fund a clawback vesting account with negative amounts.
// // Avoiding this requires an additional validation of the amount in the
// // MsgFundVestingAccount's ValidateBasic method.
// var _ = Describe("Clawback Vesting Account - Barberry bug", func() {
// 	var (
// 		s *KeeperTestSuite
// 		// coinsNoNegAmount is a Coins struct with a positive and a negative amount of the same
// 		// denomination.
// 		coinsNoNegAmount = sdk.Coins{
// 			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)},
// 		}
// 		// coinsWithNegAmount is a Coins struct with a positive and a negative amount of the same
// 		// denomination.
// 		coinsWithNegAmount = sdk.Coins{
// 			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)},
// 			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(-1e18)},
// 		}
// 		// coinsWithZeroAmount is a Coins struct with a positive and a zero amount of the same
// 		// denomination.
// 		coinsWithZeroAmount = sdk.Coins{
// 			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)},
// 			sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(0)},
// 		}
// 		// emptyCoins is an Coins struct
// 		emptyCoins = sdk.Coins{}
// 		// funder and funderPriv are the address and private key of the account funding the vesting account
// 		funder, funderPriv = utiltx.NewAccAddressAndKey()
// 		// gasPrice is the gas price to be used in the transactions executed by the vesting account so that
// 		// the transaction fees can be deducted from the expected account balance
// 		gasPrice = math.NewInt(1e9)
// 		// vestingAddr and vestingPriv are the address and private key of the vesting account to be created
// 		vestingAddr, vestingPriv = utiltx.NewAccAddressAndKey()
// 		// vestingLength is a period of time in seconds to be used for the creation of the vesting
// 		// account.
// 		vestingLength = int64(60 * 60 * 24 * 30) // 30 days in seconds

// 		// txCost is the cost of a transaction to be deducted from the expected account balance
// 		txCost int64
// 	)

// 	BeforeEach(func() {
// 		s = new(KeeperTestSuite)
// 		s.SetupTest()

// 		// Initialize the account at the vesting address and the funder accounts by funding them
// 		fundedCoins := sdk.Coins{{Denom: utils.BaseDenom, Amount: math.NewInt(2e18)}} // fund more than what is sent to the vesting account for transaction fees
// 		err = testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, vestingAddr, fundedCoins)
// 		Expect(err).ToNot(HaveOccurred(), "failed to fund account")
// 		err = testutil.FundAccount(s.network.GetContext(), s.network.App.BankKeeper, funder, fundedCoins)
// 		Expect(err).ToNot(HaveOccurred(), "failed to fund account")

// 		// Create a clawback vesting account
// 		msgCreate := types.NewMsgCreateClawbackVestingAccount(
// 			funder,
// 			vestingAddr,
// 			false,
// 		)

// 		res, err := testutil.DeliverTx(s.network.GetContext(), s.network.App, vestingPriv, &gasPrice, msgCreate)
// 		Expect(err).ToNot(HaveOccurred(), "failed to create clawback vesting account")
// 		txCost = gasPrice.Int64() * res.GasWanted

// 		// Check clawback acccount was created
// 		acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 		Expect(acc).ToNot(BeNil(), "clawback vesting account not created")
// 		_, ok := acc.(*types.ClawbackVestingAccount)
// 		Expect(ok).To(BeTrue(), "account is not a clawback vesting account")
// 	})

// 	Context("when funding a clawback vesting account", func() {
// 		testcases := []struct {
// 			name         string
// 			lockupCoins  sdk.Coins
// 			vestingCoins sdk.Coins
// 			expError     bool
// 			errContains  string
// 		}{
// 			{
// 				name:        "pass - positive amounts for the lockup period",
// 				lockupCoins: coinsNoNegAmount,
// 				expError:    false,
// 			},
// 			{
// 				name:         "pass - positive amounts for the vesting period",
// 				vestingCoins: coinsNoNegAmount,
// 				expError:     false,
// 			},
// 			{
// 				name:         "pass - positive amounts for both the lockup and vesting periods",
// 				lockupCoins:  coinsNoNegAmount,
// 				vestingCoins: coinsNoNegAmount,
// 				expError:     false,
// 			},
// 			{
// 				name:        "fail - negative amounts for the lockup period",
// 				lockupCoins: coinsWithNegAmount,
// 				expError:    true,
// 				errContains: errortypes.ErrInvalidCoins.Wrap(coinsWithNegAmount.String()).Error(),
// 			},
// 			{
// 				name:         "fail - negative amounts for the vesting period",
// 				vestingCoins: coinsWithNegAmount,
// 				expError:     true,
// 				errContains:  "invalid coins: invalid request",
// 			},
// 			{
// 				name:        "fail - zero amount for the lockup period",
// 				lockupCoins: coinsWithZeroAmount,
// 				expError:    true,
// 				errContains: errortypes.ErrInvalidCoins.Wrap(coinsWithZeroAmount.String()).Error(),
// 			},
// 			{
// 				name:         "fail - zero amount for the vesting period",
// 				vestingCoins: coinsWithZeroAmount,
// 				expError:     true,
// 				errContains:  "invalid coins: invalid request",
// 			},
// 			{
// 				name:         "fail - empty amount for both the lockup and vesting periods",
// 				lockupCoins:  emptyCoins,
// 				vestingCoins: emptyCoins,
// 				expError:     true,
// 				errContains:  "vesting and/or lockup schedules must be present",
// 			},
// 		}

// 		for _, tc := range testcases {
// 			tc := tc
// 			It(tc.name, func() {
// 				var (
// 					lockupPeriods  sdkvesting.Periods
// 					vestingPeriods sdkvesting.Periods
// 				)

// 				if !tc.lockupCoins.Empty() {
// 					lockupPeriods = sdkvesting.Periods{
// 						sdkvesting.Period{Length: vestingLength, Amount: tc.lockupCoins},
// 					}
// 				}

// 				if !tc.vestingCoins.Empty() {
// 					vestingPeriods = sdkvesting.Periods{
// 						sdkvesting.Period{Length: vestingLength, Amount: tc.vestingCoins},
// 					}
// 				}

// 				// Fund the clawback vesting account at the given address
// 				msg := types.NewMsgFundVestingAccount(
// 					funder,
// 					vestingAddr,
// 					s.network.GetContext().BlockTime(),
// 					lockupPeriods,
// 					vestingPeriods,
// 				)

// 				// Deliver transaction with message
// 				res, err := testutil.DeliverTx(s.network.GetContext(), s.network.App, funderPriv, nil, msg)

// 				// Get account at the new address
// 				acc := s.network.App.AccountKeeper.GetAccount(s.network.GetContext(), vestingAddr)
// 				vacc, _ := acc.(*types.ClawbackVestingAccount)

// 				if tc.expError {
// 					Expect(err).To(HaveOccurred(), "expected funding the vesting account to have failed")
// 					Expect(err.Error()).To(ContainSubstring(tc.errContains), "expected funding the vesting account to have failed")

// 					Expect(vacc.LockupPeriods).To(BeEmpty(), "expected clawback vesting account to not have been funded")
// 				} else {
// 					Expect(err).ToNot(HaveOccurred(), "failed to fund clawback vesting account")
// 					Expect(res.Code).To(Equal(uint32(0)), "failed to fund clawback vesting account")
// 					Expect(vacc.LockupPeriods).ToNot(BeEmpty(), "vesting account should have been funded")

// 					// Check that the vesting account has the correct balance
// 					balRes, err := s.handler.GetBalance(vestingAddr, utils.BaseDenom)
// 					Expect(err).To(BeNil())
// 					balance := balRes.Balance
// 					expBalance := int64(2e18) + int64(1e18) - txCost // fundedCoins + vestingCoins - txCost
// 					Expect(balance.Amount.Int64()).To(Equal(expBalance), "vesting account has incorrect balance")
// 				}
// 			})
// 		}
// 	})
// })
