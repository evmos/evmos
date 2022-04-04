package keeper_test

import (
	"math"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/v3/app"
	"github.com/tharsis/evmos/v3/testutil"
	"github.com/tharsis/evmos/v3/x/fees/types"
	inflationtypes "github.com/tharsis/evmos/v3/x/inflation/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	// distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	abci "github.com/tendermint/tendermint/abci/types"
	claimstypes "github.com/tharsis/evmos/v3/x/claims/types"
)

var _ = Describe("While", Ordered, func() {
	claimsDenom := claimstypes.DefaultClaimsDenom
	evmDenom := evmtypes.DefaultEVMDenom
	accountCount := 4

	// account initial balances
	initAmount := sdk.NewInt(int64(math.Pow10(18) * 2))
	initBalance := sdk.NewCoins(
		sdk.NewCoin(claimsDenom, initAmount),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, initAmount),
	)
	totalAmount := sdk.NewCoin(claimsDenom, initAmount.MulRaw(int64(accountCount)))

	var (
		deployerKey     *ethsecp256k1.PrivKey
		userKey         *ethsecp256k1.PrivKey
		deployerAddress sdk.AccAddress
		userAddress     sdk.AccAddress
		params          types.Params
		contractAddress common.Address
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.FeesKeeper.GetParams(s.ctx)
		params.EnableFees = true
		s.app.FeesKeeper.SetParams(s.ctx, params)

		// mint coins for claiming and send them to the claims module
		coins := sdk.NewCoins(totalAmount)
		err := testutil.FundModuleAccount(s.app.BankKeeper, s.ctx, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)

		// setup accounts
		deployerKey, _ = ethsecp256k1.GenerateKey()
		deployerAddress = getAddr(deployerKey)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress, initBalance)

		userKey, _ = ethsecp256k1.GenerateKey()
		userAddress = getAddr(userKey)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, userAddress, initBalance)
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, userAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Commit()

		// deploy contract and register it
		contractAddress = deployContract(deployerKey)
		registerFeeContract(deployerKey, &contractAddress, 0)
		fee, isRegistered := s.app.FeesKeeper.GetFee(s.ctx, contractAddress)
		Expect(isRegistered).To(Equal(true))
		Expect(fee.ContractAddress).To(Equal(contractAddress.Hex()))
		Expect(fee.DeployerAddress).To(Equal(deployerAddress.String()))
		Expect(fee.WithdrawAddress).To(Equal(deployerAddress.String()))
		s.Commit()
	})

	Context("fee distribution is disabled", func() {
		BeforeAll(func() {
			params = s.app.FeesKeeper.GetParams(s.ctx)
			params.EnableFees = false
			s.app.FeesKeeper.SetParams(s.ctx, params)
		})

		It("no tx fees go to developers", func() {
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)
			contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance))
		})
	})

	Context("fee distribution is enabled", func() {
		BeforeAll(func() {
			params = s.app.FeesKeeper.GetParams(s.ctx)
			params.EnableFees = true
			s.app.FeesKeeper.SetParams(s.ctx, params)
		})

		It("legacy tx fees are split validators-developers", func() {
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasPrice := big.NewInt(2000000000)
			res := contractInteract(userKey, &contractAddress, gasPrice, nil, nil, nil)

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasPrice))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
		})

		It("dynamic tx fees are split validators-developers", func() {
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			gasTipCap := big.NewInt(10000)
			gasFeeCap := new(big.Int).Add(s.app.FeeMarketKeeper.GetBaseFee(s.ctx), gasTipCap)
			res := contractInteract(userKey, &contractAddress, nil, gasFeeCap, gasTipCap, &ethtypes.AccessList{})

			gasUsed := getGasUsedFromResponse(res, 14)
			feeDistribution := sdk.NewInt(gasUsed).Mul(sdk.NewIntFromBigInt(gasFeeCap))
			developerFee := sdk.NewDecFromInt(feeDistribution).Mul(params.DeveloperShares)
			developerCoins := sdk.NewCoin(evmDenom, developerFee.TruncateInt())

			balance := s.app.BankKeeper.GetBalance(s.ctx, deployerAddress, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(developerCoins)))
		})
	})
})

func getGasUsedFromResponse(res abci.ResponseDeliverTx, index int64) int64 {
	registerEvent := res.GetEvents()[index]
	Expect(registerEvent.Type).To(Equal("ethereum_tx"))
	Expect(string(registerEvent.Attributes[3].Key)).To(Equal("txGasUsed"))
	gasUsed, err := strconv.ParseInt(string(registerEvent.Attributes[3].Value), 10, 64)
	s.Require().NoError(err)
	return gasUsed
}

func registerFeeContract(priv *ethsecp256k1.PrivKey, contractAddress *common.Address, nonce uint64) {
	fromAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	msg := types.NewMsgRegisterFeeContract(*contractAddress, fromAddress, fromAddress, []uint64{nonce})

	res := deliverTx(priv, msg)
	s.Commit()
	registerEvent := res.GetEvents()[4]
	Expect(registerEvent.Type).To(Equal(types.EventTypeRegisterFeeContract))
	Expect(string(registerEvent.Attributes[0].Key)).To(Equal(sdk.AttributeKeySender))
	Expect(string(registerEvent.Attributes[1].Key)).To(Equal(types.AttributeKeyContract))
}

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return sdk.AccAddress(priv.PubKey().Address().Bytes())
}

func deployContract(priv *ethsecp256k1.PrivKey) common.Address {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	data := common.Hex2Bytes("600661000e60003960066000f300612222600055")
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTxContract(chainID, nonce, nil, gasLimit, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	res := performEthTx(priv, msgEthereumTx)
	s.Commit()

	ethereumTx := res.GetEvents()[10]
	Expect(ethereumTx.Type).To(Equal("ethereum_tx"))
	Expect(string(ethereumTx.Attributes[1].Key)).To(Equal("ethereumTxHash"))

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func contractInteract(
	priv *ethsecp256k1.PrivKey,
	contractAddr *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) abci.ResponseDeliverTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTx(chainID, nonce, contractAddr, nil, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
	msgEthereumTx.From = from.String()

	res := performEthTx(priv, msgEthereumTx)
	s.Commit()
	return res
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
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
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	return res
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
	Expect(res.IsOK()).To(Equal(true), res.GetLog())
	return res
}
