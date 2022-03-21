package keeper_test

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/testutil"
	incentivestypes "github.com/tharsis/evmos/v2/x/incentives/types"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tharsis/ethermint/server/config"
	evm "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/v2/contracts"
	"github.com/tharsis/evmos/v2/x/claims/types"
)

// TODO
// params := types.DefaultParams()
// params.EnableClaims = false
// s.app.ClaimsKeeper.SetParams(s.ctx, params)

var _ = Describe("Check amount claimed depending on claim time", Ordered, func() {
	claimsAddr := s.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	distrAddr := s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	claimsDenom := types.DefaultClaimsDenom
	accountCount := 4

	claimValue := int64(math.Pow10(5) * 10)
	actionValue := int64(claimValue / 4)
	totalClaimsAmount := claimValue * int64(accountCount)
	initClaimsAmount := int64(math.Pow10(5) * 2)
	initEvmBalanceAmount := int64(math.Pow10(18))
	delegationValue := int64(math.Pow10(10) * 2)

	initBalance := sdk.NewCoins(
		sdk.NewCoin(stakeDenom, sdk.NewInt(delegationValue)),
		sdk.NewCoin(claimsDenom, sdk.NewInt(initClaimsAmount)),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(initEvmBalanceAmount)),
	)

	initBalanceAmount0 := int64(math.Pow10(18) * 2)
	initBalance0 := sdk.NewCoins(
		sdk.NewCoin(stakeDenom, sdk.NewInt(delegationValue)),
		sdk.NewCoin(claimsDenom, sdk.NewInt(initBalanceAmount0)),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(initBalanceAmount0)),
	)

	var (
		priv0         *ethsecp256k1.PrivKey
		privs         []*ethsecp256k1.PrivKey
		addr0         sdk.AccAddress
		claimsRecords []types.ClaimsRecord
		params        types.Params
		proposalId    uint64
		totalClaimed  int64
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.ClaimsKeeper.GetParams(s.ctx)
		params.EnableClaims = true
		params.AirdropStartTime = s.ctx.BlockTime()
		s.app.ClaimsKeeper.SetParams(s.ctx, params)

		// mint coins for claiming and send them to the claims module
		coins := sdk.NewCoins(sdk.NewCoin(claimsDenom, sdk.NewInt(totalClaimsAmount)))
		err := s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
		s.Require().NoError(err)

		// fund testing accounts and create claim records
		priv0, _ = ethsecp256k1.GenerateKey()
		addr0 = getAddr(priv0)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr0, initBalance0)

		for i := 0; i < accountCount; i++ {
			priv, _ := ethsecp256k1.GenerateKey()
			privs = append(privs, priv)
			addr := getAddr(priv)
			testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, initBalance)
			claimsRecord := types.NewClaimsRecord(sdk.NewInt(claimValue))
			s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr, claimsRecord)
			acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
			s.app.AccountKeeper.SetAccount(s.ctx, acc)
			claimsRecords = append(claimsRecords, claimsRecord)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(initClaimsAmount)))
		}

		// ensure community pool doesn't have the fund
		poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, claimsDenom)
		Expect(poolBalance.Amount.Uint64()).To(Equal(uint64(0)))

		// ensure module account has the escrow fund
		balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, claimsDenom)
		Expect(balanceClaims.Amount.Uint64()).To(Equal(uint64(totalClaimsAmount)))

		s.Commit()

		proposalId = govProposal(priv0)
	})

	Context("claiming before decay duration", func() {
		var actionV int64
		BeforeAll(func() {
			actionV = actionValue
		})

		It("can claim ActionDelegate", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			delegate(privs[0], 1)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()) - 1)))
			totalClaimed += actionV
		})

		It("can claim ActionEVM", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			sendEthToSelf(privs[0])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()))))
			totalClaimed += actionV
		})

		It("can claim ActionVote", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			govVote(privs[1], proposalId)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()))))
			totalClaimed += actionV
		})

		It("can claim ActionIBCTransfer", Pending, func() {
			addr := getAddr(privs[2])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			performIbcTransfer(privs[2])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()))))
			totalClaimed += actionV
		})
	})

	Context("claiming at 2/3 decay duration", func() {
		var actionV int64

		BeforeAll(func() {
			actionV = actionValue / 3
			duration := params.DecayStartTime().Sub(s.ctx.BlockHeader().Time)
			s.CommitAfter(duration)
			duration = params.GetDurationOfDecay() * 2 / 3

			// create another proposal to vote for
			testTime := s.ctx.BlockHeader().Time.Add(duration)
			s.CommitAfter(duration - time.Hour)
			proposalId = govProposal(priv0)
			s.CommitAfter(testTime.Sub(s.ctx.BlockHeader().Time))
		})

		It("can claim ActionDelegate", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			delegate(privs[1], 1)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()) - 1)))
			totalClaimed += actionV
		})

		It("can claim ActionEVM", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			sendEthToSelf(privs[1])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()))))
			totalClaimed += actionV

			sendEthToSelf(privs[2])
			totalClaimed += actionV
		})

		It("can claim ActionVote", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			govVote(privs[0], proposalId)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionV + int64(prebalance.Amount.Uint64()))))
			totalClaimed += actionV
		})

		It("cannot claim ActionDelegate a second time", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			delegate(privs[1], 1)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(prebalance.Amount.Uint64() - 1))
		})

		It("cannot claim ActionEVM a second time", func() {
			addr := getAddr(privs[1])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			sendEthToSelf(privs[1])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(prebalance.Amount.Uint64()))
		})

		It("cannot claim ActionVote a second time", func() {
			addr := getAddr(privs[0])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			govVote(privs[0], proposalId)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(prebalance.Amount.Uint64()))
		})

		It("did not clawback to the community pool", func() {
			// ensure community pool doesn't have the fund
			poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, claimsDenom)
			Expect(poolBalance.Amount.Uint64()).To(Equal(uint64(0)))

			// ensure module account has the escrow fund minus what was claimed
			balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, claimsDenom)
			Expect(balanceClaims.Amount.Uint64()).To(Equal(uint64(totalClaimsAmount - totalClaimed)))
		})
	})

	Context("claiming after decay duration", func() {
		BeforeAll(func() {
			duration := params.AirdropEndTime().Sub(s.ctx.BlockHeader().Time)
			s.CommitAfter(duration)
			proposalId = govProposal(priv0)

			// ensure module account has the unclaimed amount before airdrop ends
			moduleBalance := s.app.ClaimsKeeper.GetModuleAccountBalances(s.ctx)
			Expect(moduleBalance.AmountOf(claimsDenom).Uint64()).To(Equal(uint64(totalClaimsAmount - totalClaimed)))

			// ensure community pool has 0 funds before airdrop ends
			poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, claimsDenom)
			Expect(poolBalance.Amount.Uint64()).To(Equal(uint64(0)))

			s.Commit()
		})

		It("cannot claim after decay duration with partial claim", func() {
			addr := getAddr(privs[2])
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			delegate(privs[2], 1)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(prebalance.Amount.Uint64() - 1))
		})

		It("dust amount removed from accounts with 0 claimed", func() {
			addr := getAddr(privs[3])
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(0)))

		})

		It("final balances", func() {
			addr := getAddr(privs[0])
			finalBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			claimed := 2*actionValue + actionValue/3
			Expect(finalBalance.Amount.Uint64()).To(Equal(uint64(initClaimsAmount + claimed - 1)))

			addr = getAddr(privs[1])
			finalBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			claimed = actionValue + actionValue*2/3
			Expect(finalBalance.Amount.Uint64()).To(Equal(uint64(initClaimsAmount + claimed - 2)))

			addr = getAddr(privs[2])
			finalBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			claimed = actionValue * 1 / 3
			Expect(finalBalance.Amount.Uint64()).To(Equal(uint64(initClaimsAmount + claimed - 1)))

			addr = getAddr(privs[3])
			finalBalance = s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(finalBalance.Amount.Uint64()).To(Equal(uint64(0)))
		})

		It("can clawback unclaimed", func() {
			// ensure module account is empty
			moduleBalance := s.app.ClaimsKeeper.GetModuleAccountBalances(s.ctx)
			Expect(moduleBalance.AmountOf(claimsDenom).Uint64()).To(Equal(uint64(0)))

			// The unclaimed amount goes to the community pool
			poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, claimsDenom)
			expectedBalance := totalClaimsAmount - totalClaimed
			Expect(poolBalance.Amount.Uint64()).To(Equal(uint64(expectedBalance)))
		})
	})
})

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return sdk.AccAddress(priv.PubKey().Address().Bytes())
}

func delegate(priv *ethsecp256k1.PrivKey, amount int64) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	val, err := sdk.ValAddressFromBech32(s.validator.OperatorAddress)
	s.Require().NoError(err)

	fee := sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(amount))
	delegateMsg := stakingtypes.NewMsgDelegate(accountAddress, val, fee)
	deliverTx(priv, delegateMsg)
}

func govProposal(priv *ethsecp256k1.PrivKey) uint64 {
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	contractAddress := deployContract(priv)
	content := incentivestypes.NewRegisterIncentiveProposal(
		"test",
		"description",
		contractAddress.String(),
		sdk.DecCoins{sdk.NewDecCoinFromDec(evmtypes.DefaultEVMDenom, sdk.NewDecWithPrec(5, 2))},
		1000,
	)

	deposit := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.NewInt(100000000)))
	msg, err := govtypes.NewMsgSubmitProposal(content, deposit, accountAddress)
	s.Require().NoError(err)

	res := deliverTx(priv, msg)
	submitEvent := res.GetEvents()[4]
	Expect(submitEvent.Type).To(Equal("submit_proposal"))
	Expect(string(submitEvent.Attributes[0].Key)).To(Equal("proposal_id"))

	proposalId, err := strconv.ParseUint(string(submitEvent.Attributes[0].Value), 10, 64)
	s.Require().NoError(err)

	return proposalId
}

func govVote(priv *ethsecp256k1.PrivKey, proposalID uint64) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	voteMsg := govtypes.NewMsgVote(accountAddress, proposalID, 2)
	deliverTx(priv, voteMsg)
}

func sendEthToSelf(priv *ethsecp256k1.PrivKey) {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &from, nil, 100000, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()
	performEthTx(priv, msgEthereumTx)
}

func deployContract(priv *ethsecp256k1.PrivKey) common.Address {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "Test", "TTT", uint8(18))
	s.Require().NoError(err)

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &s.address,
		Data: (*hexutil.Bytes)(&data),
	})
	s.Require().NoError(err)

	ctx := sdk.WrapSDKContext(s.ctx)
	res, err := s.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(config.DefaultGasCap),
	})
	s.Require().NoError(err)

	msgEthereumTx := evmtypes.NewTxContract(chainID, nonce, nil, res.Gas, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	performEthTx(priv, msgEthereumTx)

	s.CommitAfter(time.Minute)

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	s.Require().NoError(err)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	s.Require().True(ok)
	builder.SetExtensionOptions(option)

	err = msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	err = txBuilder.SetMsgs(msgEthereumTx)
	s.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msgEthereumTx.Data)
	s.Require().NoError(err)

	fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewIntFromBigInt(txData.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msgEthereumTx.GetGas())

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	Expect(res.IsOK()).To(Equal(true))
}

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.ctx, accountAddress)
	s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	s.Require().NoError(err)

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	Expect(res.IsOK()).To(Equal(true))
	return res
}

func performIbcTransfer(priv *ethsecp256k1.PrivKey) {

}
