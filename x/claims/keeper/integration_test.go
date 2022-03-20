package keeper_test

import (
	"encoding/json"
	"fmt"
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

	claimValue := int64(math.Pow10(5) * 10)
	actionValue := int64(claimValue / 4)
	claimsAmount := claimValue * 10
	initBalanceAmount := int64(math.Pow10(5) * 2)
	initEvmBalanceAmount := int64(math.Pow10(18))
	delegationValue := int64(math.Pow10(10) * 2)

	initBalance := sdk.NewCoins(
		sdk.NewCoin(stakeDenom, sdk.NewInt(delegationValue)),
		sdk.NewCoin(claimsDenom, sdk.NewInt(initBalanceAmount)),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(initEvmBalanceAmount)),
	)

	initBalanceAmount0 := int64(math.Pow10(18) * 2)
	initBalance0 := sdk.NewCoins(
		sdk.NewCoin(stakeDenom, sdk.NewInt(delegationValue)),
		sdk.NewCoin(claimsDenom, sdk.NewInt(initBalanceAmount0)),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(initBalanceAmount0)),
	)

	priv0, _ := ethsecp256k1.GenerateKey()
	priv1, _ := ethsecp256k1.GenerateKey()
	addr0 := sdk.AccAddress(priv0.PubKey().Address().Bytes())
	addr1 := sdk.AccAddress(priv1.PubKey().Address().Bytes())

	var (
		claimsRecord1 types.ClaimsRecord
		params        types.Params
	)

	BeforeEach(func() {
		s.SetupTest()

		params = s.app.ClaimsKeeper.GetParams(s.ctx)
		params.EnableClaims = true
		params.AirdropStartTime = s.ctx.BlockTime()
		s.app.ClaimsKeeper.SetParams(s.ctx, params)

		coins := sdk.NewCoins(sdk.NewCoin(claimsDenom, sdk.NewInt(claimsAmount)))
		err := s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
		s.Require().NoError(err)

		// For refunding lefover gas cost - fee collector account
		// coins = sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(2000000)))
		// err = s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
		// s.Require().NoError(err)
		// err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, authtypes.FeeCollectorName, coins)
		// s.Require().NoError(err)

		balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, claimsDenom)
		Expect(balanceClaims.Amount.Uint64()).To(Equal(uint64(claimsAmount)))

		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr0, initBalance0)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, initBalance)

		claimsRecord1 = types.NewClaimsRecord(sdk.NewInt(claimValue))
		s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr1, claimsRecord1)

		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)

		balance := s.app.BankKeeper.GetBalance(s.ctx, addr0, claimsDenom)
		Expect(balance.Amount.Uint64()).To(Equal(uint64(initBalanceAmount0)))

		balance = s.app.BankKeeper.GetBalance(s.ctx, addr0, stakeDenom)
		Expect(balance.Amount.Uint64()).To(Equal(uint64(delegationValue)))

		balance = s.app.BankKeeper.GetBalance(s.ctx, addr0, evmtypes.DefaultEVMDenom)
		Expect(balance.Amount.Uint64()).To(Equal(uint64(initBalanceAmount0)))

		balance = s.app.BankKeeper.GetBalance(s.ctx, addr1, claimsDenom)
		Expect(balance.Amount.Uint64()).To(Equal(uint64(initBalanceAmount)))

		// ensure community pool doesn't have the fund
		poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, claimsDenom)
		Expect(poolBalance.Amount.Uint64()).To(Equal(uint64(0)))

		// ensure module account has the escrow fund
		moduleBalance := s.app.ClaimsKeeper.GetModuleAccountBalances(s.ctx)
		s.Require().Equal(moduleBalance.AmountOf(claimsDenom), sdk.NewInt(claimsAmount))

		s.Commit()
	})

	Context("Claiming before decay duration", func() {

		It("Claimed action ActionDelegate successfully", func() {
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, claimsDenom)

			delegate(priv1, 1)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()) - 1)))
		})

		It("Claim action ActionEVM successfully", func() {
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, claimsDenom)

			sendEthToSelf(priv1)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, claimsDenom)
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()))))
		})

		It("Claimed action ActionVote successfully", func() {
			proposalId := govProposal(priv0)
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
			govVote(priv1, proposalId)
			balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()))))

			duration := time.Until(params.AirdropEndTime()) + 10
			s.CommitAfter(duration)
			s.Commit()

			finalBalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, claimsDenom)
			Expect(finalBalance.Amount.Uint64()).To(Equal(balance.Amount.Uint64()))

			// The unclaimed amount goes to the community pool
			poolBalance := s.app.BankKeeper.GetBalance(s.ctx, distrAddr, claimsDenom)
			Expect(poolBalance.Amount.Uint64()).To(Equal(uint64(claimsAmount - actionValue)))

			// ensure module account is empty
			moduleBalance := s.app.ClaimsKeeper.GetModuleAccountBalances(s.ctx)
			Expect(moduleBalance.AmountOf(claimsDenom).Uint64()).To(Equal(uint64(0)))

			// params.EnableClaims = false
		})

		// It("Successfully claim action ActionIBCTransfer", func() {
		// 	prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())

		// 	performIbcTransfer(priv1)

		// 	balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
		// 	Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()))))
		// })
	})

	Context("Check amount claimed at 1/2 decay duration  ", func() {

	})

	Context("Check amount clawed back after decay duration  ", func() {
	})
})

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
		10,
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

	// err := msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	// s.Require().NoError(err)

	// ctx := sdk.WrapSDKContext(s.ctx)
	// rsp, err := s.app.EvmKeeper.EthereumTx(ctx, msgEthereumTx)
	// s.Require().NoError(err)
	// s.Require().Empty(rsp.VmError)

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
	fmt.Println("---res---", res.GetLog())
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
	fmt.Println("---res---", res.GetLog())
	Expect(res.IsOK()).To(Equal(true))
	return res
}

func performIbcTransfer(priv *ethsecp256k1.PrivKey) {

}
