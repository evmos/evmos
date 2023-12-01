package keeper_test

import (
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	abci "github.com/cometbft/cometbft/abci/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/utils"

	"github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/testutil"
	"github.com/evmos/evmos/v16/x/erc20/types"
)

var _ = Describe("Performing EVM transactions", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()

		params := s.app.Erc20Keeper.GetParams(s.ctx)
		params.EnableEVMHook = true
		params.EnableErc20 = true
		err := s.app.Erc20Keeper.SetParams(s.ctx, params)
		Expect(err).To(BeNil())
	})

	// Epoch mechanism for triggering allocation and distribution
	Context("with the ERC20 module and EVM Hook disabled", func() {
		BeforeEach(func() {
			params := s.app.Erc20Keeper.GetParams(s.ctx)
			params.EnableEVMHook = false
			params.EnableErc20 = false
			s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck
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
			s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck
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
			s.app.Erc20Keeper.SetParams(s.ctx, params) //nolint:errcheck
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
	amt := math.NewInt(100)
	fundsAmt, _ := math.NewIntFromString("100000000000000000000000")

	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	addr := common.BytesToAddress(addrBz)
	moduleAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName).GetAddress()

	var (
		pair      *types.TokenPair
		coin      sdk.Coin
		contract  common.Address
		contract2 common.Address
	)

	BeforeEach(func() {
		s.SetupTest()

		govParams := s.app.GovKeeper.GetParams(s.ctx)
		govParams.Quorum = "0.0000000001"
		err := s.app.GovKeeper.SetParams(s.ctx, govParams)
		Expect(err).To(BeNil())
	})

	Describe("Submitting a token pair proposal through governance", func() {
		Context("with existing coins", func() {
			BeforeEach(func() {
				// Mint coins to pay gas fee, gov deposit and registering coins in Bankkeeper
				coins := sdk.NewCoins(
					sdk.NewCoin("aevmos", fundsAmt),
					sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, fundsAmt),
					sdk.NewCoin(metadataIbc.Base, math.NewInt(1)),
					sdk.NewCoin(metadataCoin.Base, math.NewInt(1)),
				)
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, coins)
				s.Require().NoError(err)
				s.Commit()
			})
			Describe("for a single Cosmos Coin", func() {
				BeforeEach(func() {
					id, err := submitRegisterCoinProposal(s.ctx, s.app, privKey, []banktypes.Metadata{metadataIbc})
					s.Require().NoError(err)

					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", math.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)

					_, err = testutil.Vote(s.ctx, s.app, privKey, id, govv1beta1.OptionYes)
					s.Require().NoError(err)

					// Make proposal pass in EndBlocker
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + time.Hour*1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					proposal, _ = s.app.GovKeeper.GetProposal(s.ctx, id)
				})
				It("should create a token pairs owned by the erc20 module", func() {
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(1, len(tokenPairs))
					s.Require().Equal(types.OWNER_MODULE, tokenPairs[0].ContractOwner)
				})
			})
			Describe("for multiple Cosmos Coins", func() {
				BeforeEach(func() {
					id, err := submitRegisterCoinProposal(s.ctx, s.app, privKey, []banktypes.Metadata{metadataIbc, metadataCoin})
					s.Require().NoError(err)

					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", math.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)

					_, err = testutil.Vote(s.ctx, s.app, privKey, id, govv1beta1.OptionYes)
					s.Require().NoError(err)

					// Make proposal pass in EndBlocker
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					s.Commit()
				})
				It("should create a token pairs owned by the erc20 module", func() {
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(2, len(tokenPairs))
					s.Require().Equal(types.OWNER_MODULE, tokenPairs[0].ContractOwner)
				})
			})
		})
		Context("with deployed contracts", func() {
			BeforeEach(func() {
				var err error
				// Mint coins to pay gas fee, gov deposit and registering coins in Bankkeeper
				contract, err = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				s.Require().NoError(err)
				contract2, err = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				s.Require().NoError(err)

				coins := sdk.NewCoins(
					sdk.NewCoin("aevmos", fundsAmt),
					sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, fundsAmt),
				)
				err = testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, coins)
				s.Require().NoError(err)
				s.Commit()
			})
			Describe("for a single ERC20 token", func() {
				BeforeEach(func() {
					// register with sufficient deposit
					id, err := submitRegisterERC20Proposal(s.ctx, s.app, privKey, []string{contract.String()})
					s.Require().NoError(err)

					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", math.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)

					_, err = testutil.Vote(s.ctx, s.app, privKey, id, govv1beta1.OptionYes)
					s.Require().NoError(err)

					// Make proposal pass in EndBlocker
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					s.Commit()
				})
				It("should create a token pairs owned by the contract deployer", func() {
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(1, len(tokenPairs))
					s.Require().Equal(types.OWNER_EXTERNAL, tokenPairs[0].ContractOwner)
				})
			})
			Describe("for multiple ERC20 tokens", func() {
				BeforeEach(func() {
					// register with sufficient deposit
					id, err := submitRegisterERC20Proposal(s.ctx, s.app, privKey, []string{contract.String(), contract2.String()})
					s.Require().NoError(err)
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", math.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)

					_, err = testutil.Vote(s.ctx, s.app, privKey, id, govv1beta1.OptionYes)
					s.Require().NoError(err)

					// Make proposal pass in EndBlocker
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					s.Commit()
				})
				It("should create a token pairs owned by the contract deployer", func() {
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(2, len(tokenPairs))
					s.Require().Equal(types.OWNER_EXTERNAL, tokenPairs[0].ContractOwner)
				})
			})
		})
	})

	Describe("Converting", func() {
		Context("with a registered coin", func() {
			BeforeEach(func() {
				pair = s.setupRegisterCoin(metadataCoin)
				coin = sdk.NewCoin(pair.Denom, amt)

				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, fundsAmt)))
				s.Require().NoError(err)
				err = testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(coin))
				s.Require().NoError(err)
			})

			Describe("a Cosmos coin into an ERC20 token", func() {
				BeforeEach(func() {
					convertCoin(s.ctx, s.app, privKey, coin)
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
					convertCoin(s.ctx, s.app, privKey, coin)
					s.Commit()
					convertERC20(s.ctx, s.app, privKey, amt, pair.GetERC20Contract())
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

				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, fundsAmt)))
				s.Require().NoError(err)

				_ = s.MintERC20Token(contract, s.address, addr, big.NewInt(amt.Int64()))
				s.Commit()
			})

			Describe("an ERC20 token into a Cosmos coin", func() {
				BeforeEach(func() {
					convertERC20(s.ctx, s.app, privKey, amt, pair.GetERC20Contract())
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
					convertERC20(s.ctx, s.app, privKey, amt, pair.GetERC20Contract())
					s.Commit()
					convertCoin(s.ctx, s.app, privKey, coin)
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

func submitRegisterCoinProposal(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, metadata []banktypes.Metadata) (id uint64, err error) {
	content := types.NewRegisterCoinProposal("test Coin", "foo", metadata...)
	return testutil.SubmitProposal(ctx, appEvmos, pk, content, 8)
}

func submitRegisterERC20Proposal(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, addrs []string) (id uint64, err error) {
	content := types.NewRegisterERC20Proposal("test token", "foo", addrs...)
	return testutil.SubmitProposal(ctx, appEvmos, pk, content, 8)
}

func convertCoin(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, coin sdk.Coin) {
	addrBz := pk.PubKey().Address().Bytes()

	convertCoinMsg := types.NewMsgConvertCoin(coin, common.BytesToAddress(addrBz), sdk.AccAddress(addrBz))
	res, err := testutil.DeliverTx(ctx, appEvmos, pk, nil, convertCoinMsg)
	s.Require().NoError(err)

	Expect(res.IsOK()).To(BeTrue(), "failed to convert coin: %s", res.Log)
}

func convertERC20(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, amt math.Int, contract common.Address) {
	addrBz := pk.PubKey().Address().Bytes()

	convertERC20Msg := types.NewMsgConvertERC20(amt, sdk.AccAddress(addrBz), contract, common.BytesToAddress(addrBz))
	res, err := testutil.DeliverTx(ctx, appEvmos, pk, nil, convertERC20Msg)
	s.Require().NoError(err)
	Expect(res.IsOK()).To(BeTrue(), "failed to convert ERC20: %s", res.Log)
}
