package demo

//goland:noinspection SpellCheckingInspection
import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcconntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	"github.com/evmos/evmos/v16/integration_test_util"
	"math"
)

//goland:noinspection SpellCheckingInspection

func (suite *DemoTestSuite) Test_SetupIbc() {
	suite.SetupIbcTest()
	suite.testSetupIbc()
}

func (suite *DemoTestSuite) testSetupIbc() {
	suite.SetupIbcTest()
	suite.Require().NotNil(suite.IBCITS)
	suite.Require().NotNil(suite.IBCITS.Chain1)
	suite.Require().NotNil(suite.IBCITS.Chain2)
	suite.Require().NotNil(suite.IBCITS.TestChain1)
	suite.Require().NotNil(suite.IBCITS.TestChain2)
	suite.Require().NotNil(suite.IBCITS.RelayerChain1)
	suite.Require().NotNil(suite.IBCITS.RelayerChain2)

	validateChainSetup := func(chain *integration_test_util.ChainIntegrationTestSuite) {
		suite.Require().NotNil(chain)

		ibcClientsState := chain.ChainApp.IbcKeeper().ClientKeeper.GetAllClients(chain.CurrentContext)
		suite.NotEmptyf(ibcClientsState, "%s must have clients state", chain.ChainConstantsConfig.GetCosmosChainID())

		resCons, err := chain.ChainApp.IbcKeeper().ConnectionKeeper.Connections(chain.CurrentContext, &ibcconntypes.QueryConnectionsRequest{})
		suite.Require().NoError(err)
		suite.NotEmptyf(resCons.Connections, "%s must have connections", chain.ChainConstantsConfig.GetCosmosChainID())

		channels := chain.ChainApp.IbcKeeper().ChannelKeeper.GetAllChannels(chain.CurrentContext)
		if suite.NotEmptyf(channels, "%s must have connections", chain.ChainConstantsConfig.GetCosmosChainID()) {
			suite.Equal(channels[0].PortId, "transfer")
			suite.Equal(channels[0].ChannelId, "channel-0")
		}
	}

	chain1 := suite.IBCITS.Chain1
	chain2 := suite.IBCITS.Chain2

	validateChainSetup(chain1)
	validateChainSetup(chain2)
}

func (suite *DemoTestSuite) Test_Ibc_Transfer() {
	suite.SetupIbcTest()
	suite.testSetupIbc()

	fromChain, fromTestChain, relayerSourceChain, fromEndpoint := suite.IBCITS.Chain(2)
	toChain, _, _, _ := suite.IBCITS.Chain(1)

	transferCoin := fromChain.NewBaseCoin(1)

	sender := relayerSourceChain
	receiver := toChain.WalletAccounts.Number(1)

	_ = suite.IBCITS.TxMakeIbcTransfer(fromChain, fromTestChain, fromEndpoint, toChain, sender, receiver, transferCoin)

	denomTraces, err := toChain.ChainApp.IbcTransferKeeper().DenomTraces(toChain.CurrentContext, &ibctransfertypes.QueryDenomTracesRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(denomTraces)
	if suite.Len(denomTraces.DenomTraces, 1) {
		denomTrace := denomTraces.DenomTraces[0]
		suite.Equal(transferCoin.Denom, denomTrace.BaseDenom)
	}

	denomUnit := fromChain.TestConfig.SecondaryDenomUnits[0]
	intAmt := sdkmath.NewInt(1).Mul(sdkmath.NewInt(int64(math.Pow10(int(denomUnit.Exponent)))))
	transferCoin2 := sdk.NewCoin(denomUnit.Denom, intAmt)

	fromChain.MintCoin(sender, transferCoin2)
	suite.IBCITS.CommitAllChains()

	_ = suite.IBCITS.TxMakeIbcTransferFromChain2ToChain1(receiver, transferCoin2)

	denomTraces, err = toChain.ChainApp.IbcTransferKeeper().DenomTraces(toChain.CurrentContext, &ibctransfertypes.QueryDenomTracesRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(denomTraces)
	suite.Len(denomTraces.DenomTraces, 2)
}
