package ledger_test

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/hd"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	cosmosledger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmoskeyring "github.com/evmos/evmos/v10/crypto/keyring"
	"github.com/evmos/evmos/v10/tests/integration/ledger/mocks"
	testnetwork "github.com/evmos/evmos/v10/testutil/network"
)

type LedgerE2ESuite struct {
	suite.Suite

	secp256k1 *mocks.SECP256K1
	network   *testnetwork.Network

	ethAddr common.Address
	accAddr sdk.AccAddress

	privKey *ecdsa.PrivateKey
	pubKey  *ecdsa.PublicKey
}

func TestLedger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ledger Suite")
}

func (s *LedgerE2ESuite) SetupTest() {
	s.ethAddr, s.accAddr, s.pubKey, s.privKey = s.CreateKeyPair()

	s.secp256k1 = mocks.NewSECP256K1(s.T())

	s.SetupNetwork()
}

func (s *LedgerE2ESuite) SetupNetwork() {
	var err error

	cfg := testnetwork.DefaultConfig()
	cfg.NumValidators = 1
	cfg.KeyringOptions = []keyring.Option{s.MockKeyringOption(), hd.EthSecp256k1Option()}

	s.network, err = testnetwork.New(s.T(), "build", cfg)
	s.Require().NoError(err, "can't setup test network")

	s.Require().NoError(s.network.WaitForNextBlock(), "test network can't produce blocks")
}

func (suite *LedgerE2ESuite) TearDownSuite() {
	suite.T().Log("tearing down test suite...")
	suite.network.Cleanup()
}

func (s *LedgerE2ESuite) CreateKeyPair() (common.Address, sdk.AccAddress, *ecdsa.PublicKey, *ecdsa.PrivateKey) {

	sk, err := crypto.GenerateKey()
	s.Require().NoError(err)
	pk := &sk.PublicKey

	ethAddr := crypto.PubkeyToAddress(*pk)
	accAddr := sdk.AccAddress(ethAddr.Bytes())

	return ethAddr, accAddr, pk, sk
}

func (s *LedgerE2ESuite) MockKeyringOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = evmoskeyring.SupportedAlgorithms
		options.SupportedAlgosLedger = evmoskeyring.SupportedAlgorithmsLedger
		options.LedgerDerivation = func() (cosmosledger.SECP256K1, error) { return s.secp256k1, nil }
		options.LedgerCreateKey = evmoskeyring.CreatePubkey
		options.LedgerAppName = evmoskeyring.AppName
		options.LedgerSigSkipDERConv = evmoskeyring.SkipDERConversion
	}
}

func (s *LedgerE2ESuite) FormatFlag(flag string) string {
	return fmt.Sprintf("--%s", flag)
}
