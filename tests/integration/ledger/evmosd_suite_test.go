package ledger_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/evmos/evmos/v13/app"
	"github.com/evmos/evmos/v13/crypto/hd"
	"github.com/evmos/evmos/v13/tests/integration/ledger/mocks"
	utiltx "github.com/evmos/evmos/v13/testutil/tx"
	"github.com/evmos/evmos/v13/utils"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/version"

	cosmosledger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientkeys "github.com/evmos/evmos/v13/client/keys"
	evmoskeyring "github.com/evmos/evmos/v13/crypto/keyring"
	feemarkettypes "github.com/evmos/evmos/v13/x/feemarket/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	rpcclientmock "github.com/tendermint/tendermint/rpc/client/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var s *LedgerTestSuite

type LedgerTestSuite struct {
	suite.Suite

	app *app.Evmos
	ctx sdk.Context

	ledger       *mocks.SECP256K1
	accRetriever *mocks.AccountRetriever

	accAddr sdk.AccAddress

	privKey types.PrivKey
	pubKey  types.PubKey
}

func TestLedger(t *testing.T) {
	s = new(LedgerTestSuite)
	suite.Run(t, s)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Evmosd Suite")
}

func (suite *LedgerTestSuite) SetupTest() {
	var (
		err     error
		ethAddr common.Address
	)

	suite.ledger = mocks.NewSECP256K1(s.T())

	ethAddr, s.privKey = utiltx.NewAddrKey()

	s.Require().NoError(err)
	suite.pubKey = s.privKey.PubKey()

	suite.accAddr = sdk.AccAddress(ethAddr.Bytes())
}

func (suite *LedgerTestSuite) SetupEvmosApp() {
	consAddress := sdk.ConsAddress(utiltx.GenerateAddress().Bytes())

	// init app
	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
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
	s.accRetriever = mocks.NewAccountRetriever(s.T())

	initClientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		// NOTE: cmd.Execute() panics without account retriever
		WithAccountRetriever(s.accRetriever).
		WithTxConfig(encCfg.TxConfig).
		WithLedgerHasProtobuf(true).
		WithUseLedger(true).
		WithKeyring(kr).
		WithClient(mocks.MockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithChainID(utils.TestnetChainID + "-13")

	srvCtx := server.NewDefaultContext()
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &initClientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, srvCtx)

	return kr, initClientCtx, ctx
}

func (suite *LedgerTestSuite) evmosAddKeyCmd() *cobra.Command {
	cmd := keys.AddKeyCommand()

	algoFlag := cmd.Flag(flags.FlagKeyAlgorithm)
	algoFlag.DefValue = string(hd.EthSecp256k1Type)

	err := algoFlag.Value.Set(string(hd.EthSecp256k1Type))
	suite.Require().NoError(err)

	cmd.Flags().AddFlagSet(keys.Commands("home").PersistentFlags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd).WithKeyringOptions(hd.EthSecp256k1Option())
		clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}
		buf := bufio.NewReader(clientCtx.Input)
		return clientkeys.RunAddCmd(clientCtx, cmd, args, buf)
	}
	return cmd
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
