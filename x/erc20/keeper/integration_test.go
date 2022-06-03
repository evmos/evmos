package keeper_test

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/evmos/v5/app"
	"github.com/tharsis/evmos/v5/testutil"
	"github.com/tharsis/evmos/v5/x/erc20/types"

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

var _ = Describe("ERC20: Converting", Ordered, func() {
	amt := sdk.NewInt(100)
	priv, _ := ethsecp256k1.GenerateKey()
	addrBz := priv.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	addr := common.BytesToAddress(addrBz)
	moduleAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName).GetAddress()

	var (
		pair *types.TokenPair
		coin sdk.Coin
	)

	BeforeEach(func() {
		s.SetupTest()
	})

	Context("with a registered coin", func() {
		BeforeEach(func() {
			_, pair = s.setupRegisterCoin()
			coin = sdk.NewCoin(pair.Denom, amt)

			denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom
			testutil.FundAccount(s.app.BankKeeper, s.ctx, accAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000))))
			testutil.FundAccount(s.app.BankKeeper, s.ctx, accAddr, sdk.NewCoins(coin))
		})

		Describe("a Cosmos coin into an ERC20 token", func() {
			BeforeEach(func() {
				convertCoin(priv, coin)
			})

			It("should decrease coins on the sender account", func() {
				balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, accAddr, pair.Denom)
				Expect(balanceCoin.IsZero()).To(BeTrue())
			})

			It("should escrow coins on the module account", func() {
				balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, pair.Denom)
				Expect(balanceCoin).To(Equal(coin))
			})

			It("should mint tokens and send to receiver", func() {
				balanceERC20 := s.BalanceOf(pair.GetERC20Contract(), addr).(*big.Int)
				Expect(balanceERC20.Int64()).To(Equal(amt.Int64()))
			})
		})

		Describe("an ERC20 token into a Cosmos coin", func() {
			BeforeEach(func() {
				convertCoin(priv, coin)
				s.Commit()
				convertERC20(priv, amt, pair.GetERC20Contract())
			})

			It("should increase coins on the sender account", func() {
				balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, accAddr, pair.Denom)
				Expect(balanceCoin).To(Equal(coin))
			})

			It("should unescrow coins on the module account", func() {
				balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, moduleAcc, pair.Denom)
				Expect(balanceCoin.IsZero()).To(BeTrue())
			})

			It("should burn the receiver's token", func() {
				balanceERC20 := s.BalanceOf(pair.GetERC20Contract(), addr).(*big.Int)
				Expect(balanceERC20.Int64()).To(Equal(int64(0)))
			})
		})
	})

	Context("with a registered ERC20", func() {
		BeforeEach(func() {
			contract := s.setupRegisterERC20Pair(contractMinterBurner)
			id := s.app.Erc20Keeper.GetTokenPairID(s.ctx, contract.String())
			*pair, _ = s.app.Erc20Keeper.GetTokenPair(s.ctx, id)
			coin = sdk.NewCoin(pair.Denom, amt)

			denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom
			testutil.FundAccount(s.app.BankKeeper, s.ctx, accAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000))))

			_ = s.MintERC20Token(contract, s.address, addr, big.NewInt(amt.Int64()))
			s.Commit()
		})

		Describe("an ERC20 token into a Cosmos coin", func() {
			BeforeEach(func() {
				convertERC20(priv, amt, pair.GetERC20Contract())
			})

			It("should decrease tokens on the sender account", func() {
				balanceERC20 := s.BalanceOf(pair.GetERC20Contract(), addr).(*big.Int)
				Expect(balanceERC20.Int64()).To(Equal(int64(0)))
			})

			It("should escrow tokens on the module account", func() {
				moduleAddr := common.BytesToAddress(moduleAcc.Bytes())
				balanceERC20 := s.BalanceOf(pair.GetERC20Contract(), moduleAddr).(*big.Int)
				Expect(balanceERC20.Int64()).To(Equal(amt.Int64()))
			})

			It("should send coins to the recevier account", func() {
				balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, accAddr, pair.Denom)
				Expect(balanceCoin).To(Equal(coin))
			})
		})

		Describe("a Cosmos coin into an ERC20 token", func() {
			BeforeEach(func() {
				convertERC20(priv, amt, pair.GetERC20Contract())
				s.Commit()
				convertCoin(priv, coin)
			})

			It("should increase tokens on the sender account", func() {
				balanceERC20 := s.BalanceOf(pair.GetERC20Contract(), addr).(*big.Int)
				Expect(balanceERC20.Int64()).To(Equal(amt.Int64()))
			})

			It("should unescrow tokens on the module account", func() {
				moduleAddr := common.BytesToAddress(moduleAcc.Bytes())
				balanceERC20 := s.BalanceOf(pair.GetERC20Contract(), moduleAddr).(*big.Int)
				Expect(balanceERC20.Int64()).To(Equal(int64(0)))
			})

			It("should burn coins to the recevier account", func() {
				balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, accAddr, pair.Denom)
				Expect(balanceCoin.IsZero()).To(BeTrue())
			})
		})
	})
})

func convertCoin(priv *ethsecp256k1.PrivKey, coin sdk.Coin) {
	addrBz := priv.PubKey().Address().Bytes()

	convertCoinMsg := types.NewMsgConvertCoin(coin, common.BytesToAddress(addrBz), sdk.AccAddress(addrBz))
	res := deliverTx(priv, convertCoinMsg)
	s.Require().True(res.IsOK())
}

func convertERC20(priv *ethsecp256k1.PrivKey, amt sdk.Int, contract common.Address) {
	addrBz := priv.PubKey().Address().Bytes()

	convertERC20Msg := types.NewMsgConvertERC20(amt, sdk.AccAddress(addrBz), contract, common.BytesToAddress(addrBz))
	res := deliverTx(priv, convertERC20Msg)
	s.Require().True(res.IsOK())
}

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
