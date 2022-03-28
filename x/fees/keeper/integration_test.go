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

var _ = Describe("Fees", Ordered, func() {
	// denom := types.DefaultFeesDenom
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
		priv0  *ethsecp256k1.PrivKey
		privs  []*ethsecp256k1.PrivKey
		addr0  sdk.AccAddress
		params types.Params
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

		priv0, _ = ethsecp256k1.GenerateKey()
		addr0 = getAddr(priv0)
		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr0, initBalance)

		for i := 0; i < accountCount; i++ {
			priv, _ := ethsecp256k1.GenerateKey()
			privs = append(privs, priv)
			addr := getAddr(priv)
			testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, initBalance)
			acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
			s.app.AccountKeeper.SetAccount(s.ctx, acc)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, claimsDenom)
			Expect(balance.Amount).To(Equal(initAmount))
		}

		s.Commit()
	})

	Context("ctx", func() {
		var contractAddress common.Address
		var txHash common.Hash
		BeforeAll(func() {
			contractAddress, txHash = deployContract(priv0)
		})

		It("send registration message", func() {
			registerFeeContract(priv0, &contractAddress, txHash)
			fee, isRegistered := s.app.FeesKeeper.GetFee(s.ctx, contractAddress)
			Expect(isRegistered).To(Equal(true))
			Expect(fee.Contract).To(Equal(contractAddress.Hex()))
			Expect(fee.Owner).To(Equal(addr0.String()))
			Expect(fee.WithdrawAddress).To(Equal(addr0.String()))
		})

		It("interact with contract", func() {
			preBalance := s.app.BankKeeper.GetBalance(s.ctx, addr0, evmDenom)
			cfg, _ := s.app.EvmKeeper.EVMConfig(s.ctx)
			res := contractInteract(privs[0], &contractAddress)

			registerEvent := res.GetEvents()[14]
			Expect(registerEvent.Type).To(Equal("ethereum_tx"))
			Expect(string(registerEvent.Attributes[3].Key)).To(Equal("txGasUsed"))
			gasUsed, err := strconv.ParseInt(string(registerEvent.Attributes[3].Value), 10, 64)
			s.Require().NoError(err)

			feeDistribution := new(big.Int).Mul(big.NewInt(gasUsed), cfg.BaseFee)
			receivedFee := new(big.Int).Mul(feeDistribution, big.NewInt(int64(params.DeveloperPercentage)))
			receivedFee = new(big.Int).Quo(receivedFee, big.NewInt(100))
			receivedCoins := sdk.NewCoin(evmDenom, sdk.NewIntFromBigInt(receivedFee))

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr0, evmDenom)
			Expect(balance).To(Equal(preBalance.Add(receivedCoins)))
		})
	})
})

func registerFeeContract(priv *ethsecp256k1.PrivKey, contractAddress *common.Address, deploymentHash common.Hash) {
	fromAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	msg := types.NewMsgRegisterFeeContract(fromAddress, contractAddress.String(), deploymentHash.Hex(), fromAddress)

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

func deployContract(priv *ethsecp256k1.PrivKey) (common.Address, common.Hash) {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	// ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "Test", "TTT", uint8(18))
	// s.Require().NoError(err)

	// data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
	// args, err := json.Marshal(&evm.TransactionArgs{
	// 	From: &s.address,
	// 	Data: (*hexutil.Bytes)(&data),
	// })
	// s.Require().NoError(err)
	// data := common.Hex2Bytes("600b61000e600039600b6000f30061222260005260206000f3")
	data := common.Hex2Bytes("600661000e60003960066000f300612222600055")
	// args := make([]byte, 0)
	gasLimit := uint64(100000)

	// ctx := sdk.WrapSDKContext(s.ctx)
	// res, err := s.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
	// 	Args:   args,
	// 	GasCap: uint64(config.DefaultGasCap),
	// })
	// s.Require().NoError(err)
	// gasLimit := res.Gas

	msgEthereumTx := evmtypes.NewTxContract(chainID, nonce, nil, gasLimit, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	res := performEthTx(priv, msgEthereumTx)
	s.Commit()

	ethereumTx := res.GetEvents()[10]
	Expect(ethereumTx.Type).To(Equal("ethereum_tx"))
	Expect(string(ethereumTx.Attributes[1].Key)).To(Equal("ethereumTxHash"))

	txHash := common.HexToHash(string(ethereumTx.Attributes[1].Value))
	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress, txHash
}

func contractInteract(priv *ethsecp256k1.PrivKey, contractAddr *common.Address) abci.ResponseDeliverTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTx(chainID, nonce, contractAddr, nil, gasLimit, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
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
