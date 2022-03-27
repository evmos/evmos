package keeper_test

import (
	"fmt"
	"math"
	"math/big"

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
	// distrAddr := s.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	denom := claimstypes.DefaultClaimsDenom
	accountCount := 4

	// account initial balances
	initAmount := sdk.NewInt(int64(math.Pow10(18) * 2))
	initBalance := sdk.NewCoins(
		sdk.NewCoin(denom, initAmount),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, initAmount),
	)

	// account for creating the governance proposals
	initAmount0 := sdk.NewInt(int64(math.Pow10(18) * 2))
	initBalance0 := sdk.NewCoins(
		sdk.NewCoin(denom, initAmount0),
	)
	totalAmount := sdk.NewCoin(denom, initAmount.Add(initAmount0))

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

		// evmParams := s.app.EvmKeeper.GetParams(s.ctx)
		// evmParams.EvmDenom = claimtypes.DefaultClaimsDenom
		// s.app.EvmKeeper.SetParams(s.ctx, evmParams)

		// mint coins for claiming and send them to the claims module
		coins := sdk.NewCoins(totalAmount)

		err := testutil.FundModuleAccount(s.app.BankKeeper, s.ctx, inflationtypes.ModuleName, coins)
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
			acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
			s.app.AccountKeeper.SetAccount(s.ctx, acc)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
			Expect(balance.Amount).To(Equal(initAmount0))
		}

		s.Commit()
	})

	Context("ctx", func() {
		var contractAddress common.Address
		var txHash common.Hash
		BeforeAll(func() {
			// fmt.Println("----distrAddr", distrAddr)
		})

		It("deploy contract", func() {
			fmt.Println("test")
			// addr := getAddr(privs[0])
			contractAddress, txHash = deployContract(privs[0])
			fmt.Println("---contractAddress", contractAddress)
		})

		It("interact with contract", func() {
			// addr := getAddr(privs[0])
			contractInteract(privs[1], &contractAddress)
		})

		It("send registration message", func() {
			registerFeeContract(privs[1], &contractAddress, txHash)
			fee, isRegistered := s.app.FeesKeeper.GetFee(s.ctx, contractAddress)
			Expect(isRegistered).To(Equal(true))

			owner := sdk.AccAddress(privs[1].PubKey().Address().Bytes())
			Expect(fee.Contract).To(Equal(contractAddress.Hex()))
			Expect(fee.Owner).To(Equal(owner.String()))
			Expect(fee.WithdrawAddress).To(Equal(owner.String()))
			Expect(true).To(Equal(false))
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

	fmt.Println("contract deploy", res.GetLog())

	// registerEvent := res.GetEvents()[4]
	// Expect(registerEvent.Type).To(Equal(types.EventTypeRegisterFeeContract))
	// Expect(string(registerEvent.Attributes[0].Key)).To(Equal(types.AttributeKeyContract))

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress, common.Hash{}
}

func contractInteract(priv *ethsecp256k1.PrivKey, contractAddr *common.Address) {
	// amount := big.NewInt(100)

	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	// ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("mint", contractAddr, amount)
	// s.Require().NoError(err)

	// data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
	// args, err := json.Marshal(&evm.TransactionArgs{
	// 	From: &s.address,
	// 	Data: (*hexutil.Bytes)(&data),
	// })
	// s.Require().NoError(err)
	data := make([]byte, 0)
	// args := make([]byte, 0)
	gasLimit := uint64(100000)

	// ctx := sdk.WrapSDKContext(s.ctx)
	// res, err := s.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
	// 	Args:   args,
	// 	GasCap: uint64(config.DefaultGasCap),
	// })
	// fmt.Println("--err--", err)
	// s.Require().NoError(err)
	// gasLimit := res.Gas

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, contractAddr, nil, gasLimit, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	performEthTx(priv, msgEthereumTx)
	s.Commit()
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
