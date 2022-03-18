package keeper_test

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tharsis/evmos/v2/x/claims/types"
)

// TODO
// params := types.DefaultParams()
// params.EnableClaims = false
// s.app.ClaimsKeeper.SetParams(s.ctx, params)

var _ = Describe("Check amount claimed depending on claim time", Ordered, func() {
	params := types.DefaultParams()
	claimsAddr := s.app.AccountKeeper.GetModuleAddress(types.ModuleName)

	claimValue := int64(math.Pow10(5) * 10)
	actionValue := int64(claimValue / 4)
	claimsAmount := claimValue * 10
	initBalanceAmount := int64(math.Pow10(5) * 2)
	initEvmBalanceAmount := int64(math.Pow10(18))

	stakeDenom := stakingtypes.DefaultParams().BondDenom
	initBalance := sdk.NewCoins(
		sdk.NewCoin(stakeDenom, sdk.NewInt(initBalanceAmount)),
		sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(initBalanceAmount)),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(initEvmBalanceAmount)),
	)

	priv0, _ := ethsecp256k1.GenerateKey()
	priv1, _ := ethsecp256k1.GenerateKey()
	addr0 := sdk.AccAddress(priv0.PubKey().Address().Bytes())
	addr1 := sdk.AccAddress(priv1.PubKey().Address().Bytes())

	delegationValue := sdk.NewInt(1)
	delegationAmount := sdk.NewCoins(sdk.NewCoin(stakeDenom, delegationValue))

	var (
		claimsRecord1 types.ClaimsRecord
	)

	BeforeEach(func() {
		s.SetupTest()

		params := s.app.ClaimsKeeper.GetParams(s.ctx)
		params.EnableClaims = true
		params.AirdropStartTime = s.ctx.BlockTime()
		s.app.ClaimsKeeper.SetParams(s.ctx, params)

		coins := sdk.NewCoins(sdk.NewCoin(params.GetClaimsDenom(), sdk.NewInt(claimsAmount)))

		err := s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
		s.Require().NoError(err)

		balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, params.GetClaimsDenom())
		Expect(balanceClaims.Amount.Uint64()).To(Equal(uint64(claimsAmount)))

		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr0, initBalance)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, delegationAmount)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, initBalance)

		claimsRecord1 = types.NewClaimsRecord(sdk.NewInt(claimValue))
		s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr1, claimsRecord1)

		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr1)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)

		balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
		Expect(balance.Amount.Uint64()).To(Equal(uint64(initBalanceAmount)))

		s.Commit()
	})

	Context("Claim amount claimed before decay duration  ", func() {

		It("Successfully claim action ActionDelegate", func() {
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())

			delegate(priv1, 1)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()) - 1)))
		})

		// It("Successfully claim action ActionVote", func() {
		// 	govProposal(priv0)

		// 	prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())

		// 	govVote(priv1, 1)

		// 	balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
		// 	Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()))))
		// })

		It("Successfully claim action ActionEVM", func() {
			prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())

			performEthTx(priv1)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()))))
		})

		// It("Successfully claim action ActionIBCTransfer", func() {
		// 	prebalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())

		// 	performIbcTransfer(priv1)

		// 	balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
		// 	Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue + int64(prebalance.Amount.Uint64()))))
		// })
	})

	Context("Check amount claimed at 1/2 decay duration  ", func() {
		// cliffDuration := time.Duration(cliffLength)
		// s.CommitAfter(cliffDuration * time.Second)

	})

	Context("Check amount clawed back after decay duration  ", func() {
		// check community pool
		// balanceCommunityPool := s.app.DistrKeeper.GetFeePoolCommunityCoins(s.ctx)
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

// type CustomProposal struct {
// 	govtypes.TextProposal
// }

// func (m CustomProposal) GetDescription() string { return m.Description }
// func (m CustomProposal) GetTitle() string       { return m.Title }
// func (m CustomProposal) ProposalRoute() string  { return "gov" }
// func (m CustomProposal) ProposalType() string   { return "Text" }
// func (m CustomProposal) ValidateBasic() error   { return nil }
// func (m CustomProposal) String() string         { return m.Title + m.Description }

func govProposal(priv *ethsecp256k1.PrivKey) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	// proposal := govtypes.TextProposal{"Title", "Description"}
	// content := CustomProposal{TextProposal: proposal}
	// banktypes.Metadata
	// tx := NewRegisterCoinProposal(tc.title, tc.description, tc.metadata)

	content := incentivestypes.NewRegisterIncentiveProposal(
		"test",
		"description",
		tests.GenerateAddress().String(),
		sdk.DecCoins{sdk.NewDecCoinFromDec(types.DefaultClaimsDenom, sdk.NewDecWithPrec(0, 2))},
		10,
	)

	deposit := sdk.NewCoins(sdk.NewCoin(types.DefaultClaimsDenom, sdk.NewInt(1000)))
	msg, err := govtypes.NewMsgSubmitProposal(content, deposit, accountAddress)
	s.Require().NoError(err)

	deliverTx(priv, msg)
}

func govVote(priv *ethsecp256k1.PrivKey, proposalID uint64) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	// NewMsgDeposit
	voteMsg := govtypes.NewMsgVote(accountAddress, proposalID, 2)
	deliverTx(priv, voteMsg)
}

func performEthTx(priv *ethsecp256k1.PrivKey) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	addr := sdk.AccAddress(priv.PubKey().Address().Bytes())
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(addr.Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &from, nil, 100000, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

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

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) {
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
}

func performIbcTransfer(priv *ethsecp256k1.PrivKey) {

}
