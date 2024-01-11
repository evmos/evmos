package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	"crypto/ed25519"
	"fmt"
	tmtypes "github.com/cometbft/cometbft/types"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibcmock "github.com/cosmos/ibc-go/v7/testing/mock"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	"math/big"
	"time"
)

// ChainsIbcIntegrationTestSuite is a wrapper of ChainIntegrationTestSuite for IBC testing.
// Tendermint is Disabled for IBC testing.
type ChainsIbcIntegrationTestSuite struct {
	Chain1        *ChainIntegrationTestSuite
	Chain2        *ChainIntegrationTestSuite
	TestChain1    *ibctesting.TestChain
	TestChain2    *ibctesting.TestChain
	RelayerChain1 *itutiltypes.TestAccount
	RelayerChain2 *itutiltypes.TestAccount
	Path          *ibctesting.Path
	Coordinator   *ibctesting.Coordinator
}

// CreateChainsIbcIntegrationTestSuite initializes an IBC integration test suite from given chains.
// The input chain must disable Tendermint.
func CreateChainsIbcIntegrationTestSuite(chain1, chain2 *ChainIntegrationTestSuite, relayer1, relayer2 *itutiltypes.TestAccount) *ChainsIbcIntegrationTestSuite {
	if chain1.HasTendermint() {
		panic(fmt.Errorf("chain1 must disable Tendermint"))
	}
	if chain2.HasTendermint() {
		panic(fmt.Errorf("chain2 must disable Tendermint"))
	}

	if relayer1 == nil {
		relayer1 = NewTestAccount(chain1.t, nil)
		chain1.MintCoin(relayer1, chain1.NewBaseCoin(9))
	}
	if relayer2 == nil {
		relayer2 = NewTestAccount(chain2.t, nil)
		chain2.MintCoin(relayer2, chain2.NewBaseCoin(9))
	}

	baseFeeChain1 := chain1.ChainApp.FeeMarketKeeper().GetBaseFee(chain1.CurrentContext)
	baseFeeChain2 := chain2.ChainApp.FeeMarketKeeper().GetBaseFee(chain2.CurrentContext)

	chain1.ChainApp.FeeMarketKeeper().SetBaseFee(chain1.CurrentContext, big.NewInt(0))
	chain2.ChainApp.FeeMarketKeeper().SetBaseFee(chain2.CurrentContext, big.NewInt(0))

	chain1.Commit()
	chain2.Commit()

	coordinator := newIbcTestingCoordinator(chain1, chain2, relayer1, relayer2)

	testChain1 := coordinator.GetChain(chain1.ChainConstantsConfig.GetCosmosChainID())
	testChain2 := coordinator.GetChain(chain2.ChainConstantsConfig.GetCosmosChainID())

	suite := &ChainsIbcIntegrationTestSuite{
		Chain1:        chain1,
		Chain2:        chain2,
		TestChain1:    testChain1,
		TestChain2:    testChain2,
		RelayerChain1: relayer1,
		RelayerChain2: relayer2,
		Coordinator:   coordinator,
		Path:          newIbcTransferPath(testChain1, testChain2),
	}

	coordinator.CommitNBlocks(testChain1, 2)
	coordinator.CommitNBlocks(testChain2, 2)

	coordinator.Setup(suite.Path)

	chain1.CurrentContext = chain1.createNewContext(chain1.CurrentContext, testChain1.CurrentHeader)
	chain2.CurrentContext = chain2.createNewContext(chain2.CurrentContext, testChain2.CurrentHeader)

	// restore base fee which was set to 0 for IBC initialization purpose
	chain1.ChainApp.FeeMarketKeeper().SetBaseFee(chain1.CurrentContext, baseFeeChain1)
	chain2.ChainApp.FeeMarketKeeper().SetBaseFee(chain2.CurrentContext, baseFeeChain2)

	suite.CommitAllChains() // commit fee-market

	suite.Chain1.ibcSuite = suite
	suite.Chain2.ibcSuite = suite

	return suite
}

// newIbcTestingCoordinator creates a new IBC testing coordinator (provided by IBC-go) from given chains.
func newIbcTestingCoordinator(chain1, chain2 *ChainIntegrationTestSuite, relayer1, relayer2 *itutiltypes.TestAccount) *ibctesting.Coordinator {
	chains := make(map[string]*ibctesting.TestChain)
	coordinator := &ibctesting.Coordinator{
		T:           chain1.T(),
		CurrentTime: time.Now().UTC(),
	}

	ibcTestChain1 := newIbcTestingChain(coordinator, chain1, relayer1)
	chains[chain1.ChainConstantsConfig.GetCosmosChainID()] = ibcTestChain1

	ibcTestChain2 := newIbcTestingChain(coordinator, chain2, relayer2)
	chains[chain2.ChainConstantsConfig.GetCosmosChainID()] = ibcTestChain2

	coordinator.Chains = chains

	return coordinator
}

// newIbcTestingChain wraps an integration test chain onto an IBC testing chain (provided by IBC-go).
func newIbcTestingChain(coordinator *ibctesting.Coordinator, chain *ChainIntegrationTestSuite, relayer *itutiltypes.TestAccount) *ibctesting.TestChain {
	chainId := chain.ChainConstantsConfig.GetCosmosChainID()
	testApp := chain.ChainApp.IbcTestingApp()
	resRelayerAcc, err := chain.QueryClients.Auth.Account(chain.CurrentContext, &authtypes.QueryAccountRequest{
		Address: relayer.GetCosmosAddress().String(),
	})
	chain.Require().NoError(err)
	chain.Require().NotNil(resRelayerAcc)
	var relayerAcc authtypes.AccountI
	err2 := chain.EncodingConfig.Codec.UnpackAny(resRelayerAcc.Account, &relayerAcc)
	chain.Require().NoError(err2)

	signers := make(map[string]tmtypes.PrivValidator)
	for _, validatorAccount := range chain.ValidatorAccounts {
		//goland:noinspection GoDeprecation
		pv := ibcmock.PV{
			PrivKey: &cosmosed25519.PrivKey{
				Key: ed25519.NewKeyFromSeed(validatorAccount.PrivateKey.Key),
			},
		}
		pubKey, err := pv.GetPubKey()
		chain.Require().NoError(err)
		signers[pubKey.Address().String()] = pv
	}

	return &ibctesting.TestChain{
		T:             chain.t,
		Coordinator:   coordinator,
		ChainID:       chainId,
		App:           testApp,
		CurrentHeader: chain.CurrentContext.BlockHeader(),
		QueryServer:   chain.ChainApp.IbcKeeper(),
		TxConfig:      chain.EncodingConfig.TxConfig,
		Codec:         chain.EncodingConfig.Codec,
		Vals:          chain.ValidatorSet,
		Signers:       signers,
		SenderPrivKey: relayer.PrivateKey,
		SenderAccount: relayerAcc,
		NextVals:      chain.ValidatorSet,
	}
}

// newIbcTransferPath creates a new IBC path for transfer purpose.
func newIbcTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = ibctransfertypes.Version
	path.EndpointB.ChannelConfig.Version = ibctransfertypes.Version

	return path
}

// CommitAllChains is a MUST call method when IBC testing, it performs commit on all inner chains.
// Due to the complicated logic of the chain test suite, this function is required to sync required data for all chains.
func (suite *ChainsIbcIntegrationTestSuite) CommitAllChains() {
	chain1, testChain1, _, _ := suite.Chain(1)
	chain2, testChain2, _, _ := suite.Chain(2)

	suite.Coordinator.CurrentTime = chain1.CurrentContext.BlockHeader().Time

	headerChain1 := testChain1.CurrentHeader
	headerChain1.Time = suite.Coordinator.CurrentTime // sync
	headerChain2 := testChain2.CurrentHeader
	headerChain2.Time = suite.Coordinator.CurrentTime // sync

	chain1.CurrentContext = chain1.createNewContext(chain1.CurrentContext, headerChain1)
	chain2.CurrentContext = chain2.createNewContext(chain2.CurrentContext, headerChain2)

	chain1.ibcSuiteCommit()
	chain2.ibcSuiteCommit()

	testChain1.CurrentHeader = chain1.CurrentContext.BlockHeader()
	testChain2.CurrentHeader = chain2.CurrentContext.BlockHeader()

	suite.Coordinator.CurrentTime = chain1.CurrentContext.BlockHeader().Time // sync
}

// TemporarySetBaseFeeZero is a helper function that used to bypass EIP-1559 by x/feemarket module.
//
// It does temporarily set the base fee of all chains to 0 and returns a function that is used to restore the original base fee amount.
func (suite *ChainsIbcIntegrationTestSuite) TemporarySetBaseFeeZero() (releaser func()) {
	chain1 := suite.Chain1
	chain2 := suite.Chain2

	baseFeeChain1 := chain1.ChainApp.FeeMarketKeeper().GetBaseFee(chain1.CurrentContext)
	baseFeeChain2 := chain2.ChainApp.FeeMarketKeeper().GetBaseFee(chain2.CurrentContext)

	chain1.ChainApp.FeeMarketKeeper().SetBaseFee(chain1.CurrentContext, big.NewInt(0))
	chain2.ChainApp.FeeMarketKeeper().SetBaseFee(chain2.CurrentContext, big.NewInt(0))

	suite.CommitAllChains()

	return func() {
		chain1.CurrentContext = chain1.createNewContext(chain1.CurrentContext, suite.TestChain1.CurrentHeader)
		chain2.CurrentContext = chain2.createNewContext(chain2.CurrentContext, suite.TestChain2.CurrentHeader)

		// restore base fee which was set to 0 for IBC initialization purpose
		chain1.ChainApp.FeeMarketKeeper().SetBaseFee(chain1.CurrentContext, baseFeeChain1)
		chain2.ChainApp.FeeMarketKeeper().SetBaseFee(chain2.CurrentContext, baseFeeChain2)

		suite.CommitAllChains() // commit fee-market
	}
}

// Chain returns the chain suite, test chain, relayer and endpoint of given chain by number.
func (suite *ChainsIbcIntegrationTestSuite) Chain(number int) (
	chainSuite *ChainIntegrationTestSuite,
	testChain *ibctesting.TestChain,
	relayer *itutiltypes.TestAccount,
	endpoint *ibctesting.Endpoint,
) {
	if number == 1 {
		return suite.Chain1, suite.TestChain1, suite.RelayerChain1, suite.Path.EndpointA
	}
	if number == 2 {
		return suite.Chain2, suite.TestChain2, suite.RelayerChain2, suite.Path.EndpointB
	}
	panic(fmt.Errorf("not supported chain %d", number))
}

// Cleanup performs cleanup tasks on each of the inner chains.
func (suite *ChainsIbcIntegrationTestSuite) Cleanup() {
	if suite == nil {
		return
	}

	suite.Chain1.Cleanup()
	suite.Chain2.Cleanup()
}
