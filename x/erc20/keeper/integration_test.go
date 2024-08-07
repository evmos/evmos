package keeper_test

import (
	"fmt"
	"math/big"
<<<<<<< HEAD
	"testing"
=======
>>>>>>> main

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
<<<<<<< HEAD

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/testutil/integration/common/factory"
	testutils "github.com/evmos/evmos/v19/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v19/x/erc20/types"
=======
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/erc20/types"
>>>>>>> main
)

func TestPrecompileIntegrationTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "ERC20 Module Integration Tests")
}

var _ = Describe("Performing EVM transactions", Ordered, func() {
	var s *KeeperTestSuite
	BeforeEach(func() {
		s = new(KeeperTestSuite)
		s.SetupTest()
<<<<<<< HEAD
=======
		params := s.app.Erc20Keeper.GetParams(s.ctx)
		params.EnableErc20 = true
		err := s.app.Erc20Keeper.SetParams(s.ctx, params)
		Expect(err).To(BeNil())
>>>>>>> main
	})

	Context("with the ERC20 module disabled", func() {
		BeforeEach(func() {
			params := types.DefaultParams()
			params.EnableErc20 = false
			err := testutils.UpdateERC20Params(testutils.UpdateParamsInput{
				Tf:      s.factory,
				Network: s.network,
				Pk:      s.keyring.GetPrivKey(0),
				Params:  params,
			})
			Expect(err).To(BeNil())
		})
<<<<<<< HEAD
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
=======
>>>>>>> main
	})

	Context("with the ERC20 module and EVM Hook enabled", func() {
		It("should be successful", func() {
			_, err := s.DeployContract("coin", "token", erc20Decimals)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("ERC20:", Ordered, func() {
<<<<<<< HEAD
	var (
		s         *KeeperTestSuite
=======
	amt := math.NewInt(100)
	fundsAmt, _ := math.NewIntFromString("100000000000000000000000")

	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	addr := common.BytesToAddress(addrBz)

	var (
		pair      types.TokenPair
		coin      sdk.Coin
>>>>>>> main
		contract  common.Address
		contract2 common.Address

		// moduleAcc is the address of the ERC-20 module account
<<<<<<< HEAD
		moduleAcc = authtypes.NewModuleAddress(types.ModuleName)
		amt       = math.NewInt(100)
=======
		moduleAcc sdk.AccAddress
>>>>>>> main
	)

	BeforeEach(func() {
		s = new(KeeperTestSuite)
		s.SetupTest()
<<<<<<< HEAD
	})

	Describe("Submitting a token pair proposal through governance", func() {
=======

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

>>>>>>> main
		Context("with deployed contracts", func() {
			BeforeEach(func() {
				var err error
				contract, err = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				Expect(err).To(BeNil())
				contract2, err = s.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
				Expect(err).To(BeNil())
			})

			Describe("for a single ERC20 token", func() {
				BeforeEach(func() {
					// register erc20
					_, err := testutils.RegisterERC20(
						s.factory,
						s.network,
						testutils.ERC20RegistrationData{
							Addresses:    []string{contract.Hex()},
							ProposerPriv: s.keyring.GetPrivKey(0),
						},
					)
					Expect(err).To(BeNil())
				})

<<<<<<< HEAD
				It("should create a token pair owned by the contract deployer", func() {
					qc := s.network.GetERC20Client()

					res, err := qc.TokenPairs(s.network.GetContext(), &types.QueryTokenPairsRequest{})
					Expect(err).To(BeNil())

					tokenPairs := res.TokenPairs
					Expect(tokenPairs).To(HaveLen(2))
					for i, tokenPair := range tokenPairs {
						if tokenPair.Erc20Address == contract.Hex() {
							Expect(tokenPairs[i].ContractOwner).To(Equal(types.OWNER_EXTERNAL))
=======
				It("should create a token pairs owned by the contract deployer", func() {
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(2, len(tokenPairs))
					for i, tokenPair := range tokenPairs {
						if tokenPair.Erc20Address == contract.Hex() {
							s.Require().Equal(types.OWNER_EXTERNAL, tokenPairs[i].ContractOwner)
>>>>>>> main
						}
					}
				})
			})

			Describe("for multiple ERC20 tokens", func() {
				BeforeEach(func() {
					// register erc20 tokens
					_, err := testutils.RegisterERC20(
						s.factory,
						s.network,
						testutils.ERC20RegistrationData{
							Addresses:    []string{contract.Hex(), contract2.Hex()},
							ProposerPriv: s.keyring.GetPrivKey(0),
						},
					)
					Expect(err).To(BeNil())
				})

				It("should create a token pairs owned by the contract deployer", func() {
<<<<<<< HEAD
					qc := s.network.GetERC20Client()
					res, err := qc.TokenPairs(s.network.GetContext(), &types.QueryTokenPairsRequest{})
					Expect(err).To(BeNil())

					tokenPairs := res.TokenPairs
					Expect(tokenPairs).To(HaveLen(3))
					for i, tokenPair := range tokenPairs {
						if tokenPair.Erc20Address == contract2.Hex() {
							Expect(tokenPairs[i].ContractOwner).To(Equal(types.OWNER_EXTERNAL))
=======
					tokenPairs := s.app.Erc20Keeper.GetTokenPairs(s.ctx)
					s.Require().Equal(3, len(tokenPairs))
					for i, tokenPair := range tokenPairs {
						if tokenPair.Erc20Address == contract2.Hex() {
							s.Require().Equal(types.OWNER_EXTERNAL, tokenPairs[i].ContractOwner)
>>>>>>> main
						}
					}
				})
			})
		})
	})

	Describe("Converting", func() {
		Context("with a registered ERC20", func() {
			BeforeEach(func() {
<<<<<<< HEAD
				var err error
				contract, err = s.setupRegisterERC20Pair(contractMinterBurner)
				Expect(err).To(BeNil())
=======
				contract := s.setupRegisterERC20Pair(contractMinterBurner)
				id := s.app.Erc20Keeper.GetTokenPairID(s.ctx, contract.String())
				pair, _ = s.app.Erc20Keeper.GetTokenPair(s.ctx, id)
				coin = sdk.NewCoin(pair.Denom, amt)
>>>>>>> main

				res, err := s.MintERC20Token(contract, s.keyring.GetAddr(0), big.NewInt(amt.Int64()))
				Expect(err).To(BeNil())
				Expect(res.IsOK()).To(BeTrue())
			})

			Describe("an ERC20 token into a Cosmos coin", func() {
				BeforeEach(func() {
					// convert ERC20 to cosmos coin
					msg := types.NewMsgConvertERC20(amt, s.keyring.GetAccAddr(0), contract, s.keyring.GetAddr(0))
					res, err := s.factory.CommitCosmosTx(s.keyring.GetPrivKey(0), factory.CosmosTxArgs{Msgs: []sdk.Msg{msg}})
					Expect(err).To(BeNil())
					Expect(res.IsOK()).To(BeTrue())
				})

				It("should decrease tokens on the sender account", func() {
					balanceERC20, err := s.BalanceOf(contract, s.keyring.GetAddr(0))
					Expect(err).To(BeNil())
					Expect(balanceERC20.(*big.Int).Int64()).To(Equal(int64(0)))
				})

				It("should escrow tokens on the module account", func() {
					moduleAddr := common.BytesToAddress(moduleAcc.Bytes())
					balanceERC20, err := s.BalanceOf(contract, moduleAddr)
					Expect(err).To(BeNil())
					Expect(balanceERC20.(*big.Int).Int64()).To(Equal(amt.Int64()))
				})

				It("should send coins to the receiver account", func() {
<<<<<<< HEAD
					balRes, err := s.handler.GetBalance(s.keyring.GetAccAddr(0), fmt.Sprintf("erc20/%s", contract.Hex()))
					Expect(err).To(BeNil())
					balanceCoin := balRes.Balance
					Expect(balanceCoin.Amount).To(Equal(amt))
=======
					balanceCoin := s.app.BankKeeper.GetBalance(s.ctx, accAddr, pair.Denom)
					Expect(balanceCoin).To(Equal(coin))
>>>>>>> main
				})
			})
		})
	})
})
<<<<<<< HEAD
=======

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
>>>>>>> main
