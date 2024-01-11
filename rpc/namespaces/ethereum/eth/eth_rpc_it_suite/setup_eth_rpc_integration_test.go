package demo

//goland:noinspection SpellCheckingInspection
import (
	"encoding/json"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v16/integration_test_util"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	"github.com/evmos/evmos/v16/rpc/namespaces/ethereum/eth"
	"github.com/stretchr/testify/suite"
	"testing"
)

//goland:noinspection GoSnakeCaseUsage,SpellCheckingInspection
type EthRpcTestSuite struct {
	suite.Suite
	CITS *integration_test_util.ChainIntegrationTestSuite
}

func (suite *EthRpcTestSuite) App() itutiltypes.ChainApp {
	return suite.CITS.ChainApp
}

func (suite *EthRpcTestSuite) Ctx() sdk.Context {
	return suite.CITS.CurrentContext
}

func (suite *EthRpcTestSuite) Commit() {
	suite.CITS.Commit()
}

func TestEthRpcTestSuite(t *testing.T) {
	suite.Run(t, new(EthRpcTestSuite))
}

func (suite *EthRpcTestSuite) SetupSuite() {
}

func (suite *EthRpcTestSuite) SetupTest() {
	suite.CITS = integration_test_util.CreateChainIntegrationTestSuite(suite.T(), suite.Require())
	suite.CITS.EnsureTendermint() // RPC requires Tendermint
}

func (suite *EthRpcTestSuite) TearDownTest() {
	suite.CITS.Cleanup()
}

func (suite *EthRpcTestSuite) TearDownSuite() {
}

func (suite *EthRpcTestSuite) GetEthPublicAPI() *eth.PublicAPI {
	return eth.NewPublicAPI(log.NewNopLogger(), suite.CITS.RpcBackendAt(0))
}

func (suite *EthRpcTestSuite) GetEthPublicAPIAt(height int64) *eth.PublicAPI {
	return eth.NewPublicAPI(log.NewNopLogger(), suite.CITS.RpcBackendAt(height))
}

func (suite *EthRpcTestSuite) GetTxReceipt(txHash common.Hash) *ethtypes.Receipt {
	mapReceipt, err := suite.CITS.RpcBackend.GetTransactionReceipt(txHash)
	suite.Require().NoError(err)
	suite.Require().NotNil(mapReceipt)

	bzMapReceipt, err := json.Marshal(mapReceipt)
	suite.Require().NoError(err)

	var receipt ethtypes.Receipt
	err = json.Unmarshal(bzMapReceipt, &receipt)
	suite.Require().NoError(err)

	return &receipt
}

func ptrInt64(num int64) *int64 {
	return &num
}
