package keeper_test

import (
	"math/big"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v18/testutil"
	"github.com/evmos/evmos/v18/utils"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

var _ = Describe("Performing EVM transactions", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()
		params := s.app.Erc20Keeper.GetParams(s.ctx)
		params.EnableErc20 = true
		err := s.app.Erc20Keeper.SetParams(s.ctx, params)
		Expect(err).To(BeNil())
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

	var (
		pair      types.TokenPair
		coin      sdk.Coin
		contract  common.Address
		contract2 common.Address

		// moduleAcc is the address of the ERC-20 module account
		moduleAcc sdk.AccAddress
	)

	BeforeEach(func() {
		s.SetupTest()

		moduleAcc = s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName).GetAddress()

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
					sdk.NewCoin(utils.BaseDenom, fundsAmt),
					sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, fundsAmt),
					sdk.NewCoin(metadataIbc.Base, math.NewInt(1)),
					sdk.NewCoin(metadataCoin.Base, math.NewInt(1)),
				)
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, accAddr, coins)
				Expect(err).To(BeNil())
				s.Commit()
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
					sdk.NewCoin(utils.BaseDenom, fundsAmt),
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

					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin(utils.BaseDenom, math.NewInt(500000000000000000)), s.validator)
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
					for i, tokenPair := range tokenPairs {
						if tokenPair.Erc20Address == contract.Hex() {
							s.Require().Equal(types.OWNER_EXTERNAL, tokenPairs[i].ContractOwner)
						}
					}
				})
			})

			Describe("for multiple ERC20 tokens", func() {
				BeforeEach(func() {
					// register with sufficient deposit
					id, err := submitRegisterERC20Proposal(s.ctx, s.app, privKey, []string{contract.String(), contract2.String()})
					s.Require().NoError(err)
					proposal, found := s.app.GovKeeper.GetProposal(s.ctx, id)
					s.Require().True(found)

					_, err = testutil.Delegate(s.ctx, s.app, privKey, sdk.NewCoin(utils.BaseDenom, math.NewInt(500000000000000000)), s.validator)
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
					s.Require().Equal(3, len(tokenPairs))
					for i, tokenPair := range tokenPairs {
						if tokenPair.Erc20Address == contract2.Hex() {
							s.Require().Equal(types.OWNER_EXTERNAL, tokenPairs[i].ContractOwner)
						}
					}
				})
			})
		})
	})

	Describe("Converting", func() {
		Context("with a registered ERC20", func() {
			BeforeEach(func() {
				contract := s.setupRegisterERC20Pair(contractMinterBurner)
				id := s.app.Erc20Keeper.GetTokenPairID(s.ctx, contract.String())
				pair, _ = s.app.Erc20Keeper.GetTokenPair(s.ctx, id)
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

				It("should send coins to the receiver account", func() {
					balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, accAddr, pair.Denom)
					Expect(balanceCoin).To(Equal(coin))
				})
			})
		})
	})
})

func submitRegisterERC20Proposal(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, addrs []string) (id uint64, err error) {
	content := types.NewRegisterERC20Proposal("test token", "foo", addrs...)
	return testutil.SubmitProposal(ctx, appEvmos, pk, content, 8)
}

func convertERC20(ctx sdk.Context, appEvmos *app.Evmos, pk *ethsecp256k1.PrivKey, amt math.Int, contract common.Address) {
	addrBz := pk.PubKey().Address().Bytes()
	convertERC20Msg := types.NewMsgConvertERC20(amt, sdk.AccAddress(addrBz), contract, common.BytesToAddress(addrBz))
	res, err := testutil.DeliverTx(ctx, appEvmos, pk, nil, convertERC20Msg)
	s.Require().NoError(err)
	Expect(res.IsOK()).To(BeTrue(), "failed to convert ERC20: %s", res.Log)
}
