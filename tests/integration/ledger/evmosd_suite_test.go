package ledger_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"

	cosmosledger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/evmos/ethermint/tests"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/evmos/v10/app"
	evmoskeyring "github.com/evmos/evmos/v10/crypto/keyring"
	testnetwork "github.com/evmos/evmos/v10/testutil/network"
	. "github.com/onsi/ginkgo/v2"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"
	"github.com/tendermint/tendermint/version"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	evm "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/evmos/v10/tests/integration/ledger/mocks"

	"github.com/stretchr/testify/suite"

	. "github.com/onsi/gomega"
)

var s *LedgerTestSuite

type LedgerTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app            *app.Evmos
	network        *testnetwork.Network
	queryClient    bankTypes.QueryClient
	queryClientEvm evm.QueryClient
	ledger         *mocks.SECP256K1
	ethAddr        common.Address
	accAddr        sdk.AccAddress
	signer         keyring.Signer
	privKey        *ethsecp256k1.PrivKey
	pubKey         types.PubKey

	consAddress sdk.ConsAddress
}

func TestLedger(t *testing.T) {
	s = new(LedgerTestSuite)
	suite.Run(t, s)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Evmosd Suite")
}

func (suite *LedgerTestSuite) SetupTest() {
	var err error
	suite.ledger = mocks.NewSECP256K1(s.T())
	suite.privKey, err = ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	suite.pubKey = suite.privKey.PubKey()
	suite.Require().NoError(err)
	addr, err := sdk.Bech32ifyAddressBytes("evmos", s.pubKey.Address().Bytes())
	suite.Require().NoError(err)
	suite.accAddr = sdk.MustAccAddressFromBech32(addr)
}

func (s *LedgerTestSuite) SetupEvmosApp() {

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err, "can't generate private key")

	s.ethAddr = common.BytesToAddress(priv.PubKey().Address().Bytes())
	s.accAddr = sdk.AccAddress(s.ethAddr.Bytes())
	s.signer = tests.NewSigner(priv)

	// consensus kye
	privConsKey, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err, "can't generate private key")
	consAddress := sdk.ConsAddress(privConsKey.PubKey().Address())
	s.consAddress = consAddress

	eip712.SetEncodingConfig(encoding.MakeConfig(app.ModuleBasics))
	// init app
	s.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	// query clients
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	bankTypes.RegisterQueryServer(queryHelper, s.app.BankKeeper)
	s.queryClient = bankTypes.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, s.app.EvmKeeper)
	s.queryClientEvm = evm.NewQueryClient(queryHelperEvm)
}

func (suite *LedgerTestSuite) NewKeyringAndCtxs(krHome string, input io.Reader, encCfg params.EncodingConfig) (keyring.Keyring, client.Context, context.Context) {
	kr, err := keyring.New(
		sdk.KeyringServiceName(),
		keyring.BackendTest,
		krHome,
		input,
		encCfg.Codec,
		s.MockKeyringOption(),
	)
	s.Require().NoError(err)

	initClientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		// TODO: cmd.Execute() panics without account retriever
		WithAccountRetriever(mocks.MockAccountRetriever{}).
		WithTxConfig(encCfg.TxConfig).
		WithLedgerHasProtobuf(true).
		WithUseLedger(true).
		WithKeyring(kr).
		WithClient(mocks.MockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithChainID("evmos_9000-13")

	srvCtx := server.NewDefaultContext()
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &initClientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, srvCtx)

	return kr, initClientCtx, ctx
}

func (suite *LedgerTestSuite) MockKeyringOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = evmoskeyring.SupportedAlgorithms
		options.SupportedAlgosLedger = evmoskeyring.SupportedAlgorithmsLedger
		options.LedgerDerivation = func() (cosmosledger.SECP256K1, error) { return suite.ledger, nil }
		options.LedgerCreateKey = evmoskeyring.CreatePubkey
		options.LedgerAppName = evmoskeyring.AppName
		options.LedgerSigSkipDERConv = evmoskeyring.SkipDERConversion
	}
}

func (suite *LedgerTestSuite) FormatFlag(flag string) string {
	return fmt.Sprintf("--%s", flag)
}
