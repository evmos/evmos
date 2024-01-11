package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
)

// TxMakeIbcTransfer creates and submit an IBC transfer from given chain to another.
// The relayed packet will be returned.
func (suite *ChainsIbcIntegrationTestSuite) TxMakeIbcTransfer(fromChain *ChainIntegrationTestSuite, fromTestChain *ibctesting.TestChain, fromEndpoint *ibctesting.Endpoint, toChain *ChainIntegrationTestSuite, sender, receiver *itutiltypes.TestAccount, transferCoin sdk.Coin) channeltypes.Packet {
	timeoutHeight := toChain.GetIbcTimeoutHeight(100)

	msgTransfer := ibctransfertypes.NewMsgTransfer(fromEndpoint.ChannelConfig.PortID, fromEndpoint.ChannelID, transferCoin, sender.GetCosmosAddress().String(), receiver.GetCosmosAddress().String(), timeoutHeight, 0, "")

	releaser := suite.TemporarySetBaseFeeZero()

	res, err := fromTestChain.SendMsgs(msgTransfer)
	fromChain.Require().NoError(err)

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	fromChain.Require().NoError(err)

	err = suite.Path.RelayPacket(packet)
	toChain.Require().NoError(err)

	releaser()

	suite.CommitAllChains()

	return packet
}

// TxMakeIbcTransferFromChain2ToChain1 creates and submit an IBC transfer from chain2 to chain1.
// The relayed packet will be returned.
func (suite *ChainsIbcIntegrationTestSuite) TxMakeIbcTransferFromChain2ToChain1(receiver *itutiltypes.TestAccount, transferCoin sdk.Coin) channeltypes.Packet {
	fromChain, fromTestChain, relayerSourceChain, fromEndpoint := suite.Chain(2)
	toChain, _, _, _ := suite.Chain(1)

	sender := relayerSourceChain

	return suite.TxMakeIbcTransfer(fromChain, fromTestChain, fromEndpoint, toChain, sender, receiver, transferCoin)
}
