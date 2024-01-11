package demo

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	"math"
	"strings"
)

//goland:noinspection SpellCheckingInspection

func (suite *DemoTestSuite) Test_ERC20_DeployContract() {
	deployer := suite.CITS.WalletAccounts.Number(1)

	deployerBalanceBefore := suite.CITS.QueryBalance(0, deployer.GetCosmosAddress().String())
	suite.Require().Truef(deployerBalanceBefore.Amount.GT(sdk.ZeroInt()), "deployer must have balance")

	newContractAddress, _, resDeliver, err := suite.CITS.TxDeployErc20Contract(deployer, "coin", "token", 18)
	suite.Commit()
	suite.Require().NoError(err)
	suite.Require().NotNil(resDeliver)
	suite.Require().Equal(deployer.ComputeContractAddress(0), newContractAddress)
}

func (suite *DemoTestSuite) Test_ERC20_RegisterErc20CoinPair() {
	proposer := suite.CITS.WalletAccounts.Number(1)

	proposalId := suite.CITS.TxFullRegisterCoin(proposer, suite.CITS.TestConfig.SecondaryDenomUnits[0].Denom)
	suite.Commit()
	suite.Require().Equal(uint64(1), proposalId)

	tokenPairs, err := suite.CITS.QueryClients.Erc20.TokenPairs(suite.Ctx(), &erc20types.QueryTokenPairsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(tokenPairs)
	suite.Require().Equal(1, len(tokenPairs.TokenPairs))

	tokenPair := tokenPairs.TokenPairs[0]
	suite.assertContractCode(common.HexToAddress(tokenPair.Erc20Address))
}

func (suite *DemoTestSuite) Test_ERC20_RegisterIbcTokenPair() {
	suite.SetupIbcTest()
	suite.testSetupIbc()

	receiver := suite.CITS.WalletAccounts.Number(1)
	proposer := suite.CITS.WalletAccounts.Number(2)

	fromChain := suite.IBCITS.Chain2
	transferCoin := fromChain.NewBaseCoin(1)
	packet := suite.IBCITS.TxMakeIbcTransferFromChain2ToChain1(receiver, transferCoin)

	ibcDenom := suite.CITS.QueryDenomHash(packet.GetDestPort(), packet.GetDestChannel(), transferCoin.Denom)

	var ibcDenomDisplay string
	if len(transferCoin.Denom) <= 3 {
		ibcDenomDisplay = "eth"
	} else {
		ibcDenomDisplay = strings.ToLower(transferCoin.Denom[1:])
	}
	proposalId := suite.CITS.TxFullRegisterCoinWithNewBankMetadata(proposer, ibcDenom, ibcDenomDisplay, uint32(fromChain.ChainConstantsConfig.GetBaseExponent()))
	suite.Commit()
	suite.Require().Equal(uint64(1), proposalId)

	tokenPairs, err := suite.CITS.QueryClients.Erc20.TokenPairs(suite.Ctx(), &erc20types.QueryTokenPairsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(tokenPairs)
	suite.Require().Equal(1, len(tokenPairs.TokenPairs))

	tokenPair := tokenPairs.TokenPairs[0]
	suite.Equalf(ibcDenom, tokenPair.Denom, "token pair symbol %s must be equal to ibc denom %s", tokenPair.Denom, ibcDenom)
	suite.assertContractCode(common.HexToAddress(tokenPair.Erc20Address))
}

func (suite *DemoTestSuite) Test_ERC20_RegisterIbcTokenFromErc20() {
	suite.SetupIbcTest()

	deployer := suite.CITS.WalletAccounts.Number(1)
	proposer := suite.CITS.WalletAccounts.Number(2)

	const decimal = 6

	newContractAddress, _, resDeliver, err := suite.CITS.TxDeployErc20Contract(deployer, "coin", "token", decimal)
	suite.Require().NoError(err)
	suite.Require().NotNil(resDeliver)
	suite.CITS.Commit()

	_ = suite.CITS.TxFullRegisterIbcCoinFromErc20Contract(proposer, newContractAddress)

	tokenPair, err := suite.CITS.QueryFirstErc20TokenPair(true)
	suite.Require().NoError(err)
	suite.Require().Equal(newContractAddress.String(), tokenPair.Erc20Address)
	suite.Require().Equal(fmt.Sprintf("erc20/%s", tokenPair.Erc20Address), tokenPair.Denom)
}

func (suite *DemoTestSuite) Test_ERC20_TransferErc20IbcToken() {
	suite.SetupIbcTest()

	fromChain, _, relayer, _ := suite.IBCITS.Chain(2)
	toChain := suite.CITS
	deployer := fromChain.WalletAccounts.Number(1)
	proposer := fromChain.WalletAccounts.Number(2)
	sender := relayer
	receiver := toChain.WalletAccounts.Number(1)

	const decimal = 6
	const amount = 100

	newContractAddress, _, resDeliver, err := fromChain.TxDeployErc20Contract(deployer, "coin", "token", decimal)
	suite.Require().NoError(err)
	suite.Require().NotNil(resDeliver)
	fromChain.Commit()

	_ = fromChain.TxFullRegisterIbcCoinFromErc20Contract(proposer, newContractAddress)

	tokenPair, err := fromChain.QueryFirstErc20TokenPair(true)
	suite.Require().NoError(err)

	transferIntAmt := sdkmath.NewInt(amount).Mul(sdkmath.NewInt(int64(math.Pow10(decimal))))
	transferCoin := sdk.NewCoin(tokenPair.Denom, transferIntAmt)
	fromChain.MintCoin(sender, transferCoin)
	fromChain.Commit()

	packet := suite.IBCITS.TxMakeIbcTransferFromChain2ToChain1(receiver, transferCoin)

	ibcDenom := toChain.QueryDenomHash(packet.GetDestPort(), packet.GetDestChannel(), transferCoin.Denom)
	coinBalance := toChain.QueryBalanceByDenom(0, receiver.GetCosmosAddress().String(), ibcDenom)
	suite.Require().Equal(transferCoin.Amount.String(), coinBalance.Amount.String())
}
