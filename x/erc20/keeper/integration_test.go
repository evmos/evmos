package keeper_test

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/evmos/v4/app"
	"github.com/tharsis/evmos/v4/testutil"
	"github.com/tharsis/evmos/v4/x/erc20/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	abci "github.com/tendermint/tendermint/abci/types"
)

var _ = Describe("Performing EVM transactions", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()

		params := s.app.Erc20Keeper.GetParams(s.ctx)
		params.EnableEVMHook = true
		params.EnableErc20 = true
		s.app.Erc20Keeper.SetParams(s.ctx, params)
	})

	// Epoch mechanism for triggering allocation and distribution
	Context("with the ERC20 module and EVM Hook disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableEVMHook = false
			params.EnableErc20 = false
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the ERC20 module disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableErc20 = false
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the EVMHook disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableEVMHook = false
			s.app.Erc20Keeper.SetParams(s.ctx, params)
		})
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})

	Context("with the ERC20 module and EVM Hook enabled", func() {
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("ERC20: Coverting", Ordered, func() {
	amt := sdk.NewInt(100)

	var (
		priv    *ethsecp256k1.PrivKey
		addr    common.Address
		accAddr sdk.AccAddress
	)

	BeforeAll(func() {
		s.SetupTest()

		priv, _ = ethsecp256k1.GenerateKey()
		addrBz := priv.PubKey().Address().Bytes()
		accAddr = sdk.AccAddress(addrBz)
		addr = common.BytesToAddress(addrBz)
	})

	Context("with a registered coin", func() {
		var pair *types.TokenPair
		var coin sdk.Coin

		BeforeAll(func() {
			_, pair = s.setupRegisterCoin()
			coin = sdk.NewCoin(pair.Denom, amt)
		})

		Describe("a Cosmos coin into ERC20 Tokens", func() {
			BeforeAll(func() {
				denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom
				testutil.FundAccount(s.app.BankKeeper, s.ctx, accAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000))))
				testutil.FundAccount(s.app.BankKeeper, s.ctx, accAddr, sdk.NewCoins(coin))

				convertCoin(priv, coin)
			})

			It("should escrow coins on the module account", func() {
				moduleAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName).GetAddress()
				coin := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, pair.Denom)
				Expect(coin).To(Equal(coin))
			})

			It("should mint and send tokens to receiver", func() {
				token := s.BalanceOf(pair.GetERC20Contract(), addr)
				Expect(token.(*big.Int).Int64()).To(Equal(amt.Int64()))
			})
		})

	})
})

func convertCoin(priv *ethsecp256k1.PrivKey, coin sdk.Coin) {
	addrBz := priv.PubKey().Address().Bytes()

	convertCoinMsg := types.NewMsgConvertCoin(coin, common.BytesToAddress(addrBz), sdk.AccAddress(addrBz))
	res := deliverTx(priv, convertCoinMsg)
	fmt.Println(res.GetLog())
	s.Require().True(res.IsOK())
}

// TODO move to testutil
func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(100000000)
	txBuilder.SetFeeAmount(sdk.Coins{{Denom: denom, Amount: sdk.NewInt(1)}})
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
	return res
}
