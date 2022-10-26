package keeper_test

import (
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	ethermint "github.com/evmos/ethermint/types"

	"github.com/evmos/evmos/v9/app"
	"github.com/evmos/evmos/v9/testutil"
	"github.com/evmos/evmos/v9/x/erc20/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
		contract  common.Address
		contract2 common.Address
	)

	BeforeEach(func() {
		s.SetupTest()

		tallyParams := s.app.GovKeeper.GetTallyParams(s.ctx)
		tallyParams.Quorum = "0.0000000001"
		s.app.GovKeeper.SetTallyParams(s.ctx, tallyParams)

	})

	Describe("Registering Token Pairs through governance", func() {
		Context("with existing coins", func() {
			BeforeEach(func() {
				// Mint coins to pay gas fee, gov deposit and registering coins in Bankkeeper
				coins := sdk.NewCoins(
					sdk.NewCoin("aevmos", sdk.NewInt(1000000000000000000)),
					sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(10000000000)),
					sdk.NewCoin(metadataIbc.Base, sdk.NewInt(1)),
					sdk.NewCoin(metadataCoin.Base, sdk.NewInt(1)),
				)
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, coins)
				s.Require().NoError(err)
				s.Commit()

				fmt.Println("**********Balances***********")
				balances := s.app.BankKeeper.GetAccountsBalances(s.ctx)
				fmt.Println(balances)
			})
			Describe("for a single Cosmos Coin", func() {
				It("should create a governance proposal", func() {
					fmt.Println()
					// register with sufficient deposit
					fmt.Println("**********Proposal***********")
					id, err := submitRegisterCoinProposal(s.ctx, s.app, privKey, []banktypes.Metadata{metadataIbc})
					s.Require().NoError(err)
					fmt.Println(s.ctx.BlockTime(), "BlockTime before Commit")
					s.Commit()

					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)
					fmt.Println(s.ctx.BlockTime(), "BlockTime after Commit")
					fmt.Println(proposal.VotingStartTime, "VotingStartTime")
					fmt.Println(proposal.VotingEndTime, "VotingEndTime")

					// delegate
					fmt.Println("**********Delegate***********")
					s.Commit()
					fmt.Println(s.ctx.BlockTime(), "BlockTime before Commit")
					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", sdk.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)
					s.Commit()
					fmt.Println(s.ctx.BlockTime(), "BlockTime after Commit")

					delegations := s.app.StakingKeeper.GetAllDelegatorDelegations(s.ctx, accAddr)
					fmt.Println(delegations)

					validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
					fmt.Println(validators, "validators")
					fmt.Println(len(validators), "len validators")

					// vote
					fmt.Println("**********Vote***********")
					s.Commit()
					fmt.Println(s.ctx.BlockTime(), "BlockTime before Commit")
					_, err = testutil.Vote(s.ctx, s.app, privKey, id)
					s.Require().NoError(err)
					s.Commit()
					fmt.Println(s.ctx.BlockTime(), "BlockTime after Commit")

					votes := s.app.GovKeeper.GetAllVotes(s.ctx)
					fmt.Println(votes, "votes")

					proposal, _ = s.app.GovKeeper.GetProposal(s.ctx, id)
					fmt.Println(proposal.Status, "Status")

					// Make proposal pass
					fmt.Println("**********Passing End time Block***********")
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + time.Hour*1
					s.CommitAndBeginBlockAfter(duration)
					fmt.Println(s.ctx.BlockTime(), "BlockTime Before Commit")

					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					s.Commit()
					fmt.Println(s.ctx.BlockTime(), "BlockTime after Commit")
					proposal, _ = s.app.GovKeeper.GetProposal(s.ctx, id)
					fmt.Println(proposal.Status, "Status")
					fmt.Println(proposal.FinalTallyResult)

					fmt.Println("**********Token Pairs***********")
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					fmt.Println(tokenPairs)
					s.Require().Equal(1, len(tokenPairs))
				})
			})
			Describe("for multiple Cosmos Coins", func() {
				It("should create a governance proposal", func() {
					id, err := submitRegisterCoinProposal(s.ctx, s.app, privKey, []banktypes.Metadata{metadataIbc, metadataCoin})
					s.Require().NoError(err)
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", sdk.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					_, err = testutil.Vote(s.ctx, s.app, privKey, id)
					s.Require().NoError(err)
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})

					s.CommitAndBeginBlockAfter(time.Hour * 1)
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
					sdk.NewCoin("aevmos", sdk.NewInt(1000000000000000000)),
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
					id, err := submitRegisterERC20Proposal(s.ctx, s.app, privKey, []string{contract.String()})
					s.Require().NoError(err)

					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", sdk.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					_, err = testutil.Vote(s.ctx, s.app, privKey, id)
					s.Require().NoError(err)
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(1, len(tokenPairs))
				})
			})
			Describe("for multiple ERC20 tokens", func() {
				It("should create a governance proposal", func() {
					// register with sufficient deposit
					id, err := submitRegisterERC20Proposal(s.ctx, s.app, privKey, []string{contract.String(), contract2.String()})
					s.Require().NoError(err)
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					// delegate
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin("aevmos", sdk.NewInt(500000000000000000)), s.validator)
					s.Require().NoError(err)
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					// vote
					s.CommitAndBeginBlockAfter(time.Hour * 1)
					_, err = testutil.Vote(s.ctx, s.app, privKey, id)
					s.Require().NoError(err)
					s.CommitAndBeginBlockAfter(time.Hour * 1)

					// Make proposal pass
					duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
					s.CommitAndBeginBlockAfter(duration)
					s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
					s.CommitAndBeginBlockAfter(time.Hour * 1)

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

				denom := s.app.ClaimsKeeper.GetParams(s.ctx).ClaimsDenom
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000))))
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
	res, err := testutil.DeliverTx(ctx, appEvmos, pk, convertCoinMsg)
	s.Require().NoError(err)
	// res := deliverTx(pk, convertCoinMsg)
	Expect(res.IsOK()).To(BeTrue(), "failed to convert coin: %s", res.Log)
}

func convertERC20(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, amt math.Int, contract common.Address) {
	addrBz := pk.PubKey().Address().Bytes()

	convertERC20Msg := types.NewMsgConvertERC20(amt, sdk.AccAddress(addrBz), contract, common.BytesToAddress(addrBz))
	res, err := testutil.DeliverTx(ctx, appEvmos, pk, convertERC20Msg)
	s.Require().NoError(err)
	Expect(res.IsOK()).To(BeTrue(), "failed to convert ERC20: %s", res.Log)
}
