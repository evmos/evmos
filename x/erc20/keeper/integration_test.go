package keeper_test

import (
	"math/big"
	"strconv"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	ethermint "github.com/evmos/ethermint/types"

	"github.com/evmos/evmos/v9/app"
	"github.com/evmos/evmos/v9/testutil"
	"github.com/evmos/evmos/v9/x/erc20/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

var _ = Describe("ERC20:", Ordered, func() {
	amt := sdk.NewInt(100)
	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	addr := common.BytesToAddress(addrBz)
	moduleAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName).GetAddress()

	var (
		pair      *types.TokenPair
		coin      sdk.Coin
		id        uint64
		contract  common.Address
		contract2 common.Address
	)

	BeforeEach(func() {
		s.SetupTest()
	})

	Describe("Registering Token Pairs through governance", func() {
		Context("with existing coins", func() {
			BeforeEach(func() {
				// Mint coins to pay gas fee, gov deposit and registering coins in Bankkeeper
				coins := sdk.NewCoins(
					sdk.NewCoin("aevmos", sdk.NewInt(100000000000000)),
					sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(10000000000)),
					sdk.NewCoin(metadataIbc.Base, sdk.NewInt(1)),
					sdk.NewCoin(metadataCoin.Base, sdk.NewInt(1)),
				)
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, coins)
				s.Require().NoError(err)

				tallyParams := s.app.GovKeeper.GetTallyParams(s.ctx)
				tallyParams.Quorum = "0.0000001"
				s.app.GovKeeper.SetTallyParams(s.ctx, tallyParams)

				s.Commit()
			})
			Describe("for a single Cosmos Coin", func() {
				It("should create a governance proposal", func() {
					// register with sufficient deposit
					id = submitRegisterCoinProposal(privKey, []banktypes.Metadata{metadataIbc})
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					delegate(privKey, sdk.NewCoin("aevmos", sdk.NewInt(50000000000000)))
					s.Commit()

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					vote(privKey, id)
					s.Commit()

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.Commit()

					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(1, len(tokenPairs))
				})
			})
			Describe("for multiple Cosmos Coins", func() {
				It("should create a governance proposal", func() {
					id = submitRegisterCoinProposal(privKey, []banktypes.Metadata{metadataIbc, metadataCoin})
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					delegate(privKey, sdk.NewCoin("aevmos", sdk.NewInt(50000000000000)))
					s.Commit()

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					vote(privKey, id)
					s.Commit()

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.Commit()

					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(2, len(tokenPairs))
				})
			})
		})
		Context("with deployed contracts", func() {
			BeforeEach(func() {
				// Mint coins to pay gas fee, gov deposit and registering coins in Bankkeeper
				contract, _ = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				contract2, _ = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)

				coins := sdk.NewCoins(
					sdk.NewCoin("aevmos", sdk.NewInt(100000000000000)),
					sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(10000000000)),
				)
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, coins)
				s.Require().NoError(err)

				tallyParams := s.app.GovKeeper.GetTallyParams(s.ctx)
				tallyParams.Quorum = "0.0000001"
				s.app.GovKeeper.SetTallyParams(s.ctx, tallyParams)

				s.Commit()
			})
			Describe("for a single ERC20 token", func() {
				It("should create a governance proposal", func() {
					// register with sufficient deposit
					id = submitRegisterERC20Proposal(privKey, []string{contract.String()})
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					delegate(privKey, sdk.NewCoin("aevmos", sdk.NewInt(50000000000000)))
					s.Commit()

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					vote(privKey, id)
					s.Commit()

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.Commit()

					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(1, len(tokenPairs))
				})
			})
			Describe("for a single ERC20 token", func() {
				It("should create a governance proposal", func() {
					// register with sufficient deposit
					id = submitRegisterERC20Proposal(privKey, []string{contract.String(), contract2.String()})
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					delegate(privKey, sdk.NewCoin("aevmos", sdk.NewInt(50000000000000)))
					s.Commit()

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					vote(privKey, id)
					s.Commit()

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.Commit()

					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(2, len(tokenPairs))
				})
			})
		})
	})

	Describe("Converting", func() {
		Context("with a registered coin", func() {
			BeforeEach(func() {
				pair = s.setupRegisterCoin(metadataCoin)
				coin = sdk.NewCoin(pair.Denom, amt)

				denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.TokensFromConsensusPower(100, ethermint.PowerReduction))))
				s.Require().NoError(err)
				err = testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(coin))
				s.Require().NoError(err)
			})

			Describe("a Cosmos coin into an ERC20 token", func() {
				BeforeEach(func() {
					convertCoin(privKey, coin)
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
					convertCoin(privKey, coin)
					s.Commit()
					convertERC20(privKey, amt, pair.GetERC20Contract())
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
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000))))
				s.Require().NoError(err)

				_ = s.MintERC20Token(contract, s.address, addr, big.NewInt(amt.Int64()))
				s.Commit()
			})

			Describe("an ERC20 token into a Cosmos coin", func() {
				BeforeEach(func() {
					convertERC20(privKey, amt, pair.GetERC20Contract())
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
					convertERC20(privKey, amt, pair.GetERC20Contract())
					s.Commit()
					convertCoin(privKey, coin)
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
})

func submitRegisterCoinProposal(pk *ethsecp256k1.PrivKey, metadata []banktypes.Metadata) (id uint64) {
	content := types.NewRegisterCoinProposal("test Coin", "foo", metadata...)
	return submitProposal(pk, content)
}

func submitRegisterERC20Proposal(pk *ethsecp256k1.PrivKey, addrs []string) (id uint64) {
	content := types.NewRegisterERC20Proposal("test token", "foo", addrs...)
	return submitProposal(pk, content)
}

func submitProposal(pk *ethsecp256k1.PrivKey, content govv1beta1.Content) (id uint64) {
	accountAddress := sdk.AccAddress(pk.PubKey().Address().Bytes())
	stakeDenom := stakingtypes.DefaultParams().BondDenom

	deposit := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.NewInt(100000000)))
	msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, accountAddress)
	s.Require().NoError(err)

	res := deliverTx(pk, msg)
	s.Require().Equal(uint32(0), res.Code, res.Log)

	submitEvent := res.GetEvents()[8]
	Expect(submitEvent.Type).To(Equal("submit_proposal"))
	Expect(string(submitEvent.Attributes[0].Key)).To(Equal("proposal_id"))

	proposalId, err := strconv.ParseUint(string(submitEvent.Attributes[0].Value), 10, 64)
	s.Require().NoError(err)
	return proposalId
}

func vote(priv *ethsecp256k1.PrivKey, proposalID uint64) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	voteMsg := govv1beta1.NewMsgVote(accountAddress, proposalID, 1)
	deliverTx(priv, voteMsg)
}

func delegate(priv *ethsecp256k1.PrivKey, delegateAmount sdk.Coin) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	val, err := sdk.ValAddressFromBech32(s.validator.OperatorAddress)
	s.Require().NoError(err)

	delegateMsg := stakingtypes.NewMsgDelegate(accountAddress, val, delegateAmount)
	deliverTx(priv, delegateMsg)
}

func convertCoin(priv *ethsecp256k1.PrivKey, coin sdk.Coin) {
	addrBz := priv.PubKey().Address().Bytes()

	convertCoinMsg := types.NewMsgConvertCoin(coin, common.BytesToAddress(addrBz), sdk.AccAddress(addrBz))
	res := deliverTx(priv, convertCoinMsg)
	Expect(res.IsOK()).To(BeTrue(), "failed to convert coin: %s", res.Log)
}

func convertERC20(priv *ethsecp256k1.PrivKey, amt math.Int, contract common.Address) {
	addrBz := priv.PubKey().Address().Bytes()

	convertERC20Msg := types.NewMsgConvertERC20(amt, sdk.AccAddress(addrBz), contract, common.BytesToAddress(addrBz))
	res := deliverTx(priv, convertERC20Msg)
	Expect(res.IsOK()).To(BeTrue(), "failed to convert ERC20: %s", res.Log)
}

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(100_000_000)
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

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return sdk.AccAddress(priv.PubKey().Address().Bytes())
}
